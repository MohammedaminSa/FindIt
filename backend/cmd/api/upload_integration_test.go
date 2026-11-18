package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MohammedaminSa/FindIt/backend/internal/handler"
	"github.com/MohammedaminSa/FindIt/backend/internal/middleware"
	"github.com/MohammedaminSa/FindIt/backend/internal/service"
	"github.com/MohammedaminSa/FindIt/backend/pkg/token"
	"github.com/google/uuid"
)

var (
	integrationJPEG = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x48}
	integrationPNG  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52}
)

func createTestUploadMultipartBody(t *testing.T, fieldName string, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("failed to write part content: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	return body, writer.FormDataContentType()
}

func TestUploadAndStaticServe_EndToEnd(t *testing.T) {
	tempUploadDir := t.TempDir()

	jwtMgr, err := token.NewJWTManager("test-secret-key-at-least-32-bytes-long", time.Hour)
	if err != nil {
		t.Fatalf("failed to init JWT: %v", err)
	}
	authMw := middleware.Auth(jwtMgr)

	uploadSvc := service.NewUploadService(tempUploadDir, 5*1024*1024)
	uploadHdl := handler.NewUploadHandler(uploadSvc)

	router := setupRouter(nil, nil, nil, uploadHdl, tempUploadDir, authMw)

	// Step 1: Attempt upload without auth token -> 401 Unauthorized
	bodyNoAuth, cTypeNoAuth := createTestUploadMultipartBody(t, "image", "sample.jpg", integrationJPEG)
	reqNoAuth, _ := http.NewRequest(http.MethodPost, "/api/v1/uploads", bodyNoAuth)
	reqNoAuth.Header.Set("Content-Type", cTypeNoAuth)
	wNoAuth := httptest.NewRecorder()
	router.ServeHTTP(wNoAuth, reqNoAuth)

	if wNoAuth.Code != http.StatusUnauthorized {
		t.Fatalf("expected HTTP 401 without token, got %d", wNoAuth.Code)
	}

	// Step 2: Generate valid token and upload image -> 201 Created
	tokenStr, err := jwtMgr.GenerateToken(uuid.New(), "user@example.com")
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	bodyAuth, cTypeAuth := createTestUploadMultipartBody(t, "image", "camera.jpg", integrationJPEG)
	reqAuth, _ := http.NewRequest(http.MethodPost, "/api/v1/uploads", bodyAuth)
	reqAuth.Header.Set("Content-Type", cTypeAuth)
	reqAuth.Header.Set("Authorization", "Bearer "+tokenStr)
	wAuth := httptest.NewRecorder()
	router.ServeHTTP(wAuth, reqAuth)

	if wAuth.Code != http.StatusCreated {
		t.Fatalf("expected HTTP 201, got %d: %s", wAuth.Code, wAuth.Body.String())
	}

	var res map[string]interface{}
	if err := json.Unmarshal(wAuth.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse upload response: %v", err)
	}

	uploadedURL, ok := res["url"].(string)
	if !ok || uploadedURL == "" {
		t.Fatalf("expected non-empty 'url' in response")
	}

	// Step 3: Fetch the uploaded image through static file serving -> 200 OK
	reqStatic, _ := http.NewRequest(http.MethodGet, uploadedURL, nil)
	wStatic := httptest.NewRecorder()
	router.ServeHTTP(wStatic, reqStatic)

	if wStatic.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 fetching static file %s, got %d", uploadedURL, wStatic.Code)
	}

	if !bytes.Equal(wStatic.Body.Bytes(), integrationJPEG) {
		t.Errorf("served static content bytes did not match uploaded image bytes")
	}

	// Step 4: Delete the uploaded image via API -> 200 OK
	reqDelete, _ := http.NewRequest(http.MethodDelete, "/api/v1"+uploadedURL, nil)
	reqDelete.Header.Set("Authorization", "Bearer "+tokenStr)
	wDelete := httptest.NewRecorder()
	router.ServeHTTP(wDelete, reqDelete)

	if wDelete.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 deleting file, got %d", wDelete.Code)
	}

	// Step 5: Verify static endpoint now returns 404 Not Found
	reqStaticAfter, _ := http.NewRequest(http.MethodGet, uploadedURL, nil)
	wStaticAfter := httptest.NewRecorder()
	router.ServeHTTP(wStaticAfter, reqStaticAfter)

	if wStaticAfter.Code != http.StatusNotFound {
		t.Errorf("expected HTTP 404 after file deletion, got %d", wStaticAfter.Code)
	}
}
