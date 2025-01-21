package models

type Group struct {
	GroupId uint   `json:"id" gorm:"primaryKey;unique"`
	Name    string `json:"name"`
	Items   []Item `json:"items" gorm:"foreignKey:ItemID;type:bytes;serializer:gob"`
}
