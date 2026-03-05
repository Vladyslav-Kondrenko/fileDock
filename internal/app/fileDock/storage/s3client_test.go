package storage

import (
	"context"
	"testing"
)

func TestNewS3Client_MissingEnv(t *testing.T) {
	t.Setenv("MINIO_ENDPOINT", "")
	t.Setenv("MINIO_ACCESS_KEY", "ak")
	t.Setenv("MINIO_SECRET_KEY", "sk")
	t.Setenv("MINIO_BUCKET", "bucket")

	client, err := NewS3Client(context.Background())
	if err == nil {
		t.Fatal("expected error when required env vars are missing")
	}
	if client != nil {
		t.Fatal("expected nil client on error")
	}
}

func TestNewS3Client_Success(t *testing.T) {
	t.Setenv("MINIO_ENDPOINT", "http://localhost:9000")
	t.Setenv("MINIO_ACCESS_KEY", "ak")
	t.Setenv("MINIO_SECRET_KEY", "sk")
	t.Setenv("MINIO_BUCKET", "uploads")

	client, err := NewS3Client(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.bucket != "uploads" {
		t.Fatalf("expected bucket %q, got %q", "uploads", client.bucket)
	}
}

func TestS3Client_ObjectURL(t *testing.T) {
	t.Setenv("MINIO_ENDPOINT", "http://localhost:9000")

	client := &S3Client{bucket: "uploads"}
	got := client.ObjectURL("image.png")
	want := "http://localhost:9000/uploads/image.png"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
