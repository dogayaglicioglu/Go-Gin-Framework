package controller

import (
	"fmt"
	"konzek_assg/database"
	"konzek_assg/helper"
	"konzek_assg/model"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func LoginHandler(logger *log.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		Login(context, logger)
	}
}

func RegisterHandler(logger *log.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		Register(context, logger)
	}
}

func Register(context *gin.Context, logger *log.Logger) {
	var input model.AuthenticationInput

	if err := context.ShouldBindJSON(&input); err != nil {
		logger.Println("Error binding JSON: ", err)
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		}
		context.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	tx := database.Database.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Println("Recovered from error: ", r)
		}
	}()

	user := model.User{
		Username: input.Username,
		Password: input.Password,
	}
	savedUser, err := user.SaveInTransaction(tx)
	if err != nil {
		tx.Rollback()
		logger.Println("Error creating user: ", err)
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusBadRequest,
			Message:    err.Error(),
		}
		context.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		logger.Println("Error committing transaction: ", err)
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "internal server error",
		}
		context.JSON(http.StatusInternalServerError, errorResponse)
		return
	}
	successResponse := model.SuccessResponse{
		StatusCode: http.StatusCreated,
		Message:    "User created successfully.",
		Data:       savedUser.Username,
	}
	context.JSON(http.StatusCreated, successResponse)
	logger.Println("User created successfully.")

}

func Login(context *gin.Context, logger *log.Logger) {
	var input model.AuthenticationInput

	if err := context.ShouldBindJSON(&input); err != nil {
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "Error in binding JSON",
		}
		context.JSON(http.StatusBadRequest, errorResponse)
		return
	}

	user, err := model.FindUserByUsername(input.Username)
	if err != nil {
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "User not found",
		}
		context.JSON(http.StatusNotFound, errorResponse)
		return
	}

	err = user.ValidatePassword(input.Password)
	if err != nil {
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "Invalid username or password",
		}
		context.JSON(http.StatusUnauthorized, errorResponse)
		return
	}

	jwt, err := helper.GenerateJWT(user)
	if err != nil {
		errorResponse := model.ErrorResponse{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to generate jwt. ",
		}
		context.JSON(http.StatusInternalServerError, errorResponse)
		return
	}

	logger.Printf("User %s logged in successfully.", input.Username)
	successResponse := model.SuccessResponse{
		StatusCode: http.StatusOK,
		Message:    fmt.Sprintf("User %s logged in successfully.", input.Username),
		Data:       gin.H{"jwt": jwt},
	}
	context.JSON(http.StatusOK, successResponse)

}
