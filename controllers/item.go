package controllers

import (
	"metalab/events-inventory-tracker/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreateItemInput struct {
	Name     string  `json:"name" binding:"required"`
	Quantity uint    `json:"quantity"`
	Price    float32 `json:"price"` //not required so price can be set to 0 and still work
}

func CreateItem(c *gin.Context) {
	var input CreateItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := models.Item{Name: input.Name, Quantity: input.Quantity, Price: input.Price}
	models.DB.Create(&item)

	c.JSON(http.StatusOK, gin.H{"data": item})
}

func FindItems(c *gin.Context) {
	var items []models.Item
	models.DB.Find(&items)

	c.JSON(http.StatusOK, gin.H{"data": items})
}

func FindItem(c *gin.Context) {
	var item models.Item

	if err := models.DB.Where("item_id = ?", c.Param("id")).First(&item).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": item})
}

func FindItemById(id uint) models.Item {
	var item models.Item

	if err := models.DB.Where("item_id = ?", id).First(&item).Error; err != nil {
		return models.Item{Name: "No item found", Quantity: 0, Price: 0.00}
	}

	return item
}

type UpdateItemInput struct {
	Name     string  `json:"name" binding:"required"`
	Quantity uint    `json:"quantity"`
	Price    float32 `json:"price"` //not required so price can be set to 0 and still work
}

func UpdateItem(c *gin.Context) {
	var item models.Item
	if err := models.DB.Where("item_id = ?", c.Param("id")).First(&item).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	var input UpdateItemInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedItem := models.Item{Name: input.Name, Quantity: input.Quantity, Price: input.Price}

	models.DB.Model(&item).Updates(&updatedItem)
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func DeleteItem(c *gin.Context) {
	var item models.Item
	if err := models.DB.Where("item_id = ?", c.Param("id")).First(&item).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	models.DB.Delete(&item)
	c.JSON(http.StatusOK, gin.H{"data": "success"})
}
