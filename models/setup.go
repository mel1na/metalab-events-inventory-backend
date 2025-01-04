package models

import (
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
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

	if database.Limit(1).Find(&User{Name: "admin"}).RowsAffected == 0 {

		key := []byte(os.Getenv("JWT_SECRET"))
		t := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
			"iss":   "metalab-events-backend",
			"sub":   "admin",
			"admin": true,
		})
		s, err := t.SignedString(key)
		fmt.Println(s)
		if err != nil {
			fmt.Println(err)
		} else {
			database.Create(&User{Name: "admin", Token: s})
			fmt.Printf("Default admin user created with token %s\n", s)
		}
	}

	DB = database
}
