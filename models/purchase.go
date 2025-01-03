package models

import (
	"time"

	"github.com/google/uuid"
)

type Purchase struct {
	PurchaseID  uuid.UUID `json:"id" gorm:"primaryKey;unique;type:uuid;default:gen_random_uuid()"`
	Items       []Item    `json:"items" gorm:"foreignKey:ItemID;type:bytes;serializer:gob"`
	PaymentType string    `json:"payment_type"`
	Tip         float32   `json:"tip,omitempty"`
	FinalCost   float32   `json:"final_cost"`
	CreatedAt   time.Time `json:"created_at"`
}
