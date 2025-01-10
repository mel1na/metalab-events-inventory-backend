package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	UserID    uuid.UUID      `json:"id" gorm:"primaryKey;unique;type:uuid;default:gen_random_uuid()"`
	Name      string         `json:"name" gorm:"unique"`
	Token     string         `json:"token"`
	IsAdmin   string         `json:"is_admin" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
