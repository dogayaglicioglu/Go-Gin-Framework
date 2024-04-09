package database

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var Database *gorm.DB

func loadEnv() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current working directory: %v", err)
	}
	log.Printf("Current working directory: %s", cwd)

	err = godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
}

func Connect() (*gorm.DB, error) {
	loadEnv()
	var err error
	host := os.Getenv("DB_HOST")
	username := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	databaseName := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Africa/Lagos", host, username, password, databaseName, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)

	}
	fmt.Println("Successfully connected to the database.")
	Database = db

	err = Database.AutoMigrate(&User{})
	if err != nil {
		log.Fatalf("failed to auto migrate users table: %v", err)
		return nil, err
	}

	err = Database.AutoMigrate(&Task{})
	if err != nil {
		log.Fatalf("failed to auto migrate tasks table: %v", err)
		return nil, err
	}
	log.Printf("Task table migrated successfully.")

	return db, nil
}
