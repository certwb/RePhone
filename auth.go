package main

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

// Request models
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// handleRegister обрабатывает регистрацию пользователя
func handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Хэширование пароля
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Сохранение в БД
	res, err := DB.Exec(`INSERT INTO users (email, password_hash) VALUES (?, ?)`, req.Email, string(hash))
	if err != nil {
		http.Error(w, "Email already exists or DB error", http.StatusConflict)
		return
	}

	userID, _ := res.LastInsertId()

	// Создаем сессию
	if err := createSession(w, int(userID)); err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "User registered successfully"}`))
}

// handleLogin обрабатывает вход пользователя
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Поиск пользователя в БД
	var user User
	err := DB.QueryRow(`SELECT id, password_hash FROM users WHERE email = ?`, req.Email).Scan(&user.ID, &user.PasswordHash)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	// Создаем сессию
	if err := createSession(w, user.ID); err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Logged in successfully"}`))
}

// handleLogout обрабатывает выход пользователя
func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	clearSession(w, r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Logged out successfully"}`))
}

// requireAuth это простая middleware для проверки сессии
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := getUserIDFromSession(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// Заглушки для OAuth (Google / GitHub)
func handleOAuthGoogleLogin(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Redirect to Google OAuth provider"))
}

func handleOAuthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Process Google OAuth callback, create user and session"))
}
