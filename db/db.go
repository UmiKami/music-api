package db

import (
	"log"
	"os"
	"umikami/go-music/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect()  {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dsn := os.Getenv("DB_URI")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}

	DB = db

	// run migrations

	err = models.MigrateArtist(DB)

	if err != nil {
		log.Fatal("Failed to migrate Artist schema", err)
	}

	err = models.MigrateMusicFile(DB)

	if err != nil {
		log.Fatal("Failed to migrate Music schema", err)
	}

	err = models.MigrateUser(DB)

	if err != nil {
		log.Fatal("Failed to migrate User schema", err)
	}

	log.Println("Database connection and migration successful!")
}