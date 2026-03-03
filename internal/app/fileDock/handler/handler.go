package handler

import (
	"github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/storage"
	"github.com/gin-gonic/gin"
)

func SignUp(c *gin.Context) {
	storage.SignUp()
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
