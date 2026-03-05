package storage

import (
	"context"
	"errors"
	"os"
	"strconv"
	"time"

	filedock "github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/fileDock"
	"github.com/Vladyslav-Kondrenko/fileDock/internal/pkg/passwords"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var usersColl *mongo.Collection
var filesColl *mongo.Collection

var ErrEmailExists = errors.New("email already registered")
var ErrInvalidCredentials = errors.New("invalid credentials")

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

func SignUp(ctx context.Context, credentials filedock.UserCredentials) error {
	user := filedock.User{
		Email:     credentials.Email,
		Password:  credentials.Password,
		CreatedAt: time.Now(),
	}

	result, err := usersColl.InsertOne(ctx, user)
	if err != nil {
		var we mongo.WriteException
		if errors.As(err, &we) && len(we.WriteErrors) > 0 && we.WriteErrors[0].Code == 11000 {
			return ErrEmailExists
		}
		return err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func SignIn(ctx context.Context, credentials filedock.UserCredentials) (string, error) {

	ttl := os.Getenv("TTL")
	if ttl == "" {
		return "", errors.New("TTL is not set")
	}

	ttlInt, err := strconv.Atoi(ttl)
	if err != nil {
		return "", errors.New("TTL is not a number")
	}

	user := filedock.User{}
	err = usersColl.FindOne(ctx, bson.M{"email": credentials.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	if !passwords.CheckPasswordHash(credentials.Password, user.Password) {
		return "", ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID.Hex(),
		"exp": time.Now().Add(time.Duration(ttlInt) * time.Second).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", errors.New("failed to sign token")
	}

	return tokenString, nil
}

func SaveFile(ctx context.Context, file filedock.File) (filedock.File, error) {
	result, err := filesColl.InsertOne(ctx, file)
	if err != nil {
		return filedock.File{}, err
	}
	file.ID = result.InsertedID.(primitive.ObjectID)
	return file, nil
}

func DeleteFile(ctx context.Context, id primitive.ObjectID) error {
	_, err := filesColl.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func GetFiles(ctx context.Context, userId primitive.ObjectID) ([]filedock.File, error) {
	cursor, err := filesColl.Find(ctx, bson.M{"user_id": userId})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	files := []filedock.File{}
	for cursor.Next(ctx) {
		var file filedock.File
		err = cursor.Decode(&file)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}
