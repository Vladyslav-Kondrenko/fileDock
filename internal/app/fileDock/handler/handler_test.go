package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	filedock "github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/fileDock"
	"github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/storage"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestSignUp_InvalidPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/sign-up", SignUp)

	req := httptest.NewRequest(http.MethodPost, "/sign-up", bytes.NewBufferString(`{"email":"bad","password":"123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestSignIn_InvalidPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/sign-in", SignIn)

	req := httptest.NewRequest(http.MethodPost, "/sign-in", bytes.NewBufferString(`{"email":"not-an-email","password":"short"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUploadFile_NoUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/upload", UploadFile)

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestUploadFile_MissingFile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/upload", setUserIDMiddleware(primitive.NewObjectID().Hex()), UploadFile)

	req := httptest.NewRequest(http.MethodPost, "/upload", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUploadFile_InvalidFileType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/upload", setUserIDMiddleware(primitive.NewObjectID().Hex()), UploadFile)

	req := mustNewUploadRequest(t, "file.txt", []byte("hello"), "text/plain")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUploadFile_FileTooLarge(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/upload", setUserIDMiddleware(primitive.NewObjectID().Hex()), UploadFile)

	tooLarge := bytes.Repeat([]byte{1}, (10<<20)+1)
	req := mustNewUploadRequest(t, "big.png", tooLarge, "image/png")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUploadFile_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.POST("/upload", setUserIDMiddleware("not-a-valid-objectid"), UploadFile)

	req := mustNewUploadRequest(t, "image.png", []byte{1, 2, 3}, "image/png")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestGetFiles_InvalidUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/files", setUserIDMiddleware("not-a-valid-objectid"), GetFiles)

	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func setUserIDMiddleware(userID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

func mustNewUploadRequest(t *testing.T, filename string, content []byte, contentType string) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, filename))
	header.Set("Content-Type", contentType)

	part, err := writer.CreatePart(header)
	if err != nil {
		t.Fatalf("CreatePart: %v", err)
	}

	if _, err = io.Copy(part, bytes.NewReader(content)); err != nil {
		t.Fatalf("write multipart content: %v", err)
	}

	if err = writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

type mockStorage struct {
	signUpFn   func(*gin.Context, filedock.UserCredentials) error
	signInFn   func(*gin.Context, filedock.UserCredentials) (string, error)
	saveFileFn func(*gin.Context, filedock.File) (filedock.File, error)
	deleteFile func(*gin.Context, primitive.ObjectID) error
	getFilesFn func(*gin.Context, primitive.ObjectID) ([]filedock.File, error)
}

func (m mockStorage) SignUp(c *gin.Context, credentials filedock.UserCredentials) error {
	if m.signUpFn == nil {
		return nil
	}
	return m.signUpFn(c, credentials)
}

func (m mockStorage) SignIn(c *gin.Context, credentials filedock.UserCredentials) (string, error) {
	if m.signInFn == nil {
		return "", nil
	}
	return m.signInFn(c, credentials)
}

func (m mockStorage) SaveFile(c *gin.Context, file filedock.File) (filedock.File, error) {
	if m.saveFileFn == nil {
		return filedock.File{}, nil
	}
	return m.saveFileFn(c, file)
}

func (m mockStorage) DeleteFile(c *gin.Context, id primitive.ObjectID) error {
	if m.deleteFile == nil {
		return nil
	}
	return m.deleteFile(c, id)
}

func (m mockStorage) GetFiles(c *gin.Context, userID primitive.ObjectID) ([]filedock.File, error) {
	if m.getFilesFn == nil {
		return nil, nil
	}
	return m.getFilesFn(c, userID)
}

type mockPassword struct {
	hashFn func(string) (string, error)
}

func (m mockPassword) Hash(password string) (string, error) {
	if m.hashFn == nil {
		return password, nil
	}
	return m.hashFn(password)
}

type mockS3 struct {
	uploadFn    func(*gin.Context, string, io.Reader, int64, string) error
	objectURLFn func(string) string
}

func (m mockS3) Upload(c *gin.Context, key string, body io.Reader, contentLength int64, contentType string) error {
	if m.uploadFn == nil {
		return nil
	}
	return m.uploadFn(c, key, body, contentLength, contentType)
}

func (m mockS3) ObjectURL(key string) string {
	if m.objectURLFn == nil {
		return "http://example.local/" + key
	}
	return m.objectURLFn(key)
}

func TestSignIn_Success_WithInjectedStorage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := New(
		mockStorage{
			signInFn: func(_ *gin.Context, credentials filedock.UserCredentials) (string, error) {
				if credentials.Email != "test@example.com" {
					return "", storage.ErrInvalidCredentials
				}
				return "signed-token", nil
			},
		},
		mockPassword{},
		mockS3{},
	)

	router := gin.New()
	router.POST("/sign-in", h.SignIn)

	req := httptest.NewRequest(http.MethodPost, "/sign-in", bytes.NewBufferString(`{"email":"test@example.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "signed-token") {
		t.Fatalf("expected token in response body, got: %s", w.Body.String())
	}
}

func TestUploadFile_UploadError_TriggersDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	deleteCalled := false
	fileID := primitive.NewObjectID()
	h := New(
		mockStorage{
			saveFileFn: func(_ *gin.Context, file filedock.File) (filedock.File, error) {
				file.ID = fileID
				return file, nil
			},
			deleteFile: func(_ *gin.Context, id primitive.ObjectID) error {
				deleteCalled = true
				if id != fileID {
					t.Fatalf("expected delete id %s, got %s", fileID.Hex(), id.Hex())
				}
				return nil
			},
		},
		mockPassword{},
		mockS3{
			uploadFn: func(_ *gin.Context, _ string, _ io.Reader, _ int64, _ string) error {
				return errors.New("upload failed")
			},
		},
	)

	router := gin.New()
	router.POST("/upload", setUserIDMiddleware(primitive.NewObjectID().Hex()), h.UploadFile)

	req := mustNewUploadRequest(t, "image.png", []byte{1, 2, 3}, "image/png")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusInternalServerError, w.Code, w.Body.String())
	}
	if !deleteCalled {
		t.Fatal("expected delete file to be called on upload error")
	}
}

func TestSignUp_HashError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := New(
		mockStorage{},
		mockPassword{
			hashFn: func(string) (string, error) {
				return "", errors.New("hash failed")
			},
		},
		mockS3{},
	)

	router := gin.New()
	router.POST("/sign-up", h.SignUp)

	req := httptest.NewRequest(http.MethodPost, "/sign-up", bytes.NewBufferString(`{"email":"test@example.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestSignUp_EmailExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := New(
		mockStorage{
			signUpFn: func(_ *gin.Context, _ filedock.UserCredentials) error {
				return storage.ErrEmailExists
			},
		},
		mockPassword{
			hashFn: func(password string) (string, error) {
				return "hashed-" + password, nil
			},
		},
		mockS3{},
	)

	router := gin.New()
	router.POST("/sign-up", h.SignUp)

	req := httptest.NewRequest(http.MethodPost, "/sign-up", bytes.NewBufferString(`{"email":"test@example.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, w.Code)
	}
}

func TestSignUp_Success_UsesHashedPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	storageCalled := false
	h := New(
		mockStorage{
			signUpFn: func(_ *gin.Context, credentials filedock.UserCredentials) error {
				storageCalled = true
				if credentials.Password != "hashed-password123" {
					t.Fatalf("expected hashed password, got %q", credentials.Password)
				}
				return nil
			},
		},
		mockPassword{
			hashFn: func(password string) (string, error) {
				return "hashed-" + password, nil
			},
		},
		mockS3{},
	)

	router := gin.New()
	router.POST("/sign-up", h.SignUp)

	req := httptest.NewRequest(http.MethodPost, "/sign-up", bytes.NewBufferString(`{"email":"test@example.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	if !storageCalled {
		t.Fatal("expected storage SignUp to be called")
	}
}

func TestSignIn_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := New(
		mockStorage{
			signInFn: func(_ *gin.Context, _ filedock.UserCredentials) (string, error) {
				return "", storage.ErrInvalidCredentials
			},
		},
		mockPassword{},
		mockS3{},
	)

	router := gin.New()
	router.POST("/sign-in", h.SignIn)

	req := httptest.NewRequest(http.MethodPost, "/sign-in", bytes.NewBufferString(`{"email":"test@example.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSignIn_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := New(
		mockStorage{
			signInFn: func(_ *gin.Context, _ filedock.UserCredentials) (string, error) {
				return "", errors.New("db down")
			},
		},
		mockPassword{},
		mockS3{},
	)

	router := gin.New()
	router.POST("/sign-in", h.SignIn)

	req := httptest.NewRequest(http.MethodPost, "/sign-in", bytes.NewBufferString(`{"email":"test@example.com","password":"password123"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestGetFiles_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := primitive.NewObjectID()
	h := New(
		mockStorage{
			getFilesFn: func(_ *gin.Context, gotUserID primitive.ObjectID) ([]filedock.File, error) {
				if gotUserID != userID {
					t.Fatalf("expected user id %s, got %s", userID.Hex(), gotUserID.Hex())
				}
				return []filedock.File{{FileName: "a.png"}}, nil
			},
		},
		mockPassword{},
		mockS3{},
	)

	router := gin.New()
	router.GET("/files", setUserIDMiddleware(userID.Hex()), h.GetFiles)

	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "a.png") {
		t.Fatalf("expected file name in response body, got: %s", w.Body.String())
	}
}

func TestGetFiles_StorageError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := New(
		mockStorage{
			getFilesFn: func(_ *gin.Context, _ primitive.ObjectID) ([]filedock.File, error) {
				return nil, errors.New("db failed")
			},
		},
		mockPassword{},
		mockS3{},
	)

	router := gin.New()
	router.GET("/files", setUserIDMiddleware(primitive.NewObjectID().Hex()), h.GetFiles)

	req := httptest.NewRequest(http.MethodGet, "/files", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestUploadFile_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := primitive.NewObjectID()
	savedID := primitive.NewObjectID()
	uploadCalled := false
	h := New(
		mockStorage{
			saveFileFn: func(_ *gin.Context, file filedock.File) (filedock.File, error) {
				if file.UserID != userID {
					t.Fatalf("expected user id %s, got %s", userID.Hex(), file.UserID.Hex())
				}
				file.ID = savedID
				return file, nil
			},
		},
		mockPassword{},
		mockS3{
			uploadFn: func(_ *gin.Context, key string, _ io.Reader, _ int64, contentType string) error {
				uploadCalled = true
				if key != savedID.Hex() {
					t.Fatalf("expected upload key %s, got %s", savedID.Hex(), key)
				}
				if contentType != "image/png" {
					t.Fatalf("unexpected content type %q", contentType)
				}
				return nil
			},
		},
	)
	h.newUUID = func() string { return "fixed-uuid" }

	router := gin.New()
	router.POST("/upload", setUserIDMiddleware(userID.Hex()), h.UploadFile)

	req := mustNewUploadRequest(t, "image.png", []byte{1, 2, 3}, "image/png")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, w.Code, w.Body.String())
	}
	if !uploadCalled {
		t.Fatal("expected upload to be called")
	}
	if !strings.Contains(w.Body.String(), "fixed-uuid.png") {
		t.Fatalf("expected generated filename in response body, got: %s", w.Body.String())
	}
}
