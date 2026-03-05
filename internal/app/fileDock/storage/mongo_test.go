package storage

import (
	"context"
	"testing"
)

func TestSignIn_TTLNotSet(t *testing.T) {
	t.Setenv("TTL", "")

	_, err := SignIn(context.Background(), "user@example.com", "password123")
	if err == nil {
		t.Fatal("expected error when TTL is not set")
	}
	if err.Error() != "TTL is not set" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSignIn_TTLNotNumber(t *testing.T) {
	t.Setenv("TTL", "abc")

	_, err := SignIn(context.Background(), "user@example.com", "password123")
	if err == nil {
		t.Fatal("expected error when TTL is not a number")
	}
	if err.Error() != "TTL is not a number" {
		t.Fatalf("unexpected error: %v", err)
	}
}
