package service

import (
	"bytes"
	"errors"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Helper fixtures for image headers
var (
	// JPEG SOI (Start of Image) marker followed by JFIF marker
	validJPEGBytes = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01, 0x01, 0x01, 0x00, 0x48}
	// PNG signature (89 50 4E 47 0D 0A 1A 0A) followed by IHDR chunk
	validPNGBytes = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52}
	// WebP RIFF header (RIFF....WEBPVP8 )
	validWebPBytes = []byte("RIFF\x24\x00\x00\x00WEBPVP8 \x18\x00\x00\x00test-webp-payload-bytes")
	// Plain text header
	plainTextBytes = []byte("This is plain text and definitely not an image file.")
)

// createTestFileHeader builds a realistic *multipart.FileHeader in memory.
func createTestFileHeader(t *testing.T, fieldName, filename string, content []byte) *multipart.FileHeader {
	t.Helper()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatalf("failed to write content: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(int64(body.Len()) + 1024)
	if err != nil {
		t.Fatalf("failed to read multipart form: %v", err)
	}

	files := form.File[fieldName]
	if len(files) == 0 {
		t.Fatalf("no files found under field %s", fieldName)
	}
	return files[0]
}

func TestNewUploadService_Defaults(t *testing.T) {
	svc := NewUploadService("", 0)
	if svc.GetUploadDir() != "uploads" {
		t.Errorf("expected default uploadDir 'uploads', got '%s'", svc.GetUploadDir())
	}
	if svc.GetMaxFileSize() != DefaultMaxFileSize {
		t.Errorf("expected default maxFileSize %d, got %d", DefaultMaxFileSize, svc.GetMaxFileSize())
	}
}

func TestUploadImage_Success_Formats(t *testing.T) {
	tests := []struct {
		name          string
		filename      string
		content       []byte
		expectedExt   string
	}{
		{
			name:        "Valid JPEG",
			filename:    "photo.jpeg",
			content:     validJPEGBytes,
			expectedExt: ".jpg",
		},
		{
			name:        "Valid PNG",
			filename:    "photo.png",
			content:     validPNGBytes,
			expectedExt: ".png",
		},
		{
			name:        "Valid WebP",
			filename:    "photo.webp",
			content:     validWebPBytes,
			expectedExt: ".webp",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			svc := NewUploadService(tempDir, 1024*1024)

			header := createTestFileHeader(t, "image", tc.filename, tc.content)
			url, err := svc.UploadImage(header)
			if err != nil {
				t.Fatalf("unexpected error uploading %s: %v", tc.name, err)
			}

			if !strings.HasPrefix(url, "/uploads/") {
				t.Errorf("expected URL to start with /uploads/, got %s", url)
			}
			if !strings.HasSuffix(url, tc.expectedExt) {
				t.Errorf("expected URL to end with %s, got %s", tc.expectedExt, url)
			}

			// Verify file was written to disk
			diskFilename := strings.TrimPrefix(url, "/uploads/")
			diskPath := filepath.Join(tempDir, diskFilename)
			savedContent, err := os.ReadFile(diskPath)
			if err != nil {
				t.Fatalf("failed to read saved file from disk: %v", err)
			}
			if !bytes.Equal(savedContent, tc.content) {
				t.Errorf("saved content does not match original content")
			}
		})
	}
}

func TestUploadImage_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewUploadService(tempDir, 1024*1024)

	header := createTestFileHeader(t, "image", "empty.jpg", []byte{})
	_, err := svc.UploadImage(header)
	if !errors.Is(err, ErrEmptyFile) {
		t.Errorf("expected ErrEmptyFile, got %v", err)
	}

	// Nil header check
	_, errNil := svc.UploadImage(nil)
	if !errors.Is(errNil, ErrEmptyFile) {
		t.Errorf("expected ErrEmptyFile for nil, got %v", errNil)
	}
}

