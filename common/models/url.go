package models

import (
	"time"
)

// URLMapping represents the mapping between a shortened URL code and its original URL
type URLMapping struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Code      string    `json:"code" gorm:"uniqueIndex"`
	LongURL   string    `json:"long_url" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	Visits    int       `json:"visits" gorm:"default:0"`
}
