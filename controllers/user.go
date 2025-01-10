package controllers

import (
	"fmt"
	"metalab/events-inventory-tracker/models"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

	key := []byte(os.Getenv("JWT_SECRET"))
	t := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"iss":   "metalab-events-backend",
		"sub":   input.Name,
		"admin": input.IsAdmin,
	})
	s, err := t.SignedString(key)
	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	} else {
		user := models.User{Name: input.Name, Token: s, IsAdmin: input.IsAdmin}
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

	if err := models.DB.Where("id = ?", c.Param("id")).First(&user).Error; err != nil {
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
	if err := models.DB.Where("id = ?", c.Param("id")).First(&user).Error; err != nil {
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
	if err := models.DB.Where("id = ?", c.Param("id")).First(&user).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	models.DB.Delete(&user)
	c.JSON(http.StatusOK, gin.H{"data": "success"})
}
