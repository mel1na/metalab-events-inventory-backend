package models

import (
	sumup_models "metalab/events-inventory-tracker/models/sumup"
	"time"

	"github.com/google/uuid"
)

type Purchase struct {
	PurchaseId          uuid.UUID                          `json:"id" gorm:"primaryKey;unique;type:uuid;default:gen_random_uuid()"`
	Items               []Item                             `json:"items" gorm:"foreignKey:ItemID;type:bytes;serializer:gob"`
	PaymentType         string                             `json:"payment_type"`
	TransactionStatus   sumup_models.TransactionFullStatus `json:"status"`
	ClientTransactionId string                             `json:"transaction_id,omitempty"`
	Tip                 uint                               `json:"tip,omitempty"`
	FinalCost           uint                               `json:"final_cost"`
	CreatedAt           time.Time                          `json:"created_at"`
	CreatedBy           string                             `json:"created_by"`
}

// PaymentStatus: The status of the payment object gives information about the current state of the payment.
//
// Possible values:
//
// - `unknown` - The payment status is unknown.
// - `unpaid` - The payment has not been paid for (yet).
// - `paid` - The payment has been processed and paid.
type PaymentStatus string

const (
	PaymentStatusUnknown PaymentStatus = "unknown"
	PaymentStatusUnpaid  PaymentStatus = "unpaid"
	PaymentStatusPaid    PaymentStatus = "paid"
)

// PaymentType: The type of the payment object gives information about the type of payment.
//
// Possible values:
//
// - `cash` - The payment was made with cash.
// - `unpaid` - The payment was made with a credit/debit card.
// - `other` - The payment was made in an unspecified way.
type PaymentType string

const (
	PaymentTypeCash  PaymentType = "cash"
	PaymentTypeCard  PaymentType = "card"
	PaymentTypeOther PaymentType = "other"
)
