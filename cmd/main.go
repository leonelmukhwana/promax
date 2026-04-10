package main

import (
	"api/config"
	"api/internal/middleware"
	"api/internal/models"
	"api/internal/routes"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("ℹ️  Note: .env file not found, using system environment")
	}

	// 2. Set Gin Mode (release for speed, debug for development)
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)

	// 3. Connect to Database (This calls the logic in your config/database.go)
	fmt.Println("📡 Initializing database connection...")
	config.ConnectDatabase()

	//cor handling
	// 1. ATTACH CORS FIRST
	rr := gin.Default()

	// 1. ATTACH CORS FIRST
	rr.Use(middleware.SetupCORS())

	//handle migrationsmode// Inside your database initialization or main function
	// Inside your database setup function
	err := config.DB.AutoMigrate(
		&models.User{},
		&models.OTP{},
		&models.NannyProfile{},
		&models.EmployerProfile{},
		&models.Job{},     // Add this!
		&models.Booking{}, // Add this!
	)

	if err != nil {
		log.Fatal("Migration Failed: ", err)
	}

	if err != nil {
		log.Fatal("Migration Failed: ", err)
	}

	// 4. Initialize the Gin engine
	// We use New() instead of Default() to have absolute control over middleware
	r := gin.New()

	// 5. Global Middleware
	r.Use(gin.Recovery()) // Prevents server from crashing on errors
	if mode != gin.ReleaseMode {
		r.Use(gin.Logger()) // Only log requests in debug mode for performance
	}

	// 6. Register All Routes
	// We pass config.DB which was initialized in Step 3
	routes.SetupRoutes(r, config.DB)

	// 7. Simple Health Check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ready",
			"db":     "connected",
		})
	})

	//to handle files and uploads
	r.Static("/uploads", "./uploads")

	// 8. Launch the Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("\n✅ SUCCESS: Nanny System is live at http://localhost:%s\n", port)
	fmt.Println("-------------------------------------------------------")

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("CRITICAL: Could not start server: %v", err)
	}
}
