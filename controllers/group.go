package controllers

import (
	"metalab/events-inventory-tracker/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreateGroupInput struct {
	Name      string        `json:"name" binding:"required"`
	Items     []models.Item `json:"items" binding:"required"`
	IsVisible bool          `json:"visible" binding:"required"`
}

func CreateGroup(c *gin.Context) {
	var input CreateGroupInput
	containedItemsArray := []models.Item{}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, v := range input.Items {
		item := FindItemById(v.ItemId)
		if item.Name == "No item found" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "itemid " + strconv.FormatUint(uint64(v.ItemId), 10) + " not found"})
			return
		}
		containedItemsArray = append(containedItemsArray, models.Item{ItemId: item.ItemId, Name: item.Name, Price: item.Price})
	}

	group := models.Group{Name: input.Name, Items: containedItemsArray, IsVisible: input.IsVisible}
	models.DB.Create(&group)

	c.JSON(http.StatusOK, gin.H{"data": group})
}

func FindGroups(c *gin.Context) {
	var groups []models.Group
	models.DB.Find(&groups)

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{"data": groups})
}

func FindGroup(c *gin.Context) {
	var group models.Group

	if err := models.DB.Where("group_id = ?", c.Param("id")).First(&group).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, gin.H{"data": group})
}

func UpdateGroup(c *gin.Context) {
	var group models.Group
	if err := models.DB.Where("group_id = ?", c.Param("id")).First(&group).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	var input CreateGroupInput
	containedItemsArray := []models.Item{}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, v := range input.Items {
		item := FindItemById(v.ItemId)
		if item.Name == "No item found" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "itemid " + strconv.FormatUint(uint64(v.ItemId), 10) + " not found"})
		}
		containedItemsArray = append(containedItemsArray, models.Item{ItemId: item.ItemId, Name: item.Name, Price: item.Price})
	}

	updatedGroup := models.Group{Name: input.Name, Items: containedItemsArray, IsVisible: input.IsVisible}

	models.DB.Model(&group).Updates(&updatedGroup)
	c.JSON(http.StatusOK, gin.H{"data": group})
}

func DeleteGroup(c *gin.Context) {
	var group models.Group
	if err := models.DB.Where("group_id = ?", c.Param("id")).First(&group).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "record not found"})
		return
	}

	models.DB.Delete(&group)
	c.JSON(http.StatusOK, gin.H{"data": "success"})
}
