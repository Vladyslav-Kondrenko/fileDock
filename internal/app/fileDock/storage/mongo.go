package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	filedock "github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/fileDock"
	"github.com/golang-jwt/jwt/v5"
)

func SignUp(ctx context.Context, credentials filedock.UserCredentials) {
	fmt.Println(credentials)
	panic("signUp not implemented")
}

func SignIn(ctx context.Context, credentials filedock.UserCredentials) (string, error) {
	fmt.Println(credentials)
	// TODO: Implement DB logic

	ttl := os.Getenv("TTL")
	if ttl == "" {
		return "", errors.New("TTL is not set")
	}

	ttlInt, err := strconv.Atoi(ttl)
	if err != nil {
		return "", errors.New("TTL is not a number")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": credentials.Email,
		"exp": time.Now().Add(time.Duration(ttlInt) * time.Second).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", errors.New("failed to sign token")
	}

	return tokenString, nil
}

func UploadFile() {
	panic("uploadFile not implemented")
}

func GetFiles() {
	panic("getFiles not implemented")
}
