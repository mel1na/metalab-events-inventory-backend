package models

type Item struct {
	ID       uint    `json:"id" gorm:"primaryKey"`
	ParentID uint    `json:"parent_id,omitempty"`
	Name     string  `json:"name"`
	Quantity uint    `json:"quantity,omitempty"`
	Price    float32 `json:"price"`
	//CreatedAt time.Time `json:"created_at"`
}
