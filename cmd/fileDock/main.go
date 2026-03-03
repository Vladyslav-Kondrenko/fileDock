package main

import (
	"fmt"
	"log"

	"github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/handler"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("Hello, products!")

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	// conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer conn.Close(context.Background())
	// storage.InitDB(conn)
	router := gin.Default()
	router.POST("/sign-up", handler.SignUp)
	router.POST("/sign-in", handler.SignIn)
	router.POST("/upload", handler.UploadFile)
	router.GET("/files", handler.GetFiles)
	router.Run(":8080")

}
