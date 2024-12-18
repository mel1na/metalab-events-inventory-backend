package models

import "time"

type Purchase struct {
	ID          uint      `json:"id"`
	Items       []Item    `json:"items" gorm:"foreignKey:ParentID;type:bytes;serializer:gob"`
	PaymentType string    `json:"payment_type"`
	Tip         float32   `json:"tip,omitempty"`
	FinalCost   float32   `json:"final_cost"`
	CreatedAt   time.Time `json:"created_at"`
}
