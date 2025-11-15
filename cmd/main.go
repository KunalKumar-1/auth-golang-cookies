package main

import (
	"auth-golang-cookies/handlers"
	"auth-golang-cookies/internal/config"
	"auth-golang-cookies/internal/database"
	"database/sql"
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

	var testQuery int
	err = conn.QueryRow("SELECT 1").Scan(&testQuery)
	if err != nil {
		log.Fatal("database connection test failed!!")
	} else {
		log.Println("database connection test query executed successfully!")
		log.Print("Database connection is working fine!!")
	}

	//setup API configuration

	apiConfig := &config.ApiConfig{
		DB:          database.New(conn),
		RedisClient: redisClient,
	}

	localApiConfig := &handlers.LocalApiConfig{
		ApiConfig: apiConfig,
	}

	//initialising the router
	router := gin.Default()

	router.GET("/health-check", localApiConfig.HandlerCheckReadiness)
	log.Fatal(router.Run(":8080"))
}
