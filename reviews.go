package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// handleReviews processes POST /api/reviews and GET /api/users/{id}/reviews
func handleReviews(w http.ResponseWriter, r *http.Request) {
	// 1. GET /api/users/{id}/reviews - is not handled here perfectly due to path mapping
	// Let's assume we map /api/reviews (POST to create)
	// and GET /api/reviews?user_id=123 (GET to list reviews)
	
	if r.Method == http.MethodGet {
		sellerIDStr := r.URL.Query().Get("seller_id")
		sellerID, err := strconv.Atoi(sellerIDStr)
		if err != nil {
			http.Error(w, "Invalid seller_id", http.StatusBadRequest)
			return
		}

		rows, err := DB.Query(`
			SELECT r.id, r.reviewer_id, r.rating, r.comment, r.created_at, u.name as reviewer_name, u.avatar_url as reviewer_avatar
			FROM reviews r
			JOIN users u ON r.reviewer_id = u.id
			WHERE r.seller_id = ?
			ORDER BY r.created_at DESC
		`, sellerID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type ReviewWithUser struct {
			Review
			ReviewerName   string `json:"reviewer_name"`
			ReviewerAvatar string `json:"reviewer_avatar"`
		}

		var reviews []ReviewWithUser
		for rows.Next() {
			var rev ReviewWithUser
			if err := rows.Scan(&rev.ID, &rev.ReviewerID, &rev.Rating, &rev.Comment, &rev.CreatedAt, &rev.ReviewerName, &rev.ReviewerAvatar); err == nil {
				reviews = append(reviews, rev)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(reviews)
		return
	}

	if r.Method == http.MethodPost {
		reviewerID, ok := getUserIDFromSession(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req struct {
			SellerID int    `json:"seller_id"`
			Rating   int    `json:"rating"`
			Comment  string `json:"comment"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if req.Rating < 1 || req.Rating > 5 {
			http.Error(w, "Rating must be between 1 and 5", http.StatusBadRequest)
			return
		}

		if reviewerID == req.SellerID {
			http.Error(w, "You cannot review yourself", http.StatusBadRequest)
			return
		}

		_, err := DB.Exec(`
			INSERT INTO reviews (reviewer_id, seller_id, rating, comment)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(reviewer_id, seller_id) DO UPDATE SET 
			rating=excluded.rating, comment=excluded.comment, created_at=?
		`, reviewerID, req.SellerID, req.Rating, strings.TrimSpace(req.Comment), time.Now())

		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Уведомим продавца о новом отзыве, если он онлайн
		NotifyUser(req.SellerID, "new_review", map[string]interface{}{
			"reviewer_id": reviewerID,
			"rating":      req.Rating,
		})

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message": "Review submitted successfully"}`))
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleUserInfo returns public information about a user
func handleUserInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userIDStr := r.URL.Query().Get("id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user id", http.StatusBadRequest)
		return
	}

	var u User
	err = DB.QueryRow(`
		SELECT u.id, u.name, u.avatar_url, u.city_id, u.created_at,
			   COALESCE(AVG(r.rating), 0) as average_rating,
			   COUNT(r.id) as review_count
		FROM users u
		LEFT JOIN reviews r ON u.id = r.seller_id
		WHERE u.id = ?
		GROUP BY u.id
	`, userID).
		Scan(&u.ID, &u.Name, &u.AvatarURL, &u.CityID, &u.CreatedAt, &u.AverageRating, &u.ReviewCount)

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// We don't send Email, PasswordHash, OAuth stuff here since it's a public profile.
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(u)
}
