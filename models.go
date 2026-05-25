package main

import (
	"time"
)

type User struct {
	ID            int       `json:"id"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	Name          string    `json:"name"`
	AvatarURL     string    `json:"avatar_url"`
	CityID        *int      `json:"city_id,omitempty"`
	OAuthProvider *string   `json:"oauth_provider,omitempty"`
	OAuthID       *string   `json:"oauth_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	AverageRating float64   `json:"average_rating"`
	ReviewCount   int       `json:"review_count"`
}

type Session struct {
	SessionToken string    `json:"session_token"`
	UserID       int       `json:"user_id"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type Brand struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type City struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Review struct {
	ID         int       `json:"id"`
	ReviewerID int       `json:"reviewer_id"`
	SellerID   int       `json:"seller_id"`
	Rating     int       `json:"rating"`
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
}

type Image struct {
	ID        int    `json:"id"`
	PhoneID   int    `json:"phone_id"`
	ImageURL  string `json:"image_url"`
	IsPrimary bool   `json:"is_primary"`
}

type Phone struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	Title         string    `json:"title"`
	BrandID       int       `json:"brand_id"`
	BrandName     string    `json:"brand_name,omitempty"` // For joining
	CityID        *int      `json:"city_id,omitempty"`
	CityName      string    `json:"city_name,omitempty"`
	Model         string    `json:"model"`
	Storage       int       `json:"storage"`
	BatteryHealth int       `json:"battery_health"`
	Price         int       `json:"price"`
	Condition     string    `json:"condition"`
	Description   string    `json:"description"`
	ContactPhone  string    `json:"contact_phone"`
	ViewsCount    int       `json:"views_count"`
	Status        string    `json:"status"`
	IsFavorited   bool      `json:"is_favorited"`
	CreatedAt     time.Time `json:"created_at"`
	BumpedAt      time.Time `json:"bumped_at"`
	Images        []Image   `json:"images,omitempty"`
	PromoSlug     string    `json:"promo_slug,omitempty"`
}

type Message struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"sender_id"`
	ReceiverID int       `json:"receiver_id"`
	PhoneID    int       `json:"phone_id"`
	Content    string    `json:"content"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}

type ChatDialog struct {
	OtherUserID   int       `json:"other_user_id"`
	OtherUserEmail string   `json:"other_user_email"`
	PhoneID       int       `json:"phone_id"`
	PhoneTitle    string    `json:"phone_title"`
	LastMessage   string    `json:"last_message"`
	LastMessageAt time.Time `json:"last_message_at"`
	UnreadCount   int       `json:"unread_count"`
}

type PromoPage struct {
	ID          int       `json:"id"`
	PhoneID     int       `json:"phone_id"`
	Slug        string    `json:"slug"`
	HTMLContent string    `json:"html_content"`
	ViewsCount  int       `json:"views_count"`
	CreatedAt   time.Time `json:"created_at"`
}
