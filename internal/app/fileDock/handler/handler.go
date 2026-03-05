package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	filedock "github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/fileDock"
	"github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/storage"
	"github.com/Vladyslav-Kondrenko/fileDock/internal/pkg/passwords"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	contentType := file.Header.Get("Content-Type")
	if contentType != "image/png" && contentType != "image/jpeg" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type"})
		return
	}

	f, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()

	const maxSize = 10 << 20 // 10 MB
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
		return
	}

	filename := uuid.New().String()
	ext := strings.Split(contentType, "/")[1]
	filename = filename + "." + ext

	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	savedFile, err := storage.SaveFile(c.Request.Context(), filedock.File{
		UserID:    oid,
		FileName:  filename,
		FileSize:  file.Size,
		FileType:  contentType,
		CreatedAt: time.Now(),
		URL:       storage.DefaultS3Client.ObjectURL(filename),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = storage.DefaultS3Client.Upload(c.Request.Context(), savedFile.ID.Hex(), f, file.Size, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		storage.DeleteFile(c.Request.Context(), savedFile.ID)
		return
	}
	c.JSON(http.StatusOK, gin.H{"file": savedFile})
}

func GetFiles(c *gin.Context) {
	userID := c.GetString("user_id")
	fmt.Println("userID", userID)
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	files, err := storage.GetFiles(c.Request.Context(), oid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"files": files})
}
