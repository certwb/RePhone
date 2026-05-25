package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// handlePhones обрабатывает маршрут /api/phones
func handlePhones(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getPhones(w, r)
	case http.MethodPost:
		// Для создания нужна авторизация, поэтому проверяем сессию вручную
		// (или можно обернуть этот метод в requireAuth в main.go)
		userID, ok := getUserIDFromSession(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		createPhone(w, r, userID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePhoneByID обрабатывает маршрут /api/phones/{id}
func handlePhoneByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/phones/")
	parts := strings.Split(path, "/")
	phoneID, err := strconv.Atoi(parts[0])
	if err != nil || phoneID <= 0 {
		http.Error(w, "Invalid phone ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		// Добавляем обработку /api/phones/{id}/phone
		if len(parts) > 1 && parts[1] == "phone" {
			handlePhoneContact(w, phoneID)
			return
		}
		getPhoneByID(w, phoneID)
	case http.MethodPost:
		if len(parts) > 1 {
			if parts[1] == "bump" {
				handlePhoneBump(w, r, phoneID)
				return
			}
			if parts[1] == "promo" {
				handleGeneratePromo(w, r, phoneID)
				return
			}
		}
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case http.MethodPut:
		// Обновление
		userID, ok := getUserIDFromSession(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		updatePhone(w, r, phoneID, userID)
	case http.MethodDelete:
		// Удаление
		userID, ok := getUserIDFromSession(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		deletePhone(w, phoneID, userID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getPhones(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromSession(r)

	// Базовая фильтрация
	brandID := r.URL.Query().Get("brand_id")
	minPrice := r.URL.Query().Get("min_price")
	maxPrice := r.URL.Query().Get("max_price")
	searchQuery := r.URL.Query().Get("q")
	condition := r.URL.Query().Get("condition")
	cityID := r.URL.Query().Get("city_id")
	limit := r.URL.Query().Get("limit")
	lastID := r.URL.Query().Get("last_id")

	if limit == "" {
		limit = "10"
	}

	query := `
		SELECT p.id, p.user_id, p.title, p.brand_id, b.name, p.city_id, c.name, p.model, p.storage, 
		       p.battery_health, p.price, p.condition, p.description, p.contact_phone, p.views_count, p.status, p.created_at, p.bumped_at,
		       CASE WHEN f.phone_id IS NOT NULL THEN 1 ELSE 0 END as is_favorited
		FROM phones p
		JOIN brands b ON p.brand_id = b.id
		LEFT JOIN cities c ON p.city_id = c.id
		LEFT JOIN favorites f ON p.id = f.phone_id AND f.user_id = ?
		WHERE p.status = 'active'
	`
	args := []interface{}{userID}

	if brandID != "" {
		query += " AND p.brand_id = ?"
		args = append(args, brandID)
	}
	if cityID != "" {
		query += " AND p.city_id = ?"
		args = append(args, cityID)
	}
	if minPrice != "" {
		query += " AND p.price >= ?"
		args = append(args, minPrice)
	}
	if maxPrice != "" {
		query += " AND p.price <= ?"
		args = append(args, maxPrice)
	}
	if condition != "" {
		query += " AND p.condition = ?"
		args = append(args, condition)
	}
	if searchQuery != "" {
		query += " AND (p.title LIKE ? OR p.model LIKE ?)"
		likeTerm := "%" + searchQuery + "%"
		args = append(args, likeTerm, likeTerm)
	}

	if lastID != "" && lastID != "0" {
		query += " AND p.id < ?"
		args = append(args, lastID)
	}

	query += " ORDER BY p.bumped_at DESC, p.id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := DB.Query(query, args...)
	if err != nil {
		http.Error(w, "Error fetching phones", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	phones := []Phone{}
	for rows.Next() {
		var p Phone
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.BrandID, &p.BrandName, &p.CityID, &p.CityName, &p.Model, &p.Storage,
			&p.BatteryHealth, &p.Price, &p.Condition, &p.Description, &p.ContactPhone, &p.ViewsCount, &p.Status, &p.CreatedAt, &p.BumpedAt, &p.IsFavorited); err != nil {
			continue
		}
		if len(p.ContactPhone) > 6 {
			p.ContactPhone = p.ContactPhone[:6] + " *** ** " + p.ContactPhone[len(p.ContactPhone)-2:]
		}
		phones = append(phones, p)
	}

	// Для MVP: можно загрузить главное изображение для каждого телефона
	// В реальном проекте лучше использовать LEFT JOIN с GROUP BY или подзапрос
	for i, p := range phones {
		var imgURL string
		err := DB.QueryRow(`SELECT image_url FROM images WHERE phone_id = ? AND is_primary = 1 LIMIT 1`, p.ID).Scan(&imgURL)
		if err == nil {
			phones[i].Images = []Image{{ImageURL: imgURL, IsPrimary: true}}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(phones)
}

func createPhone(w http.ResponseWriter, r *http.Request, userID int) {
	// Ограничение размера тела (10 MB)
	r.ParseMultipartForm(10 << 20)

	title := r.FormValue("title")
	brandID, _ := strconv.Atoi(r.FormValue("brand_id"))
	model := r.FormValue("model")
	storage, _ := strconv.Atoi(r.FormValue("storage"))
	batteryHealth, _ := strconv.Atoi(r.FormValue("battery_health"))
	price, _ := strconv.Atoi(r.FormValue("price"))
	condition := r.FormValue("condition")
	description := r.FormValue("description")
	contactPhone := r.FormValue("contact_phone")
	cityID, _ := strconv.Atoi(r.FormValue("city_id"))

	if title == "" || price <= 0 || price > 20000000 || brandID == 0 {
		http.Error(w, "Invalid fields (check title, brand, and price)", http.StatusBadRequest)
		return
	}

	// Валидация файлов до сохранения в БД
	files := r.MultipartForm.File["images"]
	for _, fileHeader := range files {
		if fileHeader.Size > 2<<20 {
			http.Error(w, "Каждое изображение должно быть меньше 2 МБ", http.StatusBadRequest)
			return
		}
		if !strings.HasPrefix(fileHeader.Header.Get("Content-Type"), "image/") {
			http.Error(w, "Разрешены только изображения", http.StatusBadRequest)
			return
		}
	}

	// Начинаем транзакцию
	tx, err := DB.Begin()
	if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	res, err := tx.Exec(`
		INSERT INTO phones (user_id, title, brand_id, city_id, model, storage, battery_health, price, condition, description, contact_phone, bumped_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, userID, title, brandID, cityID, model, storage, batteryHealth, price, condition, description, contactPhone)

	if err != nil {
		tx.Rollback()
		http.Error(w, "Error saving phone", http.StatusInternalServerError)
		return
	}

	phoneID, _ := res.LastInsertId()

	// Обработка загруженных файлов
	for i, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			continue
		}
		defer file.Close()

		// Генерация уникального имени
		ext := filepath.Ext(fileHeader.Filename)
		filename := fmt.Sprintf("%d_%d%s", phoneID, time.Now().UnixNano(), ext)
		path := filepath.Join("static", "uploads", filename)

		dst, err := os.Create(path)
		if err != nil {
			continue
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			continue
		}

		isPrimary := 0
		if i == 0 {
			isPrimary = 1
		}
		imgURL := "/static/uploads/" + filename
		tx.Exec(`INSERT INTO images (phone_id, image_url, is_primary) VALUES (?, ?, ?)`, phoneID, imgURL, isPrimary)
	}

	tx.Commit()

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf(`{"message": "Phone created successfully", "id": %d}`, phoneID)))
}

func getPhoneByID(w http.ResponseWriter, phoneID int) {
	var p Phone
	err := DB.QueryRow(`
		SELECT p.id, p.user_id, p.title, p.brand_id, b.name, p.model, p.storage, 
		       p.battery_health, p.price, p.condition, p.description, p.contact_phone, p.views_count, p.status, p.created_at, p.bumped_at 
		FROM phones p
		JOIN brands b ON p.brand_id = b.id
		WHERE p.id = ? AND p.status != 'deleted'
	`, phoneID).Scan(&p.ID, &p.UserID, &p.Title, &p.BrandID, &p.BrandName, &p.Model, &p.Storage,
		&p.BatteryHealth, &p.Price, &p.Condition, &p.Description, &p.ContactPhone, &p.ViewsCount, &p.Status, &p.CreatedAt, &p.BumpedAt)

	if err != nil {
		http.Error(w, "Phone not found", http.StatusNotFound)
		return
	}
	
	if len(p.ContactPhone) > 6 {
		p.ContactPhone = p.ContactPhone[:6] + " *** ** " + p.ContactPhone[len(p.ContactPhone)-2:]
	}

	// Загружаем изображения
	rows, err := DB.Query(`SELECT id, image_url, is_primary FROM images WHERE phone_id = ?`, p.ID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var img Image
			if err := rows.Scan(&img.ID, &img.ImageURL, &img.IsPrimary); err == nil {
				p.Images = append(p.Images, img)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func updatePhone(w http.ResponseWriter, r *http.Request, phoneID int, userID int) {
	// MVP: обновление только статуса или цены
	var req struct {
		Price  *int    `json:"price"`
		Status *string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Проверяем владельца
	var ownerID int
	if err := DB.QueryRow(`SELECT user_id FROM phones WHERE id = ?`, phoneID).Scan(&ownerID); err != nil || ownerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if req.Price != nil {
		DB.Exec(`UPDATE phones SET price = ? WHERE id = ?`, *req.Price, phoneID)
	}
	if req.Status != nil {
		DB.Exec(`UPDATE phones SET status = ? WHERE id = ?`, *req.Status, phoneID)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Phone updated successfully"}`))
}

func deletePhone(w http.ResponseWriter, phoneID int, userID int) {
	// Soft delete
	res, err := DB.Exec(`UPDATE phones SET status = 'deleted' WHERE id = ? AND user_id = ?`, phoneID, userID)
	if err != nil {
		http.Error(w, "Error deleting phone", http.StatusInternalServerError)
		return
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		http.Error(w, "Forbidden or not found", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Phone deleted successfully"}`))
}

func handleUserPhones(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT p.id, p.user_id, p.title, p.brand_id, b.name, p.model, p.storage, 
		       p.battery_health, p.price, p.condition, p.description, p.contact_phone, p.views_count, p.status, p.created_at, p.bumped_at,
			   CASE WHEN f.phone_id IS NOT NULL THEN 1 ELSE 0 END as is_favorited,
			   IFNULL(pr.slug, '') as promo_slug
		FROM phones p
		JOIN brands b ON p.brand_id = b.id
		LEFT JOIN favorites f ON p.id = f.phone_id AND f.user_id = ?
		LEFT JOIN promo_pages pr ON p.id = pr.phone_id
		WHERE p.user_id = ? AND p.status != 'deleted'
		ORDER BY p.created_at DESC
	`
	rows, err := DB.Query(query, userID, userID)
	if err != nil {
		http.Error(w, "Error fetching user phones", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	phones := []Phone{}
	for rows.Next() {
		var p Phone
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.BrandID, &p.BrandName, &p.Model, &p.Storage,
			&p.BatteryHealth, &p.Price, &p.Condition, &p.Description, &p.ContactPhone, &p.ViewsCount, &p.Status, &p.CreatedAt, &p.BumpedAt, &p.IsFavorited, &p.PromoSlug); err != nil {
			continue
		}
		if len(p.ContactPhone) > 6 {
			p.ContactPhone = p.ContactPhone[:6] + " *** ** " + p.ContactPhone[len(p.ContactPhone)-2:]
		}
		phones = append(phones, p)
	}

	for i, p := range phones {
		var imgURL string
		err := DB.QueryRow(`SELECT image_url FROM images WHERE phone_id = ? AND is_primary = 1 LIMIT 1`, p.ID).Scan(&imgURL)
		if err == nil {
			phones[i].Images = []Image{{ImageURL: imgURL, IsPrimary: true}}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(phones)
}

func handleUserFavorites(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := `
		SELECT p.id, p.user_id, p.title, p.brand_id, b.name, p.model, p.storage, 
		       p.battery_health, p.price, p.condition, p.description, p.contact_phone, p.views_count, p.status, p.created_at, p.bumped_at,
			   1 as is_favorited
		FROM phones p
		JOIN brands b ON p.brand_id = b.id
		JOIN favorites f ON p.id = f.phone_id
		WHERE f.user_id = ? AND p.status = 'active'
		ORDER BY f.created_at DESC
	`
	rows, err := DB.Query(query, userID)
	if err != nil {
		http.Error(w, "Error fetching favorites", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	phones := []Phone{}
	for rows.Next() {
		var p Phone
		if err := rows.Scan(&p.ID, &p.UserID, &p.Title, &p.BrandID, &p.BrandName, &p.Model, &p.Storage,
			&p.BatteryHealth, &p.Price, &p.Condition, &p.Description, &p.ContactPhone, &p.ViewsCount, &p.Status, &p.CreatedAt, &p.BumpedAt, &p.IsFavorited); err != nil {
			continue
		}
		if len(p.ContactPhone) > 6 {
			p.ContactPhone = p.ContactPhone[:6] + " *** ** " + p.ContactPhone[len(p.ContactPhone)-2:]
		}
		phones = append(phones, p)
	}

	for i, p := range phones {
		var imgURL string
		err := DB.QueryRow(`SELECT image_url FROM images WHERE phone_id = ? AND is_primary = 1 LIMIT 1`, p.ID).Scan(&imgURL)
		if err == nil {
			phones[i].Images = []Image{{ImageURL: imgURL, IsPrimary: true}}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(phones)
}

func handlePhoneContact(w http.ResponseWriter, phoneID int) {
	// Инкремент просмотров
	DB.Exec(`UPDATE phones SET views_count = views_count + 1 WHERE id = ?`, phoneID)

	var contactPhone string
	err := DB.QueryRow(`SELECT contact_phone FROM phones WHERE id = ?`, phoneID).Scan(&contactPhone)
	if err != nil {
		http.Error(w, "Phone not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"contact_phone": contactPhone})
}

func handlePhoneBump(w http.ResponseWriter, r *http.Request, phoneID int) {
	userID, ok := getUserIDFromSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var ownerID int
	var bumpedAt time.Time
	err := DB.QueryRow(`SELECT user_id, bumped_at FROM phones WHERE id = ?`, phoneID).Scan(&ownerID, &bumpedAt)
	if err != nil {
		http.Error(w, "Phone not found", http.StatusNotFound)
		return
	}

	if ownerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if 24 hours have passed since last bump
	if time.Since(bumpedAt) < 24*time.Hour {
		http.Error(w, "Вы можете поднимать объявление не чаще одного раза в 24 часа", http.StatusTooManyRequests)
		return
	}

	_, err = DB.Exec(`UPDATE phones SET bumped_at = CURRENT_TIMESTAMP WHERE id = ?`, phoneID)
	if err != nil {
		http.Error(w, "Error bumping phone", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Объявление успешно поднято в топ!"}`))
}
