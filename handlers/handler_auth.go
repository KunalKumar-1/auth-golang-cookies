package handlers

import (
	"auth-golang-cookies/models"
	"auth-golang-cookies/utils"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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
			"error": "failed to save session data to redis " + err.Error(),
		})
		return
	}

	onlineUserData := map[string]interface{}{
		"username": foundUser.Username,
		"ID":       foundUser.ID,
	}

	onlineUserDataJson, err := json.Marshal(onlineUserData)

	//creating a redis key specifically for tracking logging user in real-time
	onlineKey := "onlineUser:" + sessionId
	err = lac.RedisClient.Set(c, onlineKey, onlineUserDataJson, time.Until(expirationTime)).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to mark user as online " + err.Error(),
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

	// Delete session from redis
	err = lac.RedisClient.Del(c, sessionId).Err()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "failed to end Session from Redis",
		})
		return
	}
	// Clears the session cookie
	c.SetCookie("session_id", "", -1, "/", "", false, true)

	// remove the online userKey form redis to handle conflict and data redundancy.
	onlineKey := "onlineUser:" + sessionId

	err = lac.RedisClient.Del(c, onlineKey).Err()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "failed to end Session from Redis" + err.Error(),
		})
		return
	}

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

func (lac *LocalApiConfig) HandlerFetchOnlineUser(c *gin.Context) {
	// fetch all the keys for the online users
	keys, err := lac.RedisClient.Keys(c, "OnlineUser:*").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch online user keys from redis" + err.Error(),
		})
		return
	}

	if len(keys) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"error":       "no online user found",
			"onlineUsers": nil,
		})
		return
	}

	// using redis pipeline to fetch all data at once
	pipe := lac.RedisClient.Pipeline()
	cmds := make([]*redis.StringCmd, len(keys))
	for i, key := range keys {
		cmds[i] = pipe.Get(c, key)
	}
	_, err = pipe.Exec(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to fetch online user data from redis pipline" + err.Error(),
		})
		return
	}

	//preparing slice to hold user data
	onlineUsers := make([]map[string]interface{}, 0, len(keys))

	// get the data from pipeline to slice
	for _, cmd := range cmds {
		data, err := cmd.Result()
		if err != redis.Nil {
			// key does not exist → skip or append empty
			continue
		}

		var onlineData map[string]interface{}
		err = json.Unmarshal([]byte(data), &onlineData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to fetch Unmarshal the online user data from redis" + err.Error(),
			})
			return
		}
		onlineUsers = append(onlineUsers, onlineData)
	}

	//sending to client
	c.JSON(http.StatusOK, gin.H{
		"message":     "Fetched online users successfully",
		"onlineUsers": onlineUsers,
	})
}
