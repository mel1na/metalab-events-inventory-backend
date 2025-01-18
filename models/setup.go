package models

import (
	"strings"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	db_host := os.Getenv("DB_HOST")
	db_user := os.Getenv("DB_USER")
	db_pass := os.Getenv("DB_PASS")
	db_name := os.Getenv("DB_NAME")
	if db_host == "" {
		db_host = "events-postgres"
	}
	if db_user == "" {
		db_user = "test"
	}
	if db_pass == "" {
		db_pass = "test"
	}
	if db_name == "" {
		db_name = "test"
	}
	dsn := strings.Join([]string{"host=", db_host, " user=", db_user, " password=", db_pass, " dbname=", db_name, " port=5432 sslmode=disable timezone=Europe/Vienna"}, "")
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
			"admin": "true",
		})
		s, err := t.SignedString(key)
		if err != nil {
			fmt.Println(err)
		} else {
			database.Create(&User{Name: "admin", Token: s, IsAdmin: "true"})
			fmt.Printf("Default admin user created with token %s\n", s)
		}
	}

	DB = database
}
