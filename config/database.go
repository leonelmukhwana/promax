package config

import (
	"api/internal/models"
	"fmt"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	dsn := os.Getenv("DB_URL")
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-Migrate all models (This creates/updates tables automatically)
	err = database.AutoMigrate(
		&models.User{},
		&models.OTP{}, // We will define this below
		&models.AuditLog{},
	)

	if err != nil {
		log.Fatal("Migration Failed:", err)
	}

	fmt.Println("Database Connected & Migrated Successfully")
	DB = database
}
