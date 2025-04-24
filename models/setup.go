package models

import (
	"fmt"
	sumup_models "metalab/events-inventory-tracker/models/sumup"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := "host=events-postgres user=test password=test dbname=test port=5432 sslmode=disable timezone=Europe/Vienna"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // change the database provider if necessary

	if err != nil {
		panic("Failed to connect to database!")
	}

	database.AutoMigrate(&Item{})
	database.AutoMigrate(&Purchase{})
	database.AutoMigrate(&User{})
	database.AutoMigrate(&Group{})
	database.AutoMigrate(&Voucher{})
	database.AutoMigrate(&sumup_models.Reader{})

	if database.Limit(1).Find(&User{Name: "admin"}).RowsAffected == 0 {
		userId := uuid.New()

		key := []byte(os.Getenv("JWT_SECRET"))
		t := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"iss":    "metalab-events-backend",
			"sub":    "admin",
			"iat":    time.Now().Unix(),
			"userid": userId,
			"admin":  "true",
		})
		s, err := t.SignedString(key)
		if err != nil {
			fmt.Println(err)
		} else {
			database.Create(&User{UserID: userId, Name: "admin", Token: s, IsAdmin: "true", CreatedBy: uuid.Nil})
			fmt.Printf("Default admin user created with token %s\n", s)
		}
	}

	DB = database
}
