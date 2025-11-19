package handlers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

func (lac *LocalApiConfig) HandlerCheckReadiness(c *gin.Context) {
	log.Print("✓ Health Check Readiness")
	c.JSON(http.StatusOK, gin.H{
		"status": "Ok",
	})
}
