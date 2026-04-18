package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
)

// MustGetAESKey retrieves the AES_KEY from the environment, validates its length (32 bytes for AES-256),
// and exits the application if the key is missing or invalid.
func MustGetAESKey() []byte {
	keyStr := os.Getenv("AES_KEY")
	if keyStr == "" {
		log.Fatal("CRITICAL: AES_KEY environment variable is missing. Encryption cannot proceed.")
	}

	// Try to decode as base64 (since openssl rand -base64 was used)
	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		// If not base64, try raw bytes (fallback)
		key = []byte(keyStr)
	}

	if len(key) != 32 {
		log.Fatalf("CRITICAL: AES_KEY must be exactly 32 bytes for AES-256. Current length: %d bytes", len(key))
	}

	return key
}

// Encrypt encrypts plain text using AES-GCM with the provided key.
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts cipher text using AES-GCM with the provided key.
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
