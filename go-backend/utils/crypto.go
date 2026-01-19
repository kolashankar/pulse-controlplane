package utils

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// GenerateSecureKey generates a cryptographically secure random key
func GenerateSecureKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateAPIKey generates a Pulse API key with prefix
func GenerateAPIKey() (string, error) {
	key, err := GenerateSecureKey(16)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pulse_key_%s", key), nil
}

// GenerateAPISecret generates a Pulse API secret
func GenerateAPISecret() (string, error) {
	secret, err := GenerateSecureKey(32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pulse_secret_%s", secret), nil
}

// HashSecret hashes a secret using bcrypt
func HashSecret(secret string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash secret: %w", err)
	}
	return string(hashed), nil
}

// VerifySecret compares a secret with its hash
func VerifySecret(hashedSecret, plainSecret string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedSecret), []byte(plainSecret))
}

// GenerateRandomToken generates a random token for invitations, etc.
func GenerateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// HashPassword hashes a user password
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashed), nil
}

// VerifyPassword compares a password with its hash
func VerifyPassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
