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

	storage.DefaultS3Client, err = storage.NewS3Client(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if err := storage.DefaultS3Client.EnsureBucket(ctx); err != nil {
		log.Fatal("ensure bucket: ", err)
	}

	router := gin.Default()
	h := handler.NewDefault()
	router.POST("/sign-up", h.SignUp)
	router.POST("/sign-in", h.SignIn)
	router.POST("/upload", middleware.AuthMiddleware(), h.UploadFile)
	router.GET("/files", middleware.AuthMiddleware(), h.GetFiles)
	router.Run(":8080")

}
