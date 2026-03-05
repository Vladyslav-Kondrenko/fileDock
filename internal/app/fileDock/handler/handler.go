package handler

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/model"
	"github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/storage"
	"github.com/Vladyslav-Kondrenko/fileDock/internal/pkg/passwords"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StorageService interface {
	SignUp(ctx *gin.Context, email, password string) error
	SignIn(ctx *gin.Context, email, password string) (string, error)
	SaveFile(ctx *gin.Context, file model.File) (model.File, error)
	DeleteFile(ctx *gin.Context, id primitive.ObjectID) error
	GetFiles(ctx *gin.Context, userID primitive.ObjectID) ([]model.File, error)
}

type PasswordService interface {
	Hash(password string) (string, error)
}

type ObjectStorage interface {
	Upload(ctx *gin.Context, key string, body io.Reader, contentLength int64, contentType string) error
	ObjectURL(key string) string
}

type Handler struct {
	storage StorageService
	password PasswordService
	s3       ObjectStorage
	now      func() time.Time
	newUUID  func() string
}

type storageAdapter struct{}

func (storageAdapter) SignUp(c *gin.Context, email, password string) error {
	return storage.SignUp(c.Request.Context(), email, password)
}

func (storageAdapter) SignIn(c *gin.Context, email, password string) (string, error) {
	return storage.SignIn(c.Request.Context(), email, password)
}

func (storageAdapter) SaveFile(c *gin.Context, file model.File) (model.File, error) {
	return storage.SaveFile(c.Request.Context(), file)
}

func (storageAdapter) DeleteFile(c *gin.Context, id primitive.ObjectID) error {
	return storage.DeleteFile(c.Request.Context(), id)
}

func (storageAdapter) GetFiles(c *gin.Context, userID primitive.ObjectID) ([]model.File, error) {
	return storage.GetFiles(c.Request.Context(), userID)
}

type passwordAdapter struct{}

func (passwordAdapter) Hash(password string) (string, error) {
	return passwords.HashPassword(password)
}

type s3Adapter struct{}

func (s3Adapter) Upload(c *gin.Context, key string, body io.Reader, contentLength int64, contentType string) error {
	if storage.DefaultS3Client == nil {
		return errors.New("s3 client is not initialized")
	}
	return storage.DefaultS3Client.Upload(c.Request.Context(), key, body, contentLength, contentType)
}

func (s3Adapter) ObjectURL(key string) string {
	if storage.DefaultS3Client == nil {
		return ""
	}
	return storage.DefaultS3Client.ObjectURL(key)
}

func New(storage StorageService, password PasswordService, s3 ObjectStorage) *Handler {
	return &Handler{
		storage:  storage,
		password: password,
		s3:       s3,
		now:      time.Now,
		newUUID: func() string {
			return uuid.New().String()
		},
	}
}

func NewDefault() *Handler {
	return New(storageAdapter{}, passwordAdapter{}, s3Adapter{})
}

var defaultHandler = NewDefault()

func SignUp(c *gin.Context) {
	defaultHandler.SignUp(c)
}

func SignIn(c *gin.Context) {
	defaultHandler.SignIn(c)
}

func UploadFile(c *gin.Context) {
	defaultHandler.UploadFile(c)
}

func GetFiles(c *gin.Context) {
	defaultHandler.GetFiles(c)
}

func (h *Handler) SignUp(c *gin.Context) {
	var credentials AuthCredentials

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash, err := h.password.Hash(credentials.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	credentials.Password = hash

	err = h.storage.SignUp(c, credentials.Email, credentials.Password)
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

func (h *Handler) SignIn(c *gin.Context) {
	var credentials AuthCredentials

	if err := c.ShouldBindJSON(&credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.storage.SignIn(c, credentials.Email, credentials.Password)
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

func (h *Handler) UploadFile(c *gin.Context) {
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

	filename := h.newUUID()
	ext := strings.Split(contentType, "/")[1]
	filename = filename + "." + ext

	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	savedFile, err := h.storage.SaveFile(c, model.File{
		UserID:    oid,
		FileName:  filename,
		FileSize:  file.Size,
		FileType:  contentType,
		CreatedAt: h.now(),
		URL:       h.s3.ObjectURL(filename),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = h.s3.Upload(c, savedFile.ID.Hex(), f, file.Size, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		h.storage.DeleteFile(c, savedFile.ID)
		return
	}
	c.JSON(http.StatusOK, gin.H{"file": savedFile})
}

func (h *Handler) GetFiles(c *gin.Context) {
	userID := c.GetString("user_id")
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	files, err := h.storage.GetFiles(c, oid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"files": files})
}
