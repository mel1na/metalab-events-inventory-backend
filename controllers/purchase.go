package controllers

import (
	"fmt"
	"metalab/events-inventory-tracker/models"
	sumup_models "metalab/events-inventory-tracker/models/sumup"
	"metalab/events-inventory-tracker/sumup_integration"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type CreatePurchaseInput struct {
	Items       []models.Item `json:"items" binding:"required"`
	PaymentType string        `json:"payment_type" binding:"required"`
	ReaderId    string        `json:"reader_id"`
	Tip         uint          `json:"tip"`
}

func CreatePurchase(c *gin.Context) {
	var input CreatePurchaseInput
	var finalCost int = 0
	var clientTransactionId = ""
	var transactionDescription []string
	var transactionStatus sumup_models.TransactionFullStatus
	returnedItemsArray := []models.Item{}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, v := range input.Items {
		item := FindItemById(v.ItemId)
		finalCost += item.Price * int(v.Quantity)
		returnedItemsArray = append(returnedItemsArray, models.Item{ItemId: v.ItemId, Name: item.Name, Quantity: v.Quantity, Price: item.Price})
		transactionDescription = append(transactionDescription, fmt.Sprintf("%dx %s", v.Quantity, item.Name))
	}

	finalCost += int(input.Tip)
	var finalTransactionDescription = strings.Join(transactionDescription[:], ", ")
	if input.Tip > 0 {
		finalTransactionDescription += fmt.Sprintf(" + %.2f Tip", float64(input.Tip)/100)
	}
	if finalCost < 0 {
		finalCost = 0
	}
	if input.PaymentType == "card" {
		if finalCost <= 0 {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid purchase amount"})
			return
		}
		var err error
		transactionStatus = sumup_models.TransactionFullStatusPending
		clientTransactionId, err = sumup_integration.StartReaderCheckout(input.ReaderId, uint(finalCost), &finalTransactionDescription)
		if err != nil {
			fmt.Printf("error while creating reader checkout: %s\n", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	if input.PaymentType == "cash" {
		transactionStatus = sumup_models.TransactionFullStatusSuccessful
	}
	purchase := models.Purchase{Items: returnedItemsArray, PaymentType: input.PaymentType, ClientTransactionId: clientTransactionId, TransactionStatus: transactionStatus, Tip: input.Tip, FinalCost: uint(finalCost), CreatedBy: c.GetString("jwt-claim-sub")}
	models.DB.Create(&purchase)

	c.JSON(http.StatusOK, gin.H{"data": purchase})
}

func FindPurchases(c *gin.Context) {
	var purchases []models.Purchase
	models.DB.Find(&purchases)

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{"data": purchases})
}

func FindPurchase(c *gin.Context) {
	var purchase models.Purchase

	if err := models.DB.Where("purchase_id = ?", c.Param("id")).First(&purchase).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{"data": purchase})
}

func FindPurchaseByTransactionId(id string) (*models.Purchase, error) {
	var purchase models.Purchase
	if err := models.DB.Where("client_transaction_id = ?", id).First(&purchase).Error; err != nil {
		return nil, err
	}
	return &purchase, nil
}

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
		if item.Price >= 0 {
			finalCost += uint(item.Price) * v.Quantity
		} else {
			finalCost -= uint(item.Price) * v.Quantity
		}
		returnArray = append(returnArray, models.Item{ItemId: v.ItemId, Name: item.Name, Quantity: v.Quantity, Price: item.Price})
	}

	finalCost += input.Tip
	if finalCost < 0 {
		finalCost = 0
	}
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
