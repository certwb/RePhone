package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// generateRandomSlug generates a random 6-character hex string
func generateRandomSlug() string {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)
}

func handleGeneratePromo(w http.ResponseWriter, r *http.Request, phoneID int) {
	userID, ok := getUserIDFromSession(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var p Phone
	var ownerID int
	err := DB.QueryRow(`
		SELECT p.user_id, p.title, p.brand_id, b.name, p.model, p.storage, 
		       p.battery_health, p.price, p.condition, p.description, p.contact_phone 
		FROM phones p
		JOIN brands b ON p.brand_id = b.id
		WHERE p.id = ? AND p.status != 'deleted'
	`, phoneID).Scan(&ownerID, &p.Title, &p.BrandID, &p.BrandName, &p.Model, &p.Storage,
		&p.BatteryHealth, &p.Price, &p.Condition, &p.Description, &p.ContactPhone)

	if err != nil {
		http.Error(w, "Phone not found", http.StatusNotFound)
		return
	}

	if ownerID != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if promo already exists
	var existingSlug string
	err = DB.QueryRow(`SELECT slug FROM promo_pages WHERE phone_id = ?`, phoneID).Scan(&existingSlug)
	if err == nil {
		// Already exists
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message":    "Лендинг уже существует",
			"promo_slug": existingSlug,
		})
		return
	}

	// For now, we use a MOCK generator instead of real OpenAI call
	slug := fmt.Sprintf("%s-%s-%s", strings.ToLower(p.BrandName), strings.ToLower(p.Model), generateRandomSlug())
	slug = strings.ReplaceAll(slug, " ", "-")

	// Get images
	var imageURL string
	DB.QueryRow(`SELECT image_url FROM images WHERE phone_id = ? AND is_primary = 1 LIMIT 1`, phoneID).Scan(&imageURL)
	if imageURL == "" {
		imageURL = "https://placehold.co/600x600?text=No+Image"
	}

	htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s — Особое предложение</title>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;800&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-color: #0f172a;
            --surface-color: rgba(30, 41, 59, 0.7);
            --primary-color: #6366f1;
            --text-main: #f8fafc;
            --text-muted: #94a3b8;
        }
        body {
            margin: 0;
            font-family: 'Inter', sans-serif;
            background: var(--bg-color);
            color: var(--text-main);
            overflow-x: hidden;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            padding: 20px;
        }
        .hero {
            text-align: center;
            padding: 60px 20px 40px;
        }
        .hero h1 {
            font-size: 2.5rem;
            font-weight: 800;
            margin-bottom: 20px;
            background: linear-gradient(135deg, #a855f7, #6366f1);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        .hero img {
            max-width: 100%%;
            height: auto;
            border-radius: 24px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.4);
            margin: 20px 0;
            transition: transform 0.3s ease;
        }
        .hero img:hover {
            transform: scale(1.02);
        }
        .price-badge {
            display: inline-block;
            background: rgba(99, 102, 241, 0.1);
            border: 1px solid rgba(99, 102, 241, 0.3);
            color: #818cf8;
            padding: 10px 24px;
            border-radius: 100px;
            font-size: 1.5rem;
            font-weight: 800;
            margin-bottom: 30px;
        }
        .grid-stats {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin: 40px 0;
        }
        .stat-card {
            background: var(--surface-color);
            backdrop-filter: blur(10px);
            padding: 24px;
            border-radius: 20px;
            text-align: center;
            border: 1px solid rgba(255,255,255,0.05);
        }
        .stat-card .value {
            font-size: 1.8rem;
            font-weight: 800;
            color: var(--text-main);
            margin-bottom: 8px;
        }
        .stat-card .label {
            color: var(--text-muted);
            font-size: 0.9rem;
        }
        .story {
            background: linear-gradient(180deg, rgba(30,41,59,0) 0%%, var(--surface-color) 100%%);
            padding: 40px 20px;
            border-radius: 24px;
            margin-bottom: 80px;
        }
        .story h2 {
            font-size: 1.5rem;
            margin-bottom: 16px;
        }
        .story p {
            line-height: 1.7;
            color: var(--text-muted);
            font-size: 1.1rem;
        }
        .sticky-footer {
            position: fixed;
            bottom: 0;
            left: 0;
            right: 0;
            background: rgba(15, 23, 42, 0.9);
            backdrop-filter: blur(10px);
            padding: 20px;
            text-align: center;
            border-top: 1px solid rgba(255,255,255,0.1);
            z-index: 100;
        }
        .btn-cta {
            background: var(--primary-color);
            color: white;
            border: none;
            padding: 16px 40px;
            border-radius: 100px;
            font-size: 1.2rem;
            font-weight: 600;
            cursor: pointer;
            width: 100%%;
            max-width: 400px;
            transition: all 0.3s ease;
            box-shadow: 0 10px 20px rgba(99, 102, 241, 0.3);
        }
        .btn-cta:hover {
            transform: translateY(-2px);
            box-shadow: 0 15px 30px rgba(99, 102, 241, 0.4);
        }
        @media (max-width: 600px) {
            .hero h1 { font-size: 2rem; }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="hero">
            <h1>Тот самый %s %s в идеале, который вы искали</h1>
            <div class="price-badge">%d ₸</div>
            <img src="%s" alt="Phone Image">
        </div>

        <div class="grid-stats">
            <div class="stat-card">
                <div class="value">%d%%</div>
                <div class="label">Батарея (Оригинал)</div>
            </div>
            <div class="stat-card">
                <div class="value">%d ГБ</div>
                <div class="label">Память</div>
            </div>
            <div class="stat-card">
                <div class="value">%s</div>
                <div class="label">Состояние</div>
            </div>
        </div>

        <div class="story">
            <h2>Честная история от владельца</h2>
            <p>«%s»</p>
        </div>
    </div>

    <div class="sticky-footer">
        <button class="btn-cta" onclick="openChat()">Написать владельцу</button>
    </div>

    <script>
        function openChat() {
            // Если лендинг открыт в iframe на главном сайте
            if (window.parent !== window) {
                window.parent.postMessage({action: 'openChat', phoneId: %d, sellerId: %d}, '*');
            } else {
                // Если открыт по прямой ссылке, редиректим на сайт
                window.location.href = '/?action=chat&phone_id=%d';
            }
        }
    </script>
</body>
</html>`, p.BrandName, p.Model, p.Price, imageURL, p.BatteryHealth, p.Storage, p.Condition, p.Description, phoneID, ownerID, phoneID)

	_, err = DB.Exec(`INSERT INTO promo_pages (phone_id, slug, html_content) VALUES (?, ?, ?)`, phoneID, slug, htmlContent)
	if err != nil {
		http.Error(w, "Failed to save promo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Лендинг успешно создан!",
		"promo_slug": slug,
	})
}

func handleServePromo(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/promo/")
	if slug == "" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	var htmlContent string
	err := DB.QueryRow(`SELECT html_content FROM promo_pages WHERE slug = ?`, slug).Scan(&htmlContent)
	if err != nil {
		http.Error(w, "Promo page not found", http.StatusNotFound)
		return
	}

	// Увеличиваем счетчик просмотров
	DB.Exec(`UPDATE promo_pages SET views_count = views_count + 1 WHERE slug = ?`, slug)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(htmlContent))
}
