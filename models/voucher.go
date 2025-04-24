package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Voucher struct {
	VoucherId    uuid.UUID `json:"id" gorm:"primaryKey;unique;type:uuid;default:gen_random_uuid()"`
	CreditAmount uint      `json:"credit_amount"`
	ValidItems   []Item    `json:"valid_items"`
	ValidGroups  []Group   `json:"valid_groups"`
	//VoucherType VoucherType `json:"voucher_type,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	CreatedBy uuid.UUID      `json:"created_by"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type VoucherData struct {
	VoucherId         uuid.UUID `json:"id" gorm:"foreignKey:VoucherID;type:bytes;serializer:gob"`
	VoucherCreditUsed uint      `json:"credit_used"`
	VoucherItemsUsed  []Item    `json:"items_used"`
	VoucherGroupsUsed []Group   `json:"groups_used"`
}

/*type VoucherType string

const (
	VoucherTypeCredit   VoucherType = "credit"
	VoucherTypeItem     VoucherType = "item"
	VoucherTypeCategory VoucherType = "category"
)*/
