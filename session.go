package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"
)

const sessionCookieName = "rephone_session"

// createSessionToken generates a random base64 string
func createSessionToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// createSession создает новую сессию в БД и устанавливает cookie
func createSession(w http.ResponseWriter, userID int) error {
	token, err := createSessionToken()
	if err != nil {
		return err
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	_, err = DB.Exec(`INSERT INTO sessions (session_token, user_id, expires_at) VALUES (?, ?, ?)`, token, userID, expiresAt)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   false, // В продакшене должно быть true (HTTPS)
		Path:     "/",
	})
	return nil
}

// getUserIDFromSession извлекает ID пользователя из cookie сессии
func getUserIDFromSession(r *http.Request) (int, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return 0, false
	}

	var userID int
	var expiresAt time.Time
	err = DB.QueryRow(`SELECT user_id, expires_at FROM sessions WHERE session_token = ?`, cookie.Value).Scan(&userID, &expiresAt)
	if err != nil || time.Now().After(expiresAt) {
		return 0, false
	}

	return userID, true
}

// clearSession удаляет сессию из БД и очищает cookie
func clearSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		DB.Exec(`DELETE FROM sessions WHERE session_token = ?`, cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})
}
