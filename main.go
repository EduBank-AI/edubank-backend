package main

import (
	"fmt"
	"log"
	"os"
    "context"

	"github.com/edubank/db"
	"github.com/edubank/handlers"
	"github.com/edubank/middleware"


	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default environment variables")
	}

	ctx := context.Background()

    pool, err := db.InitDB(ctx)
    if err != nil {
        log.Fatal("failed to init db:", err)
    }

    fmt.Println("DB connected ✅", pool)

	// Setup Gin router
	r := setupRouter()

	port := getPort()
	log.Printf("Server running on http://localhost%s", port)
	if err := r.Run(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// setupRouter sets up routes and middleware
func setupRouter() *gin.Engine {
	r := gin.Default()

	// Enable CORS
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Health check route
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Server is running!"})
	})

	// Routes

	// Auth routes
	auth := r.Group("/auth")
	{
		auth.POST("/signup", handlers.SignupHandler)
		auth.POST("/login", handlers.LoginHandler)
	}

	// Protected routes
	api := r.Group("/api", middleware.AuthMiddleware())
	{
		api.POST("/datasets/upload", handlers.UploadDatasetHandler)
    	api.GET("/datasets", handlers.ListDatasetsHandler)
		api.POST("/ai", handlers.AIHandler)
	}

	// auth := r.Group("/", middleware.AuthMiddleware())
	// {
	// 	auth.POST("/load", handlers.FileUploadHandler)
	// 	auth.POST("/ai", handlers.AIHandler)
	// }

	return r
}

// getPort returns the port from env or default
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "6000"
	}
	return fmt.Sprintf(":%s", port)
}
