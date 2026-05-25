package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// handleGetChats возвращает список уникальных диалогов пользователя
func handleGetChats(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Запрос для получения последних сообщений в каждом диалоге
	// Группируем по (other_user_id, phone_id)
	query := `
		WITH RankedMessages AS (
			SELECT 
				m.id, m.sender_id, m.receiver_id, m.phone_id, m.content, m.created_at, m.is_read,
				CASE WHEN m.sender_id = ? THEN m.receiver_id ELSE m.sender_id END as other_user_id,
				ROW_NUMBER() OVER(PARTITION BY CASE WHEN m.sender_id = ? THEN m.receiver_id ELSE m.sender_id END, m.phone_id ORDER BY m.created_at DESC) as rn
			FROM messages m
			WHERE m.sender_id = ? OR m.receiver_id = ?
		)
		SELECT 
			rm.other_user_id,
			u.email,
			rm.phone_id,
			p.title,
			rm.content,
			rm.created_at,
			(SELECT COUNT(*) FROM messages WHERE receiver_id = ? AND sender_id = rm.other_user_id AND phone_id = rm.phone_id AND is_read = 0) as unread_count
		FROM RankedMessages rm
		JOIN users u ON rm.other_user_id = u.id
		JOIN phones p ON rm.phone_id = p.id
		WHERE rm.rn = 1
		ORDER BY rm.created_at DESC
	`
	rows, err := DB.Query(query, userID, userID, userID, userID, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var chats []ChatDialog
	for rows.Next() {
		var chat ChatDialog
		if err := rows.Scan(&chat.OtherUserID, &chat.OtherUserEmail, &chat.PhoneID, &chat.PhoneTitle, &chat.LastMessage, &chat.LastMessageAt, &chat.UnreadCount); err == nil {
			chats = append(chats, chat)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(chats)
}

// handleGetMessages возвращает историю сообщений по конкретному диалогу
func handleGetMessages(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Ожидаем параметры ?other_user_id=123&phone_id=456
	otherUserID, _ := strconv.Atoi(r.URL.Query().Get("other_user_id"))
	phoneID, _ := strconv.Atoi(r.URL.Query().Get("phone_id"))

	if otherUserID == 0 || phoneID == 0 {
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	// Помечаем как прочитанные
	DB.Exec(`UPDATE messages SET is_read = 1 WHERE receiver_id = ? AND sender_id = ? AND phone_id = ? AND is_read = 0`, userID, otherUserID, phoneID)

	query := `
		SELECT id, sender_id, receiver_id, phone_id, content, is_read, created_at
		FROM messages
		WHERE phone_id = ? AND 
		      ((sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?))
		ORDER BY created_at ASC
	`
	rows, err := DB.Query(query, phoneID, userID, otherUserID, otherUserID, userID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.SenderID, &m.ReceiverID, &m.PhoneID, &m.Content, &m.IsRead, &m.CreatedAt); err == nil {
			messages = append(messages, m)
		}
	}

	if messages == nil {
		messages = []Message{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// handleSendMessage отправляет сообщение
func handleSendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := getUserIDFromSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		ReceiverID int    `json:"receiver_id"`
		PhoneID    int    `json:"phone_id"`
		Content    string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.ReceiverID == 0 || req.PhoneID == 0 || strings.TrimSpace(req.Content) == "" {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	res, err := DB.Exec(`
		INSERT INTO messages (sender_id, receiver_id, phone_id, content) 
		VALUES (?, ?, ?, ?)
	`, userID, req.ReceiverID, req.PhoneID, strings.TrimSpace(req.Content))

	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	msgID, _ := res.LastInsertId()

	// Уведомляем получателя через WebSocket
	NotifyUser(req.ReceiverID, "new_message", map[string]interface{}{
		"id":          msgID,
		"sender_id":   userID,
		"receiver_id": req.ReceiverID,
		"phone_id":    req.PhoneID,
		"content":     req.Content,
		"created_at":  time.Now(),
	})

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "Sent"}`))
}
