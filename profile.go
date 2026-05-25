package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func handleUserProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		var u User
		err := DB.QueryRow(`
			SELECT u.id, u.email, u.name, u.avatar_url, u.city_id, u.created_at,
			       COALESCE(AVG(r.rating), 0) as average_rating,
			       COUNT(r.id) as review_count
			FROM users u
			LEFT JOIN reviews r ON u.id = r.seller_id
			WHERE u.id = ?
			GROUP BY u.id
		`, userID).
			Scan(&u.ID, &u.Email, &u.Name, &u.AvatarURL, &u.CityID, &u.CreatedAt, &u.AverageRating, &u.ReviewCount)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(u)
		return
	}

	if r.Method == http.MethodPut {
		r.ParseMultipartForm(5 << 20) // 5 MB
		name := strings.TrimSpace(r.FormValue("name"))
		cityID := r.FormValue("city_id")

		if name != "" {
			DB.Exec(`UPDATE users SET name = ? WHERE id = ?`, name, userID)
		}
		if cityID != "" {
			DB.Exec(`UPDATE users SET city_id = ? WHERE id = ?`, cityID, userID)
		}

		file, handler, err := r.FormFile("avatar")
		if err == nil {
			defer file.Close()
			if handler.Size <= 2<<20 && strings.HasPrefix(handler.Header.Get("Content-Type"), "image/") {
				ext := filepath.Ext(handler.Filename)
				filename := fmt.Sprintf("avatar_%d_%d%s", userID, time.Now().UnixNano(), ext)
				path := filepath.Join("static", "uploads", filename)

				dst, err := os.Create(path)
				if err == nil {
					io.Copy(dst, file)
					dst.Close()
					avatarURL := "/static/uploads/" + filename
					DB.Exec(`UPDATE users SET avatar_url = ? WHERE id = ?`, avatarURL, userID)
				}
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Profile updated successfully"}`))
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := getUserIDFromSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var currentHash string
	err := DB.QueryRow(`SELECT password_hash FROM users WHERE id = ?`, userID).Scan(&currentHash)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(req.OldPassword)); err != nil {
		http.Error(w, "Invalid old password", http.StatusUnauthorized)
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	DB.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, string(newHash), userID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Password updated successfully"}`))
}

func handleLogoutAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := getUserIDFromSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	DB.Exec(`DELETE FROM sessions WHERE user_id = ?`, userID)

	// Clear current cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Logged out from all devices"}`))
}
