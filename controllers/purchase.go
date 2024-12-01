package controllers

import (
	"metalab/events-inventory-tracker/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CreatePurchaseInput struct {
	Items       []models.Item `json:"items" binding:"required"`
	PaymentType string        `json:"payment_type" binding:"required"`
	Tip         float32       `json:"tip"`
}

func CreatePurchase(c *gin.Context) {
	var input CreatePurchaseInput
	var finalCost = float32(0.00)
	returnArray := []models.Item{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, v := range input.Items {
		item := FindItemById(v.ID)
		finalCost += (item.Price * float32(v.Quantity)) + input.Tip
		returnArray = append(returnArray, models.Item{ID: v.ID, Name: item.Name, Quantity: v.Quantity, Price: item.Price})
	}

	purchase := models.Purchase{Items: returnArray, PaymentType: input.PaymentType, Tip: input.Tip, FinalCost: finalCost}
	models.DB.Create(&purchase)

	c.JSON(http.StatusOK, gin.H{"data": purchase})
}

func FindPurchases(c *gin.Context) {
	var purchases []models.Purchase
	models.DB.Find(&purchases)

	c.JSON(http.StatusOK, gin.H{"data": purchases})
}

func FindPurchase(c *gin.Context) {
	var purchase models.Purchase

	if err := models.DB.Where("id = ?", c.Param("id")).First(&purchase).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": purchase})
}

type UpdatePurchaseInput struct {
	Items       []models.Item `json:"items" binding:"required"`
	PaymentType string        `json:"payment_type" binding:"required"`
	Tip         float32       `json:"tip"`
}

func UpdatePurchase(c *gin.Context) {
	var purchase models.Purchase
	if err := models.DB.Where("id = ?", c.Param("id")).First(&purchase).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	var input UpdatePurchaseInput
	finalCost := float32(0.00)
	returnArray := []models.Item{}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, v := range input.Items {
		item := FindItemById(v.ID)
		finalCost += (item.Price * float32(v.Quantity)) + input.Tip
		returnArray = append(returnArray, models.Item{ID: v.ID, Name: item.Name, Quantity: v.Quantity, Price: item.Price})
	}

	updatedPurchase := models.Purchase{Items: returnArray, PaymentType: input.PaymentType, Tip: input.Tip, FinalCost: finalCost}

	models.DB.Model(&purchase).Updates(&updatedPurchase)
	c.JSON(http.StatusOK, gin.H{"data": purchase})
}

func DeletePurchase(c *gin.Context) {
	var purchase models.Purchase
	if err := models.DB.Where("id = ?", c.Param("id")).First(&purchase).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	models.DB.Delete(&purchase)
	c.JSON(http.StatusOK, gin.H{"data": "success"})
}
