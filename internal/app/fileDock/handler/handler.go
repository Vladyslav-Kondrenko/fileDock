package handler

import (
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

	storage.SignUp(c.Request.Context(), credentials)
}

func SignIn(c *gin.Context) {
	storage.SignIn()
}

func UploadFile(c *gin.Context) {
	storage.UploadFile()
}

func GetFiles(c *gin.Context) {
	storage.GetFiles()
}
