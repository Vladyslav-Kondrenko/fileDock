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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var usersColl *mongo.Collection
var filesColl *mongo.Collection

func InitDB(ctx context.Context) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("DATABASE_URL")))

	if err != nil {
		return nil, err
	}

	db := client.Database("fileDock")
	usersColl = db.Collection("users")
	filesColl = db.Collection("files")

	idx := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err = usersColl.Indexes().CreateOne(ctx, idx)
	if err != nil {
		return nil, err
	}

	return client, nil
}

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
