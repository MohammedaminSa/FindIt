package hasher

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrEmptyPassword is returned when an empty password string is provided.
	ErrEmptyPassword = errors.New("password cannot be empty")
	// ErrPasswordTooLong is returned when a password exceeds the bcrypt 72-byte limit.
	ErrPasswordTooLong = errors.New("password exceeds maximum allowed length of 72 bytes")
)

// HashPassword hashes a plain text password using bcrypt with the default cost.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}
	if len(password) > 72 {
		return "", ErrPasswordTooLong
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a plain text password against a bcrypt hashed password.
// Returns true if the password matches the hash, false otherwise.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
