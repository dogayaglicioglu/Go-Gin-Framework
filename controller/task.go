package controller

import (
	"konzek_assg/helper"
	"konzek_assg/model"
	"konzek_assg/worker"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/gorm"
)

var (
	logger         *log.Logger
	taskCh         = make(chan model.Task)
	taskResultLock sync.Mutex
	taskResults    = make(map[uint][]model.Task)
)

var (
	DurationOfRequest = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "konzek_assignment_http_request_duration_seconds",
			Help: "Histogram of the duration of HTTP requests.",
		},
		[]string{"konzek_handler", "status"},
	)
)

func InitializeController() {

	taskResults = make(map[uint][]model.Task)

	worker.SetTaskResults(taskResults)
}

func SetTaskChannel(tc chan model.Task) {
	taskCh = tc
}

func observeRequestDuration(c *gin.Context, start time.Time) {
	duration := time.Since(start).Seconds()
	handler := c.Request.Method + " " + c.FullPath()
	status := strconv.Itoa(c.Writer.Status())
	DurationOfRequest.WithLabelValues(handler, status).Observe(duration)
}

func SetLogger(l *log.Logger) {
	logger = l
}

func CreateTaskHandler(c *gin.Context) {
	start := time.Now()
	var task model.Task
	task.Status = "create"
	if err := c.BindJSON(&task); err != nil {
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		}
		logger.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}
	user, err := helper.CurrentUser(c)
	if err != nil {
		logger.Println("Error getting current user:", err)
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}
	task.UserID = user.ID
	taskCh <- task

	successResponse := model.SuccessResponse{
		StatusCode: http.StatusOK,
		Message:    "Task received successfully",
	}
	c.JSON(http.StatusOK, successResponse)
	observeRequestDuration(c, start)
}

func DeleteTaskHandler(c *gin.Context) {
	start := time.Now()
	user, err := helper.CurrentUser(c)
	if err != nil {
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	taskIDStr := c.Param("taskid")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid task ID",
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	task := model.Task{
		UserID: user.ID,
		Model:  gorm.Model{ID: uint(taskID)},
		Status: "delete",
	}

	taskCh <- task
	successResponse := model.SuccessResponse{
		StatusCode: http.StatusOK,
		Message:    "Task deletion request sent to worker pool",
	}
	c.JSON(http.StatusOK, successResponse)

	observeRequestDuration(c, start)
}

func UpdateTaskHandler(c *gin.Context) {
	start := time.Now()
	user, err := helper.CurrentUser(c)
	if err != nil {
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}
	taskIDStr := c.Param("taskid")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 64)
	if err != nil {
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid task ID",
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}
	var task model.Task
	if err := c.BindJSON(&task); err != nil {
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}
	task.ID = uint(taskID)
	task.UserID = user.ID

	task.Status = "update"
	taskCh <- task

	successResponse := model.SuccessResponse{
		StatusCode: http.StatusOK,
		Message:    "Task update request sent to worker pool",
	}
	c.JSON(http.StatusOK, successResponse)

	observeRequestDuration(c, start)
}
func waitForTasks(userID uint) {
	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-time.After(100 * time.Millisecond):
			taskResultLock.Lock()
			tasks := taskResults[userID]
			taskResultLock.Unlock()
			if len(tasks) > 0 {
				return
			}
		case <-timeout:
			return
		}
	}
}

func GetTasksHandler(c *gin.Context) {
	start := time.Now()
	user, err := helper.CurrentUser(c)
	if err != nil {
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		}
		c.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	task := model.Task{
		UserID: user.ID,
		Status: "read",
	}

	taskCh <- task

	waitForTasks(task.UserID)

	taskResultLock.Lock()
	defer taskResultLock.Unlock()
	tasks := taskResults[task.UserID]
	if len(tasks) == 0 {
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    "No tasks found for the given user ID",
		}
		c.JSON(http.StatusNotFound, errorResponse)
		return
	}

	successResponse := model.SuccessResponse{
		StatusCode: http.StatusOK,
		Message:    "Tasks queried successfully",
		Data:       tasks,
	}
	c.JSON(http.StatusOK, successResponse)

	observeRequestDuration(c, start)
}
