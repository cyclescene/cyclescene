package strava

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// TokenEncryption handles encryption/decryption of Strava tokens
type TokenEncryption struct {
	key []byte // 32 bytes for AES-256
}

// NewTokenEncryption creates a new token encryption handler
// Reads encryption key from STRAVA_TOKEN_ENCRYPTION_KEY env var
func NewTokenEncryption() (*TokenEncryption, error) {
	keyB64 := os.Getenv("STRAVA_TOKEN_ENCRYPTION_KEY")
	if keyB64 == "" {
		return nil, fmt.Errorf("STRAVA_TOKEN_ENCRYPTION_KEY environment variable not set")
	}

	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return nil, fmt.Errorf("invalid encryption key format: %w", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes (AES-256), got %d bytes", len(key))
	}

	return &TokenEncryption{key: key}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns (ciphertext, nonce, error)
func (e *TokenEncryption) Encrypt(plaintext string) ([]byte, []byte, error) {
	// Create cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nil, nonce, []byte(plaintext), nil)

	return ciphertext, nonce, nil
}

// Decrypt decrypts ciphertext using the provided nonce and returns plaintext
func (e *TokenEncryption) Decrypt(ciphertext, nonce []byte) (string, error) {
	// Create cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// GenerateKey generates a new 32-byte encryption key and returns it as base64
// This is a utility function for initial setup - not used in production code
func GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
