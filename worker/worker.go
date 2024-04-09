package worker

import (
	"errors"
	"fmt"
	"konzek_assg/database"
	"konzek_assg/model"

	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

var (
	logger         *log.Logger
	mutex          sync.Mutex
	tasks          = make(map[uint]model.Task)
	taskResultLock sync.Mutex
	taskResults    = make(map[uint][]model.Task)
)

func SetTasks(t map[uint]model.Task) {
	tasks = t
}
func SetTaskResults(tr map[uint][]model.Task) {
	taskResultLock.Lock()
	defer taskResultLock.Unlock()
	taskResults = tr
}

func SetLogger(l *log.Logger) {
	logger = l
}

func Work(taskCh <-chan model.Task, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range taskCh {
		switch task.Status {
		case "create":
			err := CreateTask(task, logger)

			if err != nil {
				logger.Printf("Failed to create task for user ID %d: %v\n", task.UserID, err)
			}
		case "read":
			tasks, err := ReadTask(task.UserID, logger)

			if err != nil {
				logger.Printf("Failed to read tasks for user ID %d: %v\n", task.UserID, err)
			} else {
				taskResultLock.Lock()
				taskResults[task.UserID] = tasks
				taskResultLock.Unlock()
			}
		case "delete":
			err := DeleteTask(task, logger)
			if err != nil {
				logger.Printf("Deleting tasks for user ID: %d\n", task.UserID)
			}
		case "update":
			err := UpdateTask(task, logger)
			if err != nil {
				logger.Println("Invalid operation for task")
			}
		default:
			logger.Println("Invalid operation for task.")
		}
	}
}

func ReadTask(userId uint, logger *log.Logger) ([]model.Task, error) {
	if userId == 0 {
		logger.Println("User id is empty")
		return nil, errors.New("the user id is empty")
	}

	tasks, err := model.ReadAllTasksByUserID(userId, logger)
	if err != nil {
		logger.Println("Error in reading user's tasks:", err)
		return nil, fmt.Errorf("failed to read tasks for user %d: %w", userId, err)
	}

	if len(tasks) == 0 {
		logger.Println("No tasks found for user:", userId)
		return nil, errors.New("there are no tasks for the given user id")
	}

	logger.Printf("Tasks read successfully: %v\n", tasks)

	return tasks, nil
}

func UpdateTask(task model.Task, logger *log.Logger) error {
	if task.ID == 0 {
		logger.Println("Task ID is required for updating.")
		return errors.New("task ID is required for updating")
	}

	if task.UserID == 0 {
		logger.Println("User ID is required for updating.")
		return errors.New("user ID is required for task update")
	}

	var taskFromDB model.Task
	if err := database.Database.First(&taskFromDB, task.ID).Error; err != nil {
		logger.Println("The given task is not founded in the database.")
		return errors.New("the given task is not founded in the database")
	}

	if taskFromDB.UserID != task.UserID {
		logger.Println("You are not authorized to update this task.")
		return errors.New("you are not authorized to update this task")
	}

	tx := database.Database.Begin()
	defer tx.Rollback()

	if err := tx.Model(&model.Task{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"title":       task.Title,
		"description": task.Description,
		"status":      task.Status,
	}).Error; err != nil {
		logger.Println("Error in updating task.")
		return errors.New("error in updating task")
	}

	if err := tx.Commit().Error; err != nil {
		logger.Println("Error in committing task.")
		return errors.New("error in committing")
	}
	logger.Println("Task is updated successfully.")
	return nil
}

func DeleteTask(task model.Task, logger *log.Logger) error {
	userid := task.UserID

	if task.ID == 0 {
		logger.Println("Task id is required for deletion.")
		return errors.New("task ID is required for deletion")
	}
	var taskFromDB model.Task
	if err := database.Database.First(&taskFromDB, task.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Println("Task is not found.")
			return errors.New("task not found")
		}
		return err
	}

	if taskFromDB.UserID != userid {
		logger.Println("You are not authorized to delete this task.")
		return errors.New("you are not authorized to delete this task")
	}

	var deletedTime gorm.DeletedAt
	deletedTime.Time = time.Now()
	taskFromDB.DeletedAt = deletedTime

	tx := database.Database.Begin()

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Println("recovered from error:", r)
		}
	}()

	if err := tx.Delete(&taskFromDB).Error; err != nil {
		tx.Rollback()
		logger.Println("Error in deleting task:", err)
		return errors.New("error in deleting task")
	}
	if err := tx.Commit().Error; err != nil {
		logger.Println("Error in committing delete task:", err)
		return errors.New("error in committing delete task")
	}
	logger.Println("Task is deleted successfully.")
	return nil
}

func CreateTask(task model.Task, logger *log.Logger) error {
	if task.Title == "" {
		logger.Println("Task title is empty.")
		return errors.New("the title field shouldn't be empty")
	}
	if task.Status == "" {
		logger.Println("Task status is empty.")
		return errors.New("the status field shouldn't be empty")
	}

	tx := database.Database.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Println("Recovered from error:", r)
		}
	}()

	mutex.Lock()
	defer mutex.Unlock()

	task.CreatedAt = time.Now()

	if _, err := task.SaveInTransaction(tx, logger); err != nil {
		tx.Rollback()
		logger.Println("Failed to save task.")
		return errors.New("failed to save task")
	}

	tasks[task.ID] = task

	if err := tx.Commit().Error; err != nil {
		logger.Println("Failed to commit transaction:", err)
		return errors.New("failed to commit transaction")
	}
	logger.Println("Task created successfully.")
	return nil
}
