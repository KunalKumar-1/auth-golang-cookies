package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (lac *LocalApiConfig) HandlerCheckReadiness(c *gin.Context) {
	log.Print("✓ Health Check Readiness")
	c.JSON(http.StatusOK, gin.H{
		"status": "Ok",
	})
}
