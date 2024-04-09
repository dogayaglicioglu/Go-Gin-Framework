package database

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"size:255;not null;unique" json:"username"`
	Password string `gorm:"size:255;not null;" json:"password"`
	Tasks    []Task `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"` // Define the association
}

type Task struct {
	gorm.Model
	UserID      uint   `gorm:"size:255;not null;" json:"userid"`
	Title       string `gorm:"size:255;not null;" json:"title"`
	Description string `gorm:"size:255;not null;" json:"description"`
	Status      string `gorm:"size:255;not null;" json:"status"`
}

var MockDatabase *gorm.DB

func LoadMockDatabase() {
	var err error
	host := "172.17.0.2"
	username := "mock_user"
	password := "mock_password"
	databaseName := "mock_database"
	port := "5433"

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Africa/Lagos", host, username, password, databaseName, port)

	MockDatabase, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("failed to connect to mock database: %v", err)
	}

	fmt.Println("Successfully connected to the mock database.")

	MockDatabase.AutoMigrate(&User{})
	MockDatabase.AutoMigrate(&Task{})

}
