package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func (lac *LocalApiConfig) HandlerPusherAuth(c *gin.Context) {

	// extracting the token from the Authorization header
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "No Authorization header found",
		})
		return
	}

	//strip `bearer` prefix if present
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	}

	// parse the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "failed parsing token",
		})
		return
	}

	//check if the token was valid
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		_ = claims.UserId //Assuming Claims struct includes UserID

		// extracting params for pusher authorization
		params, _ := io.ReadAll(c.Request.Body)
		response, err := lac.PusherClient.AuthenticatePrivateChannel(params)
		if err != nil {
			fmt.Println(err.Error())
			c.JSON(http.StatusOK, gin.H{
				"error": "authentication with pusher failed",
				"err":   err.Error(),
			})
			return
		}
		// Return the auth response to the client
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Writer.Write(response)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "invalid token",
		})
		return
	}
}

func (lac *LocalApiConfig) HandlerPusherWS(c *gin.Context) {
	type NewLogin struct {
		UserId   string `json:"userId"`
		UserName string `json:"username"`
	}

	//new login struct to handle login struct data
	newLogin := &NewLogin{
		UserId:   "kunal123",
		UserName: "kunal kumar",
	}

	err := lac.PusherClient.Trigger("my-channel", "my-event", newLogin)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "failed triggering new login",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data sent to real time pusher service for the client",
	})
}

func (lac *LocalApiConfig) HandlerSendMessage(c *gin.Context) {
	type NewMessage struct {
		Message  string `json:"message"`
		UserName string `json:"username"`
	}

	newMessage := &NewMessage{}
	if err := c.ShouldBindJSON(&newMessage); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
		})
		return
	}

	userIdInterface, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "userId failed to load in gin context",
		})
		return
	}

	userId, ok := userIdInterface.(uuid.UUID) // Type assertion
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed at type assertion",
		})
		return
	}

	channelName := fmt.Sprintf("private-%s", userId.String())

	err := lac.PusherClient.Trigger(channelName, "new-message", newMessage)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error in event triggering from the pusher" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Data sent to real time pusher service for the client",
		"userId":      userId.String(),
		"text":        newMessage,
		"channelName": channelName,
	})
}

func (lac *LocalApiConfig) HandlerNotifySubscribe(c *gin.Context) {
	type UserToNotify struct {
		UserId uuid.UUID `json:"userId"`
	}

	var user UserToNotify

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error to bind the userId" + err.Error()})
		return
	}

	// finding user from the database
	_user, err := lac.DB.FindUserById(c, user.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "NO USER FOUND ERROR",
		})
		return
	}

	type NewLogin struct {
		Username string `json:"username"`
		UserId   string `json:"userId"`
	}

	// setup a new login struct
	newLogin := &NewLogin{
		Username: _user.Name,
		UserId:   _user.ID.String(),
	}

	err = lac.PusherClient.Trigger("myuser-login", "new-login", newLogin)
	if err != nil {
		fmt.Println(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to trigger the pusher event" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data send to client in real time",
	})
}
