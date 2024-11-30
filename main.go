package main

import (
	"metalab/events-inventory-tracker/controllers"
	"metalab/events-inventory-tracker/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.Use(cors.Default())

	models.ConnectDatabase()

	router.POST("/api/items", controllers.CreateItem)
	router.GET("/api/items", controllers.FindItems)
	router.GET("/api/items/:id", controllers.FindItem)
	router.PATCH("/api/items/:id", controllers.UpdateItem)
	router.DELETE("/api/items/:id", controllers.DeleteItem)

	router.POST("/api/purchases", controllers.CreatePurchase)
	router.GET("/api/purchases", controllers.FindPurchases)
	router.GET("/api/purchases/:id", controllers.FindPurchase)
	router.PATCH("/api/purchases/:id", controllers.UpdatePurchase)
	router.DELETE("/api/purchases/:id", controllers.DeletePurchase)

	router.Run("localhost:8080")
}
