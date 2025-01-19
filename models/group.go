package models

import "time"

type Group struct {
	GroupID   uint      `json:"id" gorm:"primaryKey;unique"`
	Name      string    `json:"name"`
	Items     []Item    `json:"items" gorm:"foreignKey:ItemID;type:bytes;serializer:gob"`
	CreatedAt time.Time `json:"created_at"`
}
