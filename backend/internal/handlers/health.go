package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck returns 200 OK when the server is running.
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}