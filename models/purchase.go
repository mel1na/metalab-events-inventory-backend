package models

import (
	"time"

	"github.com/google/uuid"
)

type Purchase struct {
	PurchaseId    uuid.UUID `json:"id" gorm:"primaryKey;unique;type:uuid;default:gen_random_uuid()"`
	Items         []Item    `json:"items" gorm:"foreignKey:ItemID;type:bytes;serializer:gob"`
	PaymentType   string    `json:"payment_type"`
	TransactionId string    `json:"transaction_id,omitempty"`
	Tip           uint      `json:"tip,omitempty"`
	FinalCost     uint      `json:"final_cost"`
	CreatedAt     time.Time `json:"created_at"`
}
