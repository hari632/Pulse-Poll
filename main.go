package main

import (
	"log"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"pulse-backend/database"
	"pulse-backend/middleware"
	"pulse-backend/routes"
	"pulse-backend/utils"
)

func main() {
	// ==========================================
	// LOAD ENVIRONMENT VARIABLES
	// ==========================================
	// Load .env file from the backend folder.
	_ = utils.LoadEnv(".env")

	// ==========================================
	// MONGODB CONFIGURATION
	// ==========================================

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
    log.Println("MONGODB_URI: NOT SET")
} else if strings.HasPrefix(mongoURI, "mongodb+srv://") {
    log.Println("MONGODB_URI: Atlas URI detected")
} else {
    log.Println("MONGODB_URI: value detected, but not Atlas")
}

	// Fallback to local MongoDB only if
	// MONGODB_URI is not provided.
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDatabaseName := os.Getenv("MONGODB_DATABASE")

	// Default database name
	if mongoDatabaseName == "" {
		mongoDatabaseName = "pulsepoll"
	}

	// Safely show which MongoDB type is being used.
	// This does NOT print username/password.
	if strings.HasPrefix(mongoURI, "mongodb+srv://") {
		log.Println("MongoDB target: MongoDB Atlas")
	} else if strings.HasPrefix(mongoURI, "mongodb://") {
		log.Println("MongoDB target: MongoDB connection")
	}

	log.Printf("MongoDB database: %s", mongoDatabaseName)

	// Connect to MongoDB
	mongoDB, err := database.ConnectMongoDB(
		mongoURI,
		mongoDatabaseName,
	)

	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}

	log.Println("MongoDB connected successfully")

	// ==========================================
	// REDIS CONFIGURATION
	// ==========================================

	// First try REDIS_URL.
	redisTarget := os.Getenv("REDIS_URL")

	// If REDIS_URL is not available,
	// use REDIS_ADDR.
	if redisTarget == "" {
		redisTarget = os.Getenv("REDIS_ADDR")
	}

	// Final fallback to local Redis.
	if redisTarget == "" {
		redisTarget = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	// Connect to Redis
	redisDB, err := database.ConnectRedis(
		redisTarget,
		redisPassword,
	)

	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected successfully")

	// ==========================================
	// GIN ROUTER
	// ==========================================

	router := gin.Default()

	// Enable CORS for frontend
	router.Use(middleware.CORSMiddleware())

	// ==========================================
	// REGISTER API ROUTES
	// ==========================================

	routes.RegisterRoutes(
		router,
		mongoDB.DB,
		redisDB.Client,
	)

	// ==========================================
	// SERVER PORT
	// ==========================================

	port := os.Getenv("PORT")

	// Default backend port
	if port == "" {
		port = "8080"
	}

	log.Printf(
		"PulsePoll backend starting on http://localhost:%s\n",
		port,
	)

	// Start server
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}