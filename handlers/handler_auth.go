package handlers

import (
	"auth-golang-cookies/models"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"os"
	"time"
)

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

type JWTOutput struct {
	Token  string    `json:"string"`
	Expire time.Time `json:"expires"`
}

type SessionData struct {
	Token  string    `json:"string"`
	UserId uuid.UUID `json:"user_id"`
}

func (lac *LocalApiConfig) SignInHandler(c *gin.Context) {
	var userToAuth models.UserToAuth

	if err := c.ShouldBindJSON(&userToAuth); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	//fetch the users from database to check if user exists on not
	foundUser, err := lac.DB.FindUserByEmail(c, userToAuth.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// TODO: In production → NEVER compare plain strings: use bcrypt.CompareHashAndPassword
	if foundUser.Password != userToAuth.Password {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Incorrect Password",
		})
	}

	// Create JWT claims
	expirationTime := time.Now().Add(10 * time.Minute)
	claims := &Claims{
		Email: userToAuth.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create a signed JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Generating session ID (separate from JWT)
	sessionId := uuid.New().String()
	sessionData := map[string]interface{}{
		"token":  tokenString,
		"userId": foundUser.ID,
	}

	// Converting the session map to JSON for storage
	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to marshal session data into sessionDataJSON",
		})
		return
	}

	// Save session to Redis with expiration matching JWT expiry
	err = lac.RedisClient.Set(c, sessionId, sessionDataJSON, time.Until(expirationTime)).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Setting httpCookie
	c.SetCookie("session_id", sessionId, int(time.Until(expirationTime)), "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"message": "Session created successfully",
		"expires": expirationTime,
	})
}

func (lac *LocalApiConfig) LogOutHandler(c *gin.Context) {
	//Retrieve the session from the cookie
	sessionId, err := c.Cookie("session_id")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized request",
		})
		return
	}

	// delete session from redis
	err = lac.RedisClient.Del(c, sessionId).Err()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "failed to end Session from Redis",
		})
		return
	}
	// Clears the session cookie
	c.SetCookie("session_id", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"message": "Redis Session Removed successfully",
	})
}
