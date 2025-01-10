package controllers

import (
	"github.com/gin-gonic/gin"
)

func ValidateToken(c *gin.Context) {
	c.AbortWithStatus(204)
}
