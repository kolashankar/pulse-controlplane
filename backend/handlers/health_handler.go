package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheck returns the health status of the service
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "pulse-control-plane",
		"version": "1.0.0",
	})
}

// StatusCheck returns the operational status
func StatusCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "operational",
		"message": "Pulse Control Plane is running",
	})
}
