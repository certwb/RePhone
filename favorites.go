package main

import (
	"encoding/json"
	"net/http"
)

type ToggleFavoriteRequest struct {
	PhoneID int `json:"phone_id"`
}

func handleToggleFavorite(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := getUserIDFromSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req ToggleFavoriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Проверяем, есть ли уже в избранном
	var exists bool
	err := DB.QueryRow(`SELECT EXISTS(SELECT 1 FROM favorites WHERE user_id = ? AND phone_id = ?)`, userID, req.PhoneID).Scan(&exists)
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	if exists {
		// Удаляем
		_, err = DB.Exec(`DELETE FROM favorites WHERE user_id = ? AND phone_id = ?`, userID, req.PhoneID)
		if err != nil {
			http.Error(w, "Error removing favorite", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "removed"}`))
	} else {
		// Добавляем
		_, err = DB.Exec(`INSERT INTO favorites (user_id, phone_id) VALUES (?, ?)`, userID, req.PhoneID)
		if err != nil {
			http.Error(w, "Error adding favorite", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "added"}`))
	}
}
