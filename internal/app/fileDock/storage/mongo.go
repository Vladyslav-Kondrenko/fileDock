package storage

import (
	"context"
	"fmt"

	filedock "github.com/Vladyslav-Kondrenko/fileDock/internal/app/fileDock/fileDock"
)

func SignUp(ctx context.Context, credentials filedock.UserCredentials) {
	fmt.Println(credentials)
	panic("signUp not implemented")
}

func SignIn() {
	panic("signIn not implemented")
}

func UploadFile() {
	panic("uploadFile not implemented")
}

func GetFiles() {
	panic("getFiles not implemented")
}
