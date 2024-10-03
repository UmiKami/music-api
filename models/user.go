package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID			uint		`gorm:"primaryKey"`
	Username	string		`gorm:"size:255"`
	Email		string		`gorm:"size:255"`
	Password	string		`gorm:"size:255"`
	LastLogin	time.Time
	UpdatedAt	time.Time
	CreatedAt	time.Time	`gorm:"autoCreateTime"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	LastLogin time.Time `json:"last_login"`
}

func MigrateUser(db *gorm.DB) error {
	return db.AutoMigrate(&User{})
}