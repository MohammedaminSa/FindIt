package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const (
	// DefaultMaxFileSize defines the maximum allowed image file size (5 MB).
	DefaultMaxFileSize int64 = 5 * 1024 * 1024
	// MaxFilesPerBatch defines the upper limit for images uploaded in a single batch.
	MaxFilesPerBatch = 5
)

var (
	// ErrEmptyFile is returned when an uploaded file is nil or contains 0 bytes.
	ErrEmptyFile = errors.New("uploaded file is empty")
	// ErrFileTooLarge is returned when a file exceeds the maximum allowed file size.
	ErrFileTooLarge = errors.New("file size exceeds maximum allowed limit (5MB)")
	// ErrInvalidFileType is returned when a file format is not JPEG, PNG, or WebP.
	ErrInvalidFileType = errors.New("unsupported file format: only JPEG, PNG, and WebP images are allowed")
	// ErrTooManyFiles is returned when an upload batch exceeds the maximum allowed count.
	ErrTooManyFiles = errors.New("too many files: maximum 5 images allowed per upload")
	// ErrFileNotFound is returned when attempting to access or delete a file that does not exist.
	ErrFileNotFound = errors.New("file not found")
	// ErrInvalidFileName is returned when a file path or name is invalid, unsafe, or malformed.
	ErrInvalidFileName = errors.New("invalid file name")
)

// UploadService defines business capabilities for validating, persisting, and deleting media files.
type UploadService interface {
	UploadImage(file *multipart.FileHeader) (string, error)
	UploadImages(files []*multipart.FileHeader) ([]string, error)
	ValidateImage(file *multipart.FileHeader) error
	DeleteImage(imagePath string) error
	GetUploadDir() string
	GetMaxFileSize() int64
}

// LocalUploadService implements UploadService with storage on the local filesystem.
type LocalUploadService struct {
	uploadDir   string
	maxFileSize int64
}

// NewUploadService creates a new LocalUploadService with the provided directory and size threshold.
func NewUploadService(uploadDir string, maxFileSize int64) *LocalUploadService {
	if uploadDir == "" {
		uploadDir = "uploads"
	}
	if maxFileSize <= 0 {
		maxFileSize = DefaultMaxFileSize
	}
	return &LocalUploadService{
		uploadDir:   uploadDir,
		maxFileSize: maxFileSize,
	}
}

// GetUploadDir returns the configured upload directory path.
func (s *LocalUploadService) GetUploadDir() string {
	return s.uploadDir
}

// GetMaxFileSize returns the maximum permitted file size in bytes.
func (s *LocalUploadService) GetMaxFileSize() int64 {
	return s.maxFileSize
}

// ValidateImage inspects a multipart file header, verifies non-zero size, enforces the
// maximum size boundary, and sniffs content magic bytes to assure it is a genuine JPEG, PNG, or WebP image.
func (s *LocalUploadService) ValidateImage(file *multipart.FileHeader) error {
	if file == nil || file.Size == 0 {
		return ErrEmptyFile
	}

	if file.Size > s.maxFileSize {
		return ErrFileTooLarge
	}

	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Sniff magic bytes from the first 512 bytes
	buf := make([]byte, 512)
	n, err := io.ReadFull(src, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return fmt.Errorf("failed to read file header: %w", err)
	}

	mimeType := detectMIMEType(buf[:n])
	if !isAllowedMIMEType(mimeType) {
		return ErrInvalidFileType
	}

	return nil
}

