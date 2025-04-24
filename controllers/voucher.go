package controllers

import (
	"fmt"
	"metalab/events-inventory-tracker/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateVoucherInput struct {
	CreditAmount uint           `json:"credit_amount"` //defined in cents
	ValidItems   []models.Item  `json:"valid_items"`
	ValidGroups  []models.Group `json:"valid_groups"`
}

func CreateVoucher(c *gin.Context) {
	var input CreateVoucherInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.CreditAmount == 0 && len(input.ValidItems) == 0 && len(input.ValidGroups) == 0 {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "either credit_amount, valid_items or valid_groups must be specified"})
		return
	}

	voucherId := uuid.New()

	creator_id, err := uuid.Parse(c.GetString("jwt-claim-userid"))
	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	voucher := models.Voucher{VoucherId: voucherId, CreditAmount: input.CreditAmount, ValidItems: input.ValidItems, ValidGroups: input.ValidGroups, CreatedBy: creator_id}
	result := models.DB.Create(&voucher)
	if result.Error != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Bad Request"})
	} else {
		c.JSON(http.StatusOK, gin.H{"data": voucher})
	}
}

func FindVouchers(c *gin.Context) {
	var vouchers []models.Voucher
	models.DB.Find(&vouchers)

	c.JSON(http.StatusOK, gin.H{"data": vouchers})
}

func FindVoucher(c *gin.Context) {
	var voucher models.Voucher

	if err := models.DB.Where("voucher_id = ?", c.Param("id")).First(&voucher).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": voucher})
}

func UpdateVoucher(c *gin.Context) {
	var voucher models.Voucher
	if err := models.DB.Where("voucher_id = ?", c.Param("id")).First(&voucher).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	var input CreateVoucherInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedVoucher := models.Voucher{CreditAmount: input.CreditAmount, ValidItems: input.ValidItems, ValidGroups: input.ValidGroups}

	models.DB.Model(&voucher).Updates(&updatedVoucher)
	c.JSON(http.StatusOK, gin.H{"data": voucher})
}

func DeleteVoucher(c *gin.Context) {
	var voucher models.Voucher
	if err := models.DB.Where("voucher_id = ?", c.Param("id")).First(&voucher).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	models.DB.Delete(&voucher)
	c.JSON(http.StatusOK, gin.H{"data": "success"})
}
