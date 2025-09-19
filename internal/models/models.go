package models

import (
	"time"
)

type User struct {
	ID               uint   `gorm:"primaryKey" json:"id"`
	Email            string `gorm:"uniqueIndex;size:255" json:"email"`
	Password         string `json:"-"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
	CurrentSessionID *uint `json:"-"`
}

type Session struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	UserID    uint `gorm:"index" json:"userId"`
	CreatedAt time.Time
	RevokedAt *time.Time
}

type RateSnapshot struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Provider  string    `gorm:"size:64" json:"provider"`
	Base      string    `gorm:"size:3" json:"base"`
	RatesJSON string    `gorm:"type:text" json:"-"`
	FetchedAt time.Time `json:"fetchedAt"`
	CreatedAt time.Time
}
