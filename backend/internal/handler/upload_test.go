package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MohammedaminSa/FindIt/backend/internal/service"
	"github.com/gin-gonic/gin"
)

var (
	testJPEG = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x48}
	testPNG  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52}
)

func createMultipartBody(t *testing.T, fieldName string, files map[string][]byte) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for filename, content := range files {
		part, err := writer.CreateFormFile(fieldName, filename)
		if err != nil {
			t.Fatalf("failed to create form file: %v", err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatalf("failed to write part content: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close multipart writer: %v", err)
	}

	return body, writer.FormDataContentType()
}

func setupUploadTestRouter(svc service.UploadService, authMw gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	h := NewUploadHandler(svc)
	h.RegisterRoutes(v1, authMw)
	return router
}

func TestUploadHandler_Upload_SingleSuccess(t *testing.T) {
	tempDir := t.TempDir()
	svc := service.NewUploadService(tempDir, 1024*1024)
	router := setupUploadTestRouter(svc, nil)

	body, contentType := createMultipartBody(t, "image", map[string][]byte{
		"photo.jpg": testJPEG,
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/uploads", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected HTTP 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var res map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	url, ok := res["url"].(string)
	if !ok || url == "" {
		t.Errorf("expected 'url' in response, got %v", res["url"])
	}

	urls, ok := res["urls"].([]interface{})
	if !ok || len(urls) != 1 {
		t.Errorf("expected 'urls' array with 1 item, got %v", res["urls"])
	}
}

func TestUploadHandler_Upload_AliasSuccess(t *testing.T) {
	tempDir := t.TempDir()
	svc := service.NewUploadService(tempDir, 1024*1024)
	router := setupUploadTestRouter(svc, nil)

	body, contentType := createMultipartBody(t, "file", map[string][]byte{
		"photo.png": testPNG,
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/upload", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected HTTP 201 on /upload alias, got %d", w.Code)
	}
}

func TestUploadHandler_Upload_MultipleSuccess(t *testing.T) {
	tempDir := t.TempDir()
	svc := service.NewUploadService(tempDir, 1024*1024)
	router := setupUploadTestRouter(svc, nil)

	body, contentType := createMultipartBody(t, "images", map[string][]byte{
		"photo1.jpg": testJPEG,
		"photo2.png": testPNG,
	})

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/uploads/multiple", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected HTTP 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var res map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response JSON: %v", err)
	}

	urls, ok := res["urls"].([]interface{})
	if !ok || len(urls) != 2 {
		t.Fatalf("expected 2 URLs in response, got %v", res["urls"])
	}
}

func TestUploadHandler_Upload_Errors(t *testing.T) {
	tempDir := t.TempDir()
	svc := service.NewUploadService(tempDir, 1024*1024)
	router := setupUploadTestRouter(svc, nil)

	t.Run("No files provided", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/uploads", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected HTTP 400, got %d", w.Code)
		}
	})

	t.Run("Invalid file type", func(t *testing.T) {
		body, contentType := createMultipartBody(t, "image", map[string][]byte{
			"malicious.txt": []byte("just plain text"),
		})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/uploads", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected HTTP 400 for invalid file type, got %d", w.Code)
		}
	})

	t.Run("File too large", func(t *testing.T) {
		smallSvc := service.NewUploadService(tempDir, 5)
		smallRouter := setupUploadTestRouter(smallSvc, nil)

		body, contentType := createMultipartBody(t, "image", map[string][]byte{
			"large.jpg": testJPEG,
		})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/uploads", body)
		req.Header.Set("Content-Type", contentType)
		w := httptest.NewRecorder()
		smallRouter.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected HTTP 400 for oversized file, got %d", w.Code)
		}
	})
}

func TestUploadHandler_Delete(t *testing.T) {
	tempDir := t.TempDir()
	svc := service.NewUploadService(tempDir, 1024*1024)
	router := setupUploadTestRouter(svc, nil)

	// First upload a file
	body, contentType := createMultipartBody(t, "image", map[string][]byte{
		"photo.jpg": testJPEG,
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/uploads", body)
	req.Header.Set("Content-Type", contentType)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	url := res["url"].(string) // e.g. "/uploads/uuid.jpg"

	t.Run("Delete existing file", func(t *testing.T) {
		reqDel, _ := http.NewRequest(http.MethodDelete, "/api/v1"+url, nil)
		wDel := httptest.NewRecorder()
		router.ServeHTTP(wDel, reqDel)

		if wDel.Code != http.StatusOK {
			t.Errorf("expected HTTP 200 on delete, got %d", wDel.Code)
		}
	})

	t.Run("Delete non-existent file", func(t *testing.T) {
		reqDel, _ := http.NewRequest(http.MethodDelete, "/api/v1/uploads/doesnotexist.jpg", nil)
		wDel := httptest.NewRecorder()
		router.ServeHTTP(wDel, reqDel)

		if wDel.Code != http.StatusNotFound {
			t.Errorf("expected HTTP 404 for non-existent file, got %d", wDel.Code)
		}
	})
}
