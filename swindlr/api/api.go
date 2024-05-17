package api

import (
	"net/http"
	"net/url"

	"github.com/b0gdanp3trovic/swindlr/loadbalancer"
	"github.com/gin-gonic/gin"
)

func AddBackend(c *gin.Context, serverPool *loadbalancer.ServerPool) {
	var input struct {
		URL string `json:"URL"`
	}

	if err := c.BindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsedUrl, err := url.Parse(input.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL format"})
		return
	}

	serverPool.AddBackend(loadbalancer.CreateNewBackend(parsedUrl, serverPool))
	c.JSON(http.StatusOK, gin.H{"message": "Backend added successfully"})
}

func RemoveBackend(c *gin.Context, serverPool *loadbalancer.ServerPool) {
	url := c.Param("url")
	if err := serverPool.RemoveBackend(url); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Backend removed successfully"})
}
