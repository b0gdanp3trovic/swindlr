package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func AddBackend(c *gin.Context) {
	var backend Backend
	if err := c.BindJSON(&backend); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	serverPool.AddBackend(&backend)
	c.JSON(http.StatusOK, gin.H{"message": "Backend added successfully"})
}

func RemoveBackend(c *gin.Context) {
	url := c.Param("url")
	if err := serverPool.RemoveBackend(url); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Backend added successfully"})
}
