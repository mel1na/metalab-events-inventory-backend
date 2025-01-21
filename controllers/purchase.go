package controllers

import (
	"fmt"
	"metalab/events-inventory-tracker/models"
	"metalab/events-inventory-tracker/sumup_integration"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreatePurchaseInput struct {
	Items       []models.Item `json:"items" binding:"required"`
	PaymentType string        `json:"payment_type" binding:"required"`
	Tip         uint          `json:"tip"`
}

func CreatePurchase(c *gin.Context) {
	var input CreatePurchaseInput
	var finalCost uint = 0
	var transaction_id string = ""
	returnedItemsArray := []models.Item{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, v := range input.Items {
		item := FindItemById(v.ItemId)
		finalCost += (item.Price * v.Quantity)
		returnedItemsArray = append(returnedItemsArray, models.Item{ItemId: v.ItemId, Name: item.Name, Quantity: v.Quantity, Price: item.Price})
	}

	finalCost += input.Tip
	if input.PaymentType == "card" {
		var err error
		transaction_id, err = sumup_integration.StartReaderCheckout(string(*FindReaderIdByName("Bar")), finalCost)
		if err != nil {
			fmt.Printf("error while creating reader checkout: %s\n", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	purchase := models.Purchase{Items: returnedItemsArray, PaymentType: input.PaymentType, TransactionId: transaction_id, Tip: input.Tip, FinalCost: finalCost}
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

	if err := models.DB.Where("purchase_id = ?", c.Param("id")).First(&purchase).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": purchase})
}

/*func FindPurchaseByTransactionId(id string) (*models.Purchase, error) {
	var purchase models.Purchase
	if err := models.DB.Where("transaction_id = ?", id).First(&purchase).Error; err != nil {
		return nil, err
	}
	return &purchase, nil
}*/

type UpdatePurchaseInput struct {
	Items       []models.Item `json:"items" binding:"required"`
	PaymentType string        `json:"payment_type" binding:"required"`
	Tip         uint          `json:"tip"`
}

func UpdatePurchase(c *gin.Context) {
	var purchase models.Purchase
	if err := models.DB.Where("purchase_id = ?", c.Param("id")).First(&purchase).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	var input UpdatePurchaseInput
	var finalCost uint = 0
	returnArray := []models.Item{}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, v := range input.Items {
		item := FindItemById(v.ItemId)
		if item.Name == "No item found" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "itemid " + strconv.FormatUint(uint64(v.ItemId), 10) + " not found"})
		}
		finalCost += (item.Price * v.Quantity)
		returnArray = append(returnArray, models.Item{ItemId: v.ItemId, Name: item.Name, Quantity: v.Quantity, Price: item.Price})
	}

	finalCost += input.Tip
	updatedPurchase := models.Purchase{Items: returnArray, PaymentType: input.PaymentType, Tip: input.Tip, FinalCost: finalCost}

	models.DB.Model(&purchase).Updates(&updatedPurchase)
	c.JSON(http.StatusOK, gin.H{"data": purchase})
}

func DeletePurchase(c *gin.Context) {
	var purchase models.Purchase
	if err := models.DB.Where("purchase_id = ?", c.Param("id")).First(&purchase).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	models.DB.Delete(&purchase)
	c.JSON(http.StatusOK, gin.H{"data": "success"})
}
