package models

import "time"

type Item struct {
	ItemID    uint      `json:"id" gorm:"primaryKey;unique"`
	Name      string    `json:"name"`
	Quantity  uint      `json:"quantity,omitempty"`
	Price     float32   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}
