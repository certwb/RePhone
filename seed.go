//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	db, err := sql.Open("sqlite3", "./rephone.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Начинаем сидирование базы данных...")

	// 1. Создаем тестового пользователя
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	_, err = db.Exec(`INSERT OR IGNORE INTO users (email, password_hash) VALUES (?, ?)`, "test@rephone.kz", string(hash))
	if err != nil {
		log.Fatal("Ошибка создания пользователя: ", err)
	}
	
	// Получаем ID пользователя (может быть 1, если база чистая)
	var userID int
	err = db.QueryRow(`SELECT id FROM users WHERE email = ?`, "test@rephone.kz").Scan(&userID)
	if err != nil {
		log.Fatal("Не удалось получить ID пользователя: ", err)
	}

	// 2. Список телефонов для вставки
	phones := []struct {
		Title         string
		BrandID       int
		Model         string
		Storage       int
		BatteryHealth int
		Price         int
		Condition     string
		Description   string
		ImageURL      string
	}{
		{"Apple iPhone 15 Pro Titanium", 1, "iPhone 15 Pro", 256, 100, 550000, "Новый", "В заводской пленке. Не активирован.", "https://images.unsplash.com/photo-1695048133142-1a20484d2569?auto=format&fit=crop&q=80&w=600"},
		{"iPhone 14 Pro Max", 1, "iPhone 14 Pro Max", 512, 92, 450000, "Идеальное", "Носился всегда в чехле и защитном стекле. Полный комплект.", "https://images.unsplash.com/photo-1678652197831-2d180705cd2c?auto=format&fit=crop&q=80&w=600"},
		{"Apple iPhone 13", 1, "iPhone 13", 128, 86, 280000, "Хорошее", "Есть небольшие царапины на корпусе, экран в идеале.", "https://images.unsplash.com/photo-1632661674596-df8be070a5c5?auto=format&fit=crop&q=80&w=600"},
		{"Samsung Galaxy S24 Ultra", 2, "Galaxy S24 Ultra", 512, 100, 600000, "Новый", "Подарили, но привык к iOS. Коробка не вскрыта.", "https://images.unsplash.com/photo-1707323869152-44160cb673da?auto=format&fit=crop&q=80&w=600"},
		{"Samsung Galaxy S23", 2, "Galaxy S23", 256, 95, 320000, "Идеальное", "Пользовалась девушка, очень аккуратно. Гарантия еще 2 месяца.", "https://images.unsplash.com/photo-1675789652575-5db4f36c5f78?auto=format&fit=crop&q=80&w=600"},
		{"Samsung Galaxy A54", 2, "Galaxy A54", 128, 90, 140000, "Хорошее", "Отличный бюджетник с крутой камерой. Без торга.", "https://images.unsplash.com/photo-1610945265064-0e34e5519bbf?auto=format&fit=crop&q=80&w=600"},
		{"Xiaomi 14 Pro", 3, "14 Pro", 256, 100, 410000, "Новый", "Глобальная версия, камеры Leica. Мощный аппарат.", "https://images.unsplash.com/photo-1598327105666-5b89351aff97?auto=format&fit=crop&q=80&w=600"},
		{"Xiaomi Redmi Note 13 Pro", 3, "Redmi Note 13 Pro", 256, 98, 160000, "Идеальное", "Использовался как второй телефон для работы.", "https://images.unsplash.com/photo-1662947116812-706d860e3650?auto=format&fit=crop&q=80&w=600"},
		{"Google Pixel 8 Pro", 4, "Pixel 8 Pro", 128, 99, 440000, "Идеальное", "Лучшая камера на рынке. Чистый Android. В комплекте чехол Bellroy.", "https://images.unsplash.com/photo-1696446700622-4467c679a613?auto=format&fit=crop&q=80&w=600"},
		{"Google Pixel 7a", 4, "Pixel 7a", 128, 94, 210000, "Хорошее", "Компактный и быстрый. На стекле есть микроцарапина.", "https://images.unsplash.com/photo-1598327105666-5b89351aff97?auto=format&fit=crop&q=80&w=600"},
		{"OnePlus 12", 5, "12", 256, 100, 390000, "Новый", "Запечатанный. Цвет черный.", "https://images.unsplash.com/photo-1511707171634-5f897ff02aa9?auto=format&fit=crop&q=80&w=600"},
		{"Sony Xperia 1 V", 7, "Xperia 1 V", 256, 96, 480000, "Идеальное", "Для ценителей качественного звука и экранов 4K.", "https://images.unsplash.com/photo-1542490597-d86b245e1286?auto=format&fit=crop&q=80&w=600"},
	}

	for _, p := range phones {
		// Вставляем телефон
		res, err := db.Exec(`
			INSERT INTO phones (user_id, title, brand_id, model, storage, battery_health, price, condition, description)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, userID, p.Title, p.BrandID, p.Model, p.Storage, p.BatteryHealth, p.Price, p.Condition, p.Description)
		
		if err != nil {
			fmt.Printf("Ошибка при вставке '%s': %v\n", p.Title, err)
			continue
		}

		phoneID, _ := res.LastInsertId()

		// Вставляем картинку (как primary)
		_, err = db.Exec(`
			INSERT INTO images (phone_id, image_url, is_primary)
			VALUES (?, ?, 1)
		`, phoneID, p.ImageURL)
		
		if err != nil {
			fmt.Printf("Ошибка при вставке изображения для '%s': %v\n", p.Title, err)
		}
	}

	fmt.Printf("Успешно добавлено %d тестовых объявлений!\n", len(phones))
	fmt.Println("Тестовый пользователь: test@rephone.kz / password123")
}