func TestUploadImage_FileTooLarge(t *testing.T) {
	tempDir := t.TempDir()
	maxSize := int64(len(validJPEGBytes) - 1)
	svc := NewUploadService(tempDir, maxSize)

	header := createTestFileHeader(t, "image", "toolarge.jpg", validJPEGBytes)
	_, err := svc.UploadImage(header)
	if !errors.Is(err, ErrFileTooLarge) {
		t.Errorf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestUploadImage_InvalidFileType(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewUploadService(tempDir, 1024*1024)

	header := createTestFileHeader(t, "image", "notes.txt", plainTextBytes)
	_, err := svc.UploadImage(header)
	if !errors.Is(err, ErrInvalidFileType) {
		t.Errorf("expected ErrInvalidFileType, got %v", err)
	}

	// Executable or text disguised as JPEG
	disguised := createTestFileHeader(t, "image", "fake.jpg", []byte("#!/bin/bash\necho hello"))
	_, errDisguised := svc.UploadImage(disguised)
	if !errors.Is(errDisguised, ErrInvalidFileType) {
		t.Errorf("expected ErrInvalidFileType for disguised file, got %v", errDisguised)
	}
}

func TestUploadImages_Batch_Success(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewUploadService(tempDir, 1024*1024)

	f1 := createTestFileHeader(t, "images", "img1.jpg", validJPEGBytes)
	f2 := createTestFileHeader(t, "images", "img2.png", validPNGBytes)
	f3 := createTestFileHeader(t, "images", "img3.webp", validWebPBytes)

	urls, err := svc.UploadImages([]*multipart.FileHeader{f1, f2, f3})
	if err != nil {
		t.Fatalf("unexpected batch error: %v", err)
	}

	if len(urls) != 3 {
		t.Fatalf("expected 3 URLs, got %d", len(urls))
	}

	for _, u := range urls {
		filename := strings.TrimPrefix(u, "/uploads/")
		if _, err := os.Stat(filepath.Join(tempDir, filename)); err != nil {
			t.Errorf("expected file %s to exist on disk: %v", filename, err)
		}
	}
}

func TestUploadImages_BatchLimitsAndValidation(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewUploadService(tempDir, 1024*1024)

	// Empty batch
	_, errEmpty := svc.UploadImages([]*multipart.FileHeader{})
	if !errors.Is(errEmpty, ErrEmptyFile) {
		t.Errorf("expected ErrEmptyFile for empty batch, got %v", errEmpty)
	}

	// Exceed max files limit (6 files > MaxFilesPerBatch of 5)
	files := make([]*multipart.FileHeader, 6)
	for i := 0; i < 6; i++ {
		files[i] = createTestFileHeader(t, "images", "img.jpg", validJPEGBytes)
	}
	_, errTooMany := svc.UploadImages(files)
	if !errors.Is(errTooMany, ErrTooManyFiles) {
		t.Errorf("expected ErrTooManyFiles, got %v", errTooMany)
	}

	// Atomic validation failure: 1st file valid, 2nd invalid -> none written
	goodFile := createTestFileHeader(t, "images", "good.jpg", validJPEGBytes)
	badFile := createTestFileHeader(t, "images", "bad.txt", plainTextBytes)
	_, errAtomic := svc.UploadImages([]*multipart.FileHeader{goodFile, badFile})
	if !errors.Is(errAtomic, ErrInvalidFileType) {
		t.Errorf("expected ErrInvalidFileType, got %v", errAtomic)
	}

	// Verify temp directory contains 0 saved files
	entries, _ := os.ReadDir(tempDir)
	if len(entries) != 0 {
		t.Errorf("expected 0 files written to disk after atomic validation failure, found %d", len(entries))
	}
}

func TestDeleteImage(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewUploadService(tempDir, 1024*1024)

	// Upload a file first
	header := createTestFileHeader(t, "image", "del.jpg", validJPEGBytes)
	url, err := svc.UploadImage(header)
	if err != nil {
		t.Fatalf("failed to upload file: %v", err)
	}

	// 1. Successful deletion by URL
	if err := svc.DeleteImage(url); err != nil {
		t.Fatalf("unexpected error deleting image: %v", err)
	}

	// 2. Delete non-existent file
	if err := svc.DeleteImage(url); !errors.Is(err, ErrFileNotFound) {
		t.Errorf("expected ErrFileNotFound, got %v", err)
	}

	// 3. Delete with empty filename
	if err := svc.DeleteImage("   "); !errors.Is(err, ErrInvalidFileName) {
		t.Errorf("expected ErrInvalidFileName, got %v", err)
	}

	// 4. Directory traversal attempt
	traversalAttempts := []string{
		"../secret.txt",
		"..\\secret.txt",
		"/uploads/../../etc/passwd",
		".",
		"/",
	}
	for _, attempt := range traversalAttempts {
		err := svc.DeleteImage(attempt)
		if err == nil {
			t.Errorf("expected error on path traversal attempt '%s', got nil", attempt)
		}
	}
}

func TestValidateImage(t *testing.T) {
	tempDir := t.TempDir()
	svc := NewUploadService(tempDir, 1024*1024)

	// Valid
	validHeader := createTestFileHeader(t, "img", "test.jpg", validJPEGBytes)
	if err := svc.ValidateImage(validHeader); err != nil {
		t.Errorf("expected valid image, got error: %v", err)
	}

	// Invalid format
	invalidHeader := createTestFileHeader(t, "img", "doc.pdf", []byte("%PDF-1.4 header"))
	if err := svc.ValidateImage(invalidHeader); !errors.Is(err, ErrInvalidFileType) {
		t.Errorf("expected ErrInvalidFileType, got %v", err)
	}

	// Oversized
	svcSmall := NewUploadService(tempDir, 5)
	if err := svcSmall.ValidateImage(validHeader); !errors.Is(err, ErrFileTooLarge) {
		t.Errorf("expected ErrFileTooLarge, got %v", err)
	}
}
