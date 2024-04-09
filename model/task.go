package model

import (
	"errors"
	"konzek_assg/database"
	"log"

	"gorm.io/gorm"
)

type SuccessResponse struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

type ErrorResponse struct {
	StatusCode int    `json:"-"`
	Message    string `json:"error"`
}

type Task struct {
	gorm.Model
	UserID      uint   `gorm:"size:255;not null;" json:"userid"`
	Title       string `gorm:"size:255;not null;" json:"title"`
	Description string `gorm:"size:255;not null;" json:"description"`
	Status      string `gorm:"size:255;not null;" json:"status"`
}

func ReadAllTasksByUserID(userID uint, logger *log.Logger) ([]Task, error) {
	var tasks []Task
	if err := database.Database.Where("user_id = ?", userID).Find(&tasks).Error; err != nil {
		logger.Printf("Error in reading task from database %v", err)
		return nil, errors.New("error in reading task")
	}
	return tasks, nil
}

func (task *Task) SaveInTransaction(tx *gorm.DB, logger *log.Logger) (*Task, error) {
	if task.ID == 0 {
		err := tx.Create(task).Error
		if err != nil {
			logger.Println("Error in creating user", err)
			return nil, errors.New("error in creating user")
		}
	} else {
		err := tx.Save(task).Error
		if err != nil {
			logger.Println("Error in updating user", err)
			return nil, errors.New("error in updating user")
		}
	}
	return task, nil
}
