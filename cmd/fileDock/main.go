package main

import (
	"context"
	"log"

	"github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/handler"
	"github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/middleware"
	"github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/storage"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	ctx := context.Background()
	db, err := storage.InitDB(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Disconnect(ctx)

	s3Client, err := storage.NewS3Client(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if err := s3Client.EnsureBucket(ctx); err != nil {
		log.Fatal("ensure bucket: ", err)
	}

	router := gin.Default()
	router.POST("/sign-up", handler.SignUp)
	router.POST("/sign-in", handler.SignIn)
	router.POST("/upload", middleware.AuthMiddleware(), handler.UploadFile)
	router.GET("/files", middleware.AuthMiddleware(), handler.GetFiles)
	router.Run(":8080")

}