// UploadImage validates and persists a single image file to the upload directory under a secure UUID filename.
// Returns the public relative URL (e.g. "/uploads/<uuid>.<ext>").
func (s *LocalUploadService) UploadImage(file *multipart.FileHeader) (string, error) {
	if file == nil || file.Size == 0 {
		return "", ErrEmptyFile
	}

	if file.Size > s.maxFileSize {
		return "", ErrFileTooLarge
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Sniff magic bytes
	buf := make([]byte, 512)
	n, err := io.ReadFull(src, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", fmt.Errorf("failed to read file header: %w", err)
	}

	mimeType := detectMIMEType(buf[:n])
	if !isAllowedMIMEType(mimeType) {
		return "", ErrInvalidFileType
	}

	// Rewind the stream to the beginning
	seeker, ok := src.(io.Seeker)
	if !ok {
		return "", errors.New("file stream is not seekable")
	}
	if _, err := seeker.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to reset file reader: %w", err)
	}

	// Ensure target directory exists
	if err := os.MkdirAll(s.uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to initialize upload directory: %w", err)
	}

	// Generate unique filename with extension mapped from verified MIME type
	ext := getExtensionForMIME(mimeType)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dstPath := filepath.Join(s.uploadDir, filename)

	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}

	// Copy stream with hard limit to prevent disk bloat
	written, copyErr := io.Copy(dst, io.LimitReader(src, s.maxFileSize+1))
	closeErr := dst.Close()

	if copyErr != nil {
		_ = os.Remove(dstPath)
		return "", fmt.Errorf("failed to save file: %w", copyErr)
	}
	if closeErr != nil {
		_ = os.Remove(dstPath)
		return "", fmt.Errorf("failed to flush destination file: %w", closeErr)
	}

	if written > s.maxFileSize {
		_ = os.Remove(dstPath)
		return "", ErrFileTooLarge
	}

	return "/uploads/" + filename, nil
}

// UploadImages validates and persists a batch of up to MaxFilesPerBatch images atomically.
// If any file fails validation or fails to write, any already written files are cleaned up.
func (s *LocalUploadService) UploadImages(files []*multipart.FileHeader) ([]string, error) {
	if len(files) == 0 {
		return nil, ErrEmptyFile
	}

	if len(files) > MaxFilesPerBatch {
		return nil, ErrTooManyFiles
	}

	// Step 1: Pre-validate all files before committing any to disk
	for _, file := range files {
		if err := s.ValidateImage(file); err != nil {
			return nil, err
		}
	}

	// Step 2: Persist all files with atomic rollback on unexpected write failures
	urls := make([]string, 0, len(files))
	for _, file := range files {
		url, err := s.UploadImage(file)
		if err != nil {
			// Clean up all previously written files in this batch
			for _, savedURL := range urls {
				_ = s.DeleteImage(savedURL)
			}
			return nil, err
		}
		urls = append(urls, url)
	}

	return urls, nil
}

// DeleteImage safely removes an uploaded file from disk given its relative URL or filename.
// Prevents directory traversal attacks by extracting strictly the base filename.
func (s *LocalUploadService) DeleteImage(imagePath string) error {
	imagePath = strings.TrimSpace(imagePath)
	if imagePath == "" {
		return ErrInvalidFileName
	}

	cleanPath := filepath.Clean(imagePath)
	filename := filepath.Base(cleanPath)

	if filename == "." || filename == "/" || filename == "\\" {
		return ErrInvalidFileName
	}

	if strings.Contains(filename, "..") || strings.ContainsAny(filename, `/\`) {
		return ErrInvalidFileName
	}

	fullPath := filepath.Join(s.uploadDir, filename)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return fmt.Errorf("failed to access file: %w", err)
	}

	if info.IsDir() {
		return ErrInvalidFileName
	}

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to remove file: %w", err)
	}

	return nil
}

// detectMIMEType sniffs bytes for JPEG, PNG, and WebP format signatures.
func detectMIMEType(header []byte) string {
	if len(header) == 0 {
		return ""
	}

	// Check WebP RIFF magic container:
	// Bytes 0-3: "RIFF", Bytes 8-11: "WEBP"
	if len(header) >= 12 && string(header[0:4]) == "RIFF" && string(header[8:12]) == "WEBP" {
		return "image/webp"
	}

	mime := http.DetectContentType(header)
	if semi := strings.Index(mime, ";"); semi != -1 {
		mime = strings.TrimSpace(mime[:semi])
	}

	return mime
}

// isAllowedMIMEType checks whether the detected MIME type is accepted.
func isAllowedMIMEType(mime string) bool {
	switch mime {
	case "image/jpeg", "image/png", "image/webp":
		return true
	default:
		return false
	}
}

// getExtensionForMIME maps a verified MIME type to its canonical file extension.
func getExtensionForMIME(mime string) string {
	switch mime {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ".bin"
	}
}
