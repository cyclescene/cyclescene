package strava

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestTokenEncryption_EncryptDecrypt(t *testing.T) {
	enc := NewTestTokenEncryption(t)

	tests := []struct {
		name    string
		input   string
	}{
		{"simple token", "refresh_token_abc123"},
		{"empty string", ""},
		{"long token", "this_is_a_very_long_refresh_token_that_exceeds_typical_lengths_and_tests_handling_of_longer_strings"},
		{"special characters", "token_with_special_chars!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"unicode", "token_with_unicode_🚴‍♂️_characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			ciphertext, nonce, err := enc.Encrypt(tt.input)
			AssertNoError(t, err)

			// Nonce should be the expected size for GCM (12 bytes)
			AssertEqual(t, 12, len(nonce))

			// Ciphertext should not be empty (except for empty input where it contains only auth tag)
			// GCM adds a 16-byte auth tag, so even empty plaintext produces ciphertext
			if tt.input != "" {
				AssertTrue(t, len(ciphertext) > 16, "Ciphertext should contain data beyond auth tag")
			}

			// Decrypt
			decrypted, err := enc.Decrypt(ciphertext, nonce)
			AssertNoError(t, err)
			AssertEqual(t, tt.input, decrypted)
		})
	}
}

func TestTokenEncryption_DifferentNoncesForSameInput(t *testing.T) {
	enc := NewTestTokenEncryption(t)

	input := "same_token"

	// Encrypt twice
	ciphertext1, nonce1, err := enc.Encrypt(input)
	AssertNoError(t, err)

	ciphertext2, nonce2, err := enc.Encrypt(input)
	AssertNoError(t, err)

	// Nonces should be different (random)
	AssertTrue(t, !bytesEqual(nonce1, nonce2), "Nonces should be different for each encryption")

	// Ciphertexts should be different (due to different nonces)
	AssertTrue(t, !bytesEqual(ciphertext1, ciphertext2), "Ciphertexts should be different due to different nonces")

	// But both should decrypt to the same value
	decrypted1, _ := enc.Decrypt(ciphertext1, nonce1)
	decrypted2, _ := enc.Decrypt(ciphertext2, nonce2)

	AssertEqual(t, input, decrypted1)
	AssertEqual(t, input, decrypted2)
}

func TestTokenEncryption_WrongNonceFails(t *testing.T) {
	enc := NewTestTokenEncryption(t)

	input := "my_token"

	// Encrypt
	ciphertext, nonce, err := enc.Encrypt(input)
	AssertNoError(t, err)

	// Try to decrypt with wrong nonce
	wrongNonce := make([]byte, 12)
	for i := range wrongNonce {
		wrongNonce[i] = nonce[i] ^ 0xFF // Flip all bits
	}

	_, err = enc.Decrypt(ciphertext, wrongNonce)
	AssertError(t, err)
}

func TestTokenEncryption_TamperedCiphertextFails(t *testing.T) {
	enc := NewTestTokenEncryption(t)

	input := "my_token"

	// Encrypt
	ciphertext, nonce, err := enc.Encrypt(input)
	AssertNoError(t, err)

	// Tamper with ciphertext
	tamperedCiphertext := make([]byte, len(ciphertext))
	copy(tamperedCiphertext, ciphertext)
	tamperedCiphertext[0] ^= 0xFF // Flip bits in first byte

	_, err = enc.Decrypt(tamperedCiphertext, nonce)
	AssertError(t, err)
}

func TestNewTokenEncryption_MissingKey(t *testing.T) {
	// Ensure key is not set
	os.Unsetenv("STRAVA_TOKEN_ENCRYPTION_KEY")

	_, err := NewTokenEncryption()
	AssertError(t, err)
}

func TestNewTokenEncryption_InvalidKeyLength(t *testing.T) {
	// Set a key that decodes to wrong length (not 32 bytes)
	shortKey := base64.StdEncoding.EncodeToString([]byte("too_short"))
	os.Setenv("STRAVA_TOKEN_ENCRYPTION_KEY", shortKey)
	defer os.Unsetenv("STRAVA_TOKEN_ENCRYPTION_KEY")

	_, err := NewTokenEncryption()
	AssertError(t, err)
}

func TestNewTokenEncryption_InvalidBase64(t *testing.T) {
	// Set a key that is not valid base64
	os.Setenv("STRAVA_TOKEN_ENCRYPTION_KEY", "not_valid_base64!!!")
	defer os.Unsetenv("STRAVA_TOKEN_ENCRYPTION_KEY")

	_, err := NewTokenEncryption()
	AssertError(t, err)
}

func TestNewTokenEncryption_ValidKey(t *testing.T) {
	// Generate a valid 32-byte key
	validKey := make([]byte, 32)
	for i := range validKey {
		validKey[i] = byte(i)
	}
	encodedKey := base64.StdEncoding.EncodeToString(validKey)

	os.Setenv("STRAVA_TOKEN_ENCRYPTION_KEY", encodedKey)
	defer os.Unsetenv("STRAVA_TOKEN_ENCRYPTION_KEY")

	enc, err := NewTokenEncryption()
	AssertNoError(t, err)
	AssertTrue(t, enc != nil, "Encryption should be created")
}

// Helper function for byte comparison
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
