package handler

import (
	"errors"
	"fmt"
	"net/http"

	filedock "github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/fileDock"
	"github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/storage"
	"github.com/Vladyslav-Kondrenko/fileDock/internal/pkg/passwords"
	"github.com/gin-gonic/gin"
)

func SignUp(c *gin.Context) {
	var credentials filedock.UserCredentials

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash, err := passwords.HashPassword(credentials.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	credentials.Password = hash

	err = storage.SignUp(c.Request.Context(), credentials)
	if err != nil {
		if errors.Is(err, storage.ErrEmailExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"user": "user created successfully"})
}

func SignIn(c *gin.Context) {
	var credentials filedock.UserCredentials

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := storage.SignIn(c.Request.Context(), credentials)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})

}

func UploadFile(c *gin.Context) {

	storage.UploadFile()
}

func GetFiles(c *gin.Context) {
	userID := c.GetString("user_id")
	fmt.Println("userID", userID)
	storage.GetFiles()
}
