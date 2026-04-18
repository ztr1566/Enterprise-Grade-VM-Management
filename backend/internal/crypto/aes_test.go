package crypto

import (
	"testing"
)

func TestEncrypt(t *testing.T) {
	key := []byte("this-is-a-32-byte-long-aes-key!!") // 32 bytes
	plaintext := "test-credential"

	ciphertext, err := Encrypt([]byte(plaintext), key)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	if len(ciphertext) == 0 {
		t.Fatal("Ciphertext is empty")
	}

	// Ciphertext should at least contain the nonce
	if len(ciphertext) <= 12 { // Standard GCM nonce size
		t.Fatalf("Ciphertext too short: %d bytes", len(ciphertext))
	}

	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if string(decrypted) != plaintext {
		t.Fatalf("Decrypted text mismatch: expected %s, got %s", plaintext, string(decrypted))
	}
}
