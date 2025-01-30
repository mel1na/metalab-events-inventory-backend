package controllers

import (
	"fmt"
	"metalab/events-inventory-tracker/models"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type CreateUserInput struct {
	Name    string `json:"name" binding:"required"`
	IsAdmin string `json:"is_admin"`
}

func CreateUser(c *gin.Context) {
	var input CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userId uuid.UUID = uuid.New()

	key := []byte(os.Getenv("JWT_SECRET"))
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"iss":    "metalab-events-backend",
		"sub":    input.Name,
		"iat":    time.Now().Unix(),
		"userid": userId,
		"admin":  input.IsAdmin,
	})
	s, err := t.SignedString(key)
	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		creator_id, err := uuid.Parse(c.GetString("jwt-claim-userid"))
		if err != nil {
			fmt.Println(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user := models.User{UserID: userId, Name: input.Name, Token: s, IsAdmin: input.IsAdmin, CreatedBy: creator_id}
		result := models.DB.Create(&user)
		if result.Error != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Bad Request"})
		} else {
			c.JSON(http.StatusOK, gin.H{"data": user})
		}
	}
}

func FindUsers(c *gin.Context) {
	var users []models.User
	models.DB.Find(&users)

	c.JSON(http.StatusOK, gin.H{"data": users})
}

func FindUser(c *gin.Context) {
	var user models.User

	if err := models.DB.Where("user_id = ?", c.Param("id")).First(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": user})
}

type UpdateUserInput struct {
	Name    string `json:"name" gorm:"unique"`
	Token   string `json:"token"`
	IsAdmin string `json:"is_admin" gorm:"default:false"`
}

func UpdateUser(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("user_id = ?", c.Param("id")).First(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	var input UpdateUserInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedUser := models.User{Name: input.Name, Token: input.Token, IsAdmin: input.IsAdmin}

	models.DB.Model(&user).Updates(&updatedUser)
	c.JSON(http.StatusOK, gin.H{"data": user})
}

func DeleteUser(c *gin.Context) {
	var user models.User
	if err := models.DB.Where("user_id = ?", c.Param("id")).First(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	models.DB.Delete(&user)
	c.JSON(http.StatusOK, gin.H{"data": "success"})
}
