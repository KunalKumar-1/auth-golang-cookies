package main

import (
	"auth-golang-cookies/handlers"
	"auth-golang-cookies/internal/config"
	"auth-golang-cookies/internal/database"
	"database/sql"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
)

func main() {
	//initialise redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	//initialise the database
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("error loading env file")
	}

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("error loading env DB_URL")
	}

	conn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("could not connect to database")
	}

	// logging colors
	const (
		Red   = "\033[31m"
		Green = "\033[32m"
		Reset = "\033[0m"
	)

	// testing the db connection
	var testQuery int
	err = conn.QueryRow("SELECT 1").Scan(&testQuery)

	if err != nil {
		log.Fatal(Red + "test: database connection test failed !" + Reset)
	} else {
		log.Println(Green + "test: database connection test query executed successfully !" + Reset)
		log.Print(Green + "test: Database connection is working fine !" + Reset)
	}

	// setup API configuration
	apiConfig := &config.ApiConfig{
		DB:          database.New(conn),
		RedisClient: redisClient,
	}

	localApiConfig := &handlers.LocalApiConfig{
		ApiConfig: apiConfig,
	}

	// initialising the router
	router := gin.Default()
	//cors
	router.Use(cors.Default())
	authorized := router.Group("/")
	authorized.Use(localApiConfig.AuthMiddleware())
	{
		authorized.GET("/health-check", localApiConfig.HandlerCheckReadiness)
		authorized.GET("/test-auth", localApiConfig.HandlerAuthRoute)
	}

	router.POST("/sign-in", localApiConfig.SignInHandler)
	router.POST("/logout", localApiConfig.LogOutHandler)
	router.POST("/signup", localApiConfig.HandleCreateUser)

	log.Fatal(router.Run(":8080"))
}
