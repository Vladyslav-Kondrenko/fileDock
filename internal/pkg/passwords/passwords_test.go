package passwords

import (
	"testing"
)

const testPepper = "test-pepper"

func TestHashPassword_Success(t *testing.T) {
	t.Setenv("PEPPER", testPepper)
	password := "secret123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: unexpected error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword: expected non-empty hash")
	}
	if hash == password {
		t.Fatal("HashPassword: hash must not equal plain password")
	}
}

func TestHashPassword_NoPepper(t *testing.T) {
	t.Setenv("PEPPER", "")
	_, err := HashPassword("secret")
	if err == nil {
		t.Fatal("HashPassword: expected error when PEPPER is not set")
	}
}

func TestCheckPasswordHash_Match(t *testing.T) {
	t.Setenv("PEPPER", testPepper)
	password := "mypassword"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if !CheckPasswordHash(password, hash) {
		t.Error("CheckPasswordHash: expected true for correct password")
	}
}

func TestCheckPasswordHash_WrongPassword(t *testing.T) {
	t.Setenv("PEPPER", testPepper)
	hash, err := HashPassword("correct")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if CheckPasswordHash("wrong", hash) {
		t.Error("CheckPasswordHash: expected false for wrong password")
	}
}

func TestCheckPasswordHash_EmptyHash(t *testing.T) {
	t.Setenv("PEPPER", testPepper)
	if CheckPasswordHash("any", "") {
		t.Error("CheckPasswordHash: expected false for empty hash")
	}
}

func TestCheckPasswordHash_NoPepper(t *testing.T) {
	t.Setenv("PEPPER", testPepper)
	hash, err := HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	t.Setenv("PEPPER", "")
	if CheckPasswordHash("secret", hash) {
		t.Error("CheckPasswordHash: expected false when PEPPER is not set")
	}
}
