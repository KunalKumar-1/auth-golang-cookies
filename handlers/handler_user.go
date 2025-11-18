package handlers

import (
	"auth-golang-cookies/internal/config"
	"auth-golang-cookies/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
)

type LocalApiConfig struct {
	*config.ApiConfig
}

func (lac LocalApiConfig) HandleCreateUser(c *gin.Context) {
	type createUserParameters struct {
		Name     string `json:"name"`
		Username string `jason:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	user := createUserParameters{}

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	newUser, err := lac.DB.CreateUser(c, database.CreateUserParams{
		ID:        uuid.New(),
		Name:      user.Name,
		Username:  user.Username,
		Password:  user.Password,
		Email:     user.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusOK, newUser)
}
