package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"example.com/Product_RoadMap/models"
	route "example.com/Product_RoadMap/route"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

func initConfig() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}
}

func main() {
	// Load the secret key from the environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set in the environment")
	}

	fmt.Println("JWT Secret Key Loaded Successfully")

	// Initialize the database
	models.Init()
	initConfig()
	defer models.CloseDB()

	// Create a new Gin router
	router := gin.Default()

	// Register routes
	route.RegisterRoutes(router)

	// Get the port from the environment (default to 8080)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create an HTTP server
	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	// Start the server in a goroutine
	go func() {
		fmt.Printf("Server is running on port %s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	fmt.Println("\nShutting down server...")

	// Create a context with a timeout for the shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	fmt.Println("Server exited gracefully")
}
