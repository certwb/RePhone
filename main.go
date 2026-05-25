package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func main() {
	var err error
	// Подключение к БД SQLite
	DB, err = sql.Open("sqlite3", "./rephone.db")
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	// Проверка подключения
	if err = DB.Ping(); err != nil {
		log.Fatal(err)
	}
	log.Println("Успешное подключение к базе данных.")

	// Инициализация базы данных (создание таблиц)
	initDB(DB)

	// Инициализация WebSocket Hub
	globalHub = newHub()
	go globalHub.run()

	// Создаем маршрутизатор
	mux := http.NewServeMux()

	// Раздача статических файлов из папки static/
	// Убедитесь, что папка создана
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// WebSocket маршрут
	mux.HandleFunc("/api/ws", handleWebSocket)

	// Маршруты аутентификации
	mux.HandleFunc("/api/auth/register", handleRegister)
	mux.HandleFunc("/api/auth/login", handleLogin)
	mux.HandleFunc("/api/auth/logout", handleLogoutAll)
	mux.HandleFunc("/api/auth/change-password", requireAuth(handleChangePassword))

	// Маршруты OAuth (заглушки)
	mux.HandleFunc("/api/auth/google/login", handleOAuthGoogleLogin)
	mux.HandleFunc("/api/auth/google/callback", handleOAuthGoogleCallback)

	// Пользовательские маршруты
	mux.HandleFunc("/api/user/profile", requireAuth(handleUserProfile))
	mux.HandleFunc("/api/user/phones", requireAuth(handleUserPhones))
	mux.HandleFunc("/api/user/favorites", requireAuth(handleUserFavorites))
	mux.HandleFunc("/api/user/info", handleUserInfo) // Public user info

	// Маршруты объявлений (Phones)
	mux.HandleFunc("/api/phones", handlePhones)
	mux.HandleFunc("/api/phones/", handlePhoneByID)
	mux.HandleFunc("/promo/", handleServePromo)

	// Избранное
	mux.HandleFunc("/api/favorites/toggle", requireAuth(handleToggleFavorite))

	// Справочники
	mux.HandleFunc("/api/cities", handleGetCities)

	// Сообщения
	mux.HandleFunc("/api/user/chats", requireAuth(handleGetChats))
	mux.HandleFunc("/api/user/messages", requireAuth(handleGetMessages))
	mux.HandleFunc("/api/user/messages/send", requireAuth(handleSendMessage))

	// Отзывы
	mux.HandleFunc("/api/reviews", handleReviews)
	// Базовый обработчик главной страницы
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "static/index.html")
	})

	log.Println("Сервер запущен на http://localhost:8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}

// initDB читает schema.sql и выполняет скрипты для создания таблиц
func initDB(db *sql.DB) {
	// Простые миграции (игнорируем ошибки если колонки уже есть)
	DB.Exec(`ALTER TABLE phones ADD COLUMN contact_phone TEXT DEFAULT ''`)
	DB.Exec(`ALTER TABLE phones ADD COLUMN views_count INTEGER DEFAULT 0`)
	DB.Exec(`ALTER TABLE users ADD COLUMN name TEXT DEFAULT 'Пользователь'`)
	DB.Exec(`ALTER TABLE users ADD COLUMN avatar_url TEXT DEFAULT ''`)
	DB.Exec(`ALTER TABLE users ADD COLUMN city_id INTEGER`)
	DB.Exec(`ALTER TABLE phones ADD COLUMN city_id INTEGER`)
	DB.Exec(`ALTER TABLE phones ADD COLUMN bumped_at DATETIME`)
	DB.Exec(`UPDATE phones SET bumped_at = created_at WHERE bumped_at IS NULL`)

	// Таблица promo_pages
	DB.Exec(`
		CREATE TABLE IF NOT EXISTS promo_pages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			phone_id INTEGER NOT NULL UNIQUE,
			slug TEXT NOT NULL UNIQUE,
			html_content TEXT NOT NULL,
			views_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(phone_id) REFERENCES phones(id) ON DELETE CASCADE
		)
	`)

	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		log.Printf("Внимание: Не удалось прочитать schema.sql: %v", err)
		return
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		log.Fatalf("Ошибка при выполнении schema.sql: %v", err)
	}
	log.Println("Схема базы данных успешно применена.")
}
