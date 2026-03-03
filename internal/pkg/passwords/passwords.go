package passwords

import (
	"errors"
	"os"

	"golang.org/x/crypto/bcrypt"
)

const cost = 12

func HashPassword(password string) (string, error) {
	pepper := os.Getenv("PEPPER")
	if pepper == "" {
		return "", errors.New("PEPPER is not set")
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password+pepper), cost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	pepper := os.Getenv("PEPPER")
	if pepper == "" {
		return false
	}
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password+pepper))
	return err == nil
}
