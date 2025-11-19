package handlers

import (
	"auth-golang-cookies/models"
	"auth-golang-cookies/utils"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"os"
	"time"
)

type Claims struct {
	Email  string    `json:"email"`
	UserId uuid.UUID `json:"userId"`
	jwt.RegisteredClaims
}

type JWTOutput struct {
	Token  string    `json:"token"`
	Expire time.Time `json:"expires"`
}

type SessionData struct {
	Token  string    `json:"token"`
	UserId uuid.UUID `json:"userId"`
}

func (lac *LocalApiConfig) SignInHandler(c *gin.Context) {
	var userToAuth models.UserToAuth

	if err := c.ShouldBindJSON(&userToAuth); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	//insert validation here
	validationError := utils.ValidateUserToAuth(userToAuth)
	if len(validationError) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validationError,
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
		"message": "Logged in successfully",
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

func (lac *LocalApiConfig) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionId, err := c.Cookie("session_id")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized - no session",
			})
			return
		}
		sessionDataJSON, err := lac.RedisClient.Get(c, sessionId).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired Session from Redis",
			})
			return
		}

		var sessionData SessionData
		err = json.Unmarshal([]byte(sessionDataJSON), &sessionData)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Failed to decode/unmarshal session data",
			})
			return
		}

		token, err := jwt.ParseWithClaims(sessionData.Token, &Claims{},
			func(token *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("JWT_SECRET")), nil
			})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			return
		}
		c.Set("userId", sessionData.UserId)
		c.Next()
	}
}

func (lac *LocalApiConfig) HandlerAuthRoute(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Authenticated routes are working successfully",
	})
	return
}
