![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/Vladyslav-Kondrenko/fileDock)
![Github Repository Size](https://img.shields.io/github/repo-size/Vladyslav-Kondrenko/fileDock)
![License](https://img.shields.io/badge/license-MIT-green)
![GitHub last commit](https://img.shields.io/github/last-commit/Vladyslav-Kondrenko/fileDock)

# FileDock

<img align="right" width="30%" src="./image.png">

## Task description

Web service with the following endpoints:

- **POST /sign-up** — User registration (email, password).
- **POST /sign-in** — Authentication (email, password). Returns 200 and a JWT access token on success, 401 on invalid credentials.
- **POST /upload** — File upload to S3-compatible Object Storage. Only PNG and JPEG images allowed, max 10 MB. Requires JWT authorization.
- **GET /files** — List of uploaded files (size, upload date, user id, link to storage). Requires JWT authorization.

Requirements: MongoDB, HTTP router (Gin), Docker and Docker Compose, Object Storage in compose (MinIO), password hashing, JWT authentication, README with run instructions.

## Implementation

- **Stack:** Go, Gin, MongoDB, MinIO (S3 API), JWT.
- **Structure:** `cmd/fileDock` — entry point; `internal/app/fileDock` — handler, storage (Mongo + S3 client), middleware (JWT); `internal/pkg` — passwords (bcrypt).
- **Run:** Application and dependencies (Mongo, MinIO) are started via Docker Compose. MinIO bucket is created on application startup.
- **Upload:** Type and size validation → save metadata to Mongo → upload to S3; on S3 error — rollback Mongo record.

## How to run

1. Clone the repository:
   ```bash
   git clone https://github.com/Vladyslav-Kondrenko/fileDock.git
   cd fileDock
   ```

2. Create `.env` from the example and edit if needed:
   ```bash
   cp .env.examle .env
   ```
   Set `JWT_SECRET` and `TTL` (in seconds). MinIO and Mongo `DATABASE_URL` are set in `docker-compose.yml` and can be overridden via `.env`.

3. Start services:
   ```bash
   docker compose up -d
   ```
   API is available at `http://localhost:8080`. MinIO Console at `http://localhost:9001` (login/password from `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` in `.env`).

4. Example requests:
   - Sign up: `curl -X POST http://localhost:8080/sign-up -H "Content-Type: application/json" -d '{"email":"user@example.com","password":"password123"}'`
   - Sign in: `curl -X POST http://localhost:8080/sign-in -H "Content-Type: application/json" -d '{"email":"user@example.com","password":"password123"}'`
   - Upload file: `curl -X POST http://localhost:8080/upload -H "Authorization: Bearer <token>" -F "file=@image.png"`
   - List files: `curl http://localhost:8080/files -H "Authorization: Bearer <token>"`

## Tests

```bash
go test -v ./...
```
