package controller_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"konzek_assg/controller"
	"konzek_assg/database"
	"konzek_assg/helper"
	"konzek_assg/model"
	"konzek_assg/worker"

	"github.com/gin-gonic/gin"
)

func TestRegisterHandler(t *testing.T) {
	database.Connect()
	var buf bytes.Buffer
	logger := log.New(&buf, "", log.Ldate|log.Ltime)

	username := "testuser22"
	defer database.Database.Exec("DELETE FROM users WHERE username = $1", username)

	requestBody := []byte(`{"username": "` + username + `", "password": "testpassword2"}`)
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(rr)
	c.Request = req

	controller.RegisterHandler(logger)(c)

	if rr.Code != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusCreated)
	}

	expectedBody := model.SuccessResponse{
		StatusCode: http.StatusCreated,
		Message:    "User created successfully.",
		Data:       username,
	}

	var actualBody model.SuccessResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &actualBody); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expectedBody, actualBody) {
		t.Errorf("handler returned unexpected body: got %+v want %+v", actualBody, expectedBody)
	}

	if buf.Len() == 0 {
		t.Errorf("log message is not written to the buffer")
	}
	logger.Println("User is registered successfully.")

}

func TestLoginHandler(t *testing.T) {
	database.Connect()
	var buf bytes.Buffer
	logger := log.New(&buf, "", log.Ldate|log.Ltime)

	user := model.User{
		Username: "testuser345",
		Password: "testpassword2",
	}
	if err := database.Database.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	defer database.Database.Delete(&user)
	requestBody := []byte(`{"username": "` + user.Username + `", "password": "testpassword2"}`)

	req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(rr)
	c.Request = req

	controller.LoginHandler(logger)(c)
	if rr.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
	}
	jwtT, err := helper.GenerateJWT(user)
	if err != nil {
		t.Errorf("Error in generating jwt.")
	}
	expectedBody := model.SuccessResponse{
		StatusCode: http.StatusOK,
		Message:    "User " + user.Username + " logged in successfully.",
		Data: map[string]interface{}{
			"jwt": jwtT,
		},
	}

	var actualBody model.SuccessResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &actualBody); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expectedBody, actualBody) {
		t.Errorf("handler returned unexpected body: got %+v want %+v", actualBody, expectedBody)
	}

	if buf.Len() == 0 {
		t.Errorf("log message is not written to the buffer")
	}
	if err := database.Database.Delete(&user).Error; err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}
	logger.Println("User logged in successfully.")

}

func TestCreateTask(t *testing.T) {
	database.Connect()

	var buf bytes.Buffer
	logger := log.New(&buf, "", log.Ldate|log.Ltime)

	task := model.Task{
		UserID:      1,
		Title:       "Test title",
		Description: "Test description",
		Status:      "Create",
	}
	err := worker.CreateTask(task, logger)
	if err != nil {
		t.Errorf("Error in creating task: %v", err)
	}
	logger.Println("Task is created successfully.")
}

func TestDeleteTask(t *testing.T) {
	database.Connect()

	var buf bytes.Buffer
	logger := log.New(&buf, "", log.Ldate|log.Ltime)

	task := model.Task{
		UserID:      1,
		Title:       "Test title",
		Description: "Test description",
		Status:      "Create",
	}

	if err := database.Database.Create(&task).Error; err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
	defer database.Database.Delete(&task)
	err := worker.DeleteTask(task, logger)
	if err != nil {
		t.Errorf("Error in deleting task: %v", err)
	}
	logger.Println("Task is deleted successfully.")

}

func TestUpdateTask(t *testing.T) {
	database.Connect()

	var buf bytes.Buffer
	logger := log.New(&buf, "", log.Ldate|log.Ltime)

	task := model.Task{
		UserID:      1,
		Title:       "Test title",
		Description: "Test description",
		Status:      "Create",
	}

	if err := database.Database.Create(&task).Error; err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
	defer database.Database.Delete(&task)

	taskID := task.ID

	updatedTask := model.Task{
		UserID:      1,
		Title:       "Test mock title",
		Description: "Test mock description",
		Status:      "Update",
	}
	updatedTask.ID = taskID
	err := worker.UpdateTask(updatedTask, logger)
	if err != nil {
		t.Errorf("Error in updating task: %v", err)
	}
	logger.Println("Task is updated successfully.")

}

func TestGetTask(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", log.Ldate|log.Ltime)
	database.Connect()

	baseUsername := "mockuser"
	uniqueUsername := fmt.Sprintf("%s%d", baseUsername, time.Now().UnixNano())

	user := model.User{
		Username: uniqueUsername,
		Password: "mockpassw12deneme12344",
	}

	if err := database.Database.Create(&user).Error; err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	defer database.Database.Delete(&user)
	task1 := model.Task{
		UserID:      user.ID,
		Title:       "Test title",
		Description: "Test description",
		Status:      "Read",
	}

	if err := database.Database.Create(&task1).Error; err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
	defer database.Database.Delete(&task1)
	task2 := model.Task{
		UserID:      user.ID,
		Title:       "Test title1",
		Description: "Test description1",
		Status:      "Read",
	}

	if err := database.Database.Create(&task2).Error; err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
	defer database.Database.Delete(&task2)
	readedTasks, err := worker.ReadTask(user.ID, logger)
	if err != nil {
		t.Fatalf("Failed to read tasks for user %d: %v", user.ID, err)
	}

	logger.Printf("Tasks are readed %v.", readedTasks)
}
