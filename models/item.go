package models

type Item struct {
	ItemId   uint   `json:"id" gorm:"primaryKey;unique"`
	Name     string `json:"name"`
	Quantity uint   `json:"quantity,omitempty"`
	Price    uint   `json:"price"`
}
