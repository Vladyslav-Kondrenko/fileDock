package filedock

import "time"

type User struct {
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
}

type UserCredentials struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type File struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	FileName  string    `json:"file_name"`
	FileSize  int64     `json:"file_size"`
	FileType  string    `json:"file_type"`
	CreatedAt time.Time `json:"created_at"`
}
