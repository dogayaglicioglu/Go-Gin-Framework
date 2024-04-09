package main

import (
	"konzek_assg/controller"
	"konzek_assg/database"
	"konzek_assg/middleware"
	"konzek_assg/model"
	WORKER "konzek_assg/worker"
	"log"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"

	"gorm.io/gorm"
)

var db *gorm.DB
var logger *log.Logger

func initLogger() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("Failed to open log file: ", err)
	}

	logger = log.New(logFile, "", log.Ldate|log.Ltime)
	controller.SetLogger(logger)

	WORKER.SetLogger(logger)
}

func main() {
	initLogger()
	loadDatabase()
	prometheus.MustRegister(controller.DurationOfRequest)
	serveApplication()
}

func loadDatabase() {
	db, _ = database.Connect()
}

func serveApplication() {
	const numWorkers = 5
	taskCh := make(chan model.Task)
	var wg sync.WaitGroup
	controller.InitializeController()
	controller.SetTaskChannel(taskCh)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go WORKER.Work(taskCh, &wg)
	}

	router := gin.Default()

	publicRoutes := router.Group("/auth")

	// @Summary Register User
	// @Description Create a new user with username and password.
	// @Accept json
	// @Produce json
	// @Param input body AuthenticationInput true "User credentials"
	// @Success 201 {object} SuccessResponse "User created successfully."
	// @Router /auth/register [post]
	publicRoutes.POST("/register", controller.RegisterHandler(logger))
	// LoginHandler handles user login.
	// @Summary Log User In
	// @Description Log in a user with provided credentials.
	// @Accept json
	// @Produce json
	// @Param input body AuthenticationInput true "User credentials"
	// @Success 200 {object} SuccessResponse "User logged in successfully."
	// @Header 200 {string} Token "Bearer" "Authentication token"
	// @Router /auth/login [post]
	publicRoutes.POST("/login", controller.LoginHandler(logger))

	protectedRoutes := router.Group("/api")
	protectedRoutes.Use(middleware.JWTAuthMiddleware())

	// CreateTaskHandler handles the creation of a new task.
	// @Summary Create Task
	// @Description Create a new task.
	// @Security ApiKeyAuth
	// @Accept json
	// @Produce json
	// @Param Authorization header string true "Bearer token"
	// @Param task body Task true "Task object to create"
	// @Success 200 {object} SuccessResponse "Task created successfully."
	// @Failure 400 {object} ErrorResponse "Bad request"
	// @Router /api/entry [post]
	protectedRoutes.POST("/entry", controller.CreateTaskHandler)

	// GetTasksHandler handles fetching tasks for the authenticated user.
	// @Summary Get User Tasks
	// @Description Fetch tasks associated with the current authenticated user.
	// @Security ApiKeyAuth
	// @Accept json
	// @Produce json
	// @Param Authorization header string true "Bearer token"
	// @Success 200 {object} SuccessResponse "Tasks queried successfully."
	// @Failure 400 {object} ErrorResponse "Bad request"
	// @Router /api/entry [get]
	protectedRoutes.GET("/entry", controller.GetTasksHandler)
	// DeleteTaskHandler handles the deletion of a task by ID.
	// @Summary Delete Task
	// @Description Delete a task by its ID.
	// @Security ApiKeyAuth
	// @Accept json
	// @Produce json
	// @Param Authorization header string true "Bearer token"
	// @Param taskid path int true "Task ID to delete"
	// @Success 200 {object} SuccessResponse "Task deletion request sent successfully."
	// @Failure 400 {object} ErrorResponse "Bad request"
	// @Router /api/entry/{taskid} [delete]
	protectedRoutes.DELETE("/entry/:taskid", controller.DeleteTaskHandler)
	// UpdateTaskHandler handles updating a task by ID.
	// @Summary Update Task
	// @Description Update a task by its ID.
	// @Security ApiKeyAuth
	// @Accept json
	// @Produce json
	// @Param Authorization header string true "Bearer token"
	// @Param taskid path int true "Task ID to update"
	// @Param task body Task true "Task object containing updated data"
	// @Success 200 {object} SuccessResponse "Task update request sent successfully."
	// @Failure 400 {object} ErrorResponse "Bad request"
	// @Router /api/entry/{taskid} [put]
	protectedRoutes.PUT("/entry/:taskid", controller.UpdateTaskHandler)

	router.Run(":8000")
}
