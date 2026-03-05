package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email     string             `json:"email" bson:"email"`
	Password  string             `json:"-" bson:"password"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
}

type File struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"user_id" bson:"user_id"`
	FileName  string             `json:"file_name" bson:"file_name"`
	FileSize  int64              `json:"file_size" bson:"file_size"`
	FileType  string             `json:"file_type" bson:"file_type"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	URL       string             `json:"url" bson:"url"`
}
