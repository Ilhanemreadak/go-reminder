package services

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"testing"
)

func TestEncryptionService(t *testing.T) {
	// Generate a test encryption key (32 bytes for AES-256)
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	keyStr := base64.StdEncoding.EncodeToString(key)

	// Set environment variable
	os.Setenv("ENCRYPTION_KEY", keyStr)
	defer os.Unsetenv("ENCRYPTION_KEY")

	// Create encryption service
	service, err := NewEncryptionService()
	if err != nil {
		t.Fatalf("failed to create encryption service: %v", err)
	}

	t.Run("encrypt and decrypt", func(t *testing.T) {
		plaintext := "my-secret-password"

		// Encrypt
		ciphertext, err := service.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("failed to encrypt: %v", err)
		}

		// Verify ciphertext is different from plaintext
		if ciphertext == plaintext {
			t.Error("ciphertext should be different from plaintext")
		}

		// Decrypt
		decrypted, err := service.Decrypt(ciphertext)
		if err != nil {
			t.Fatalf("failed to decrypt: %v", err)
		}

		// Verify decrypted matches original
		if decrypted != plaintext {
			t.Errorf("decrypted = %q, want %q", decrypted, plaintext)
		}
	})

	t.Run("encrypt produces different ciphertext each time", func(t *testing.T) {
		plaintext := "same-password"

		ciphertext1, err := service.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("failed to encrypt first time: %v", err)
		}

		ciphertext2, err := service.Encrypt(plaintext)
		if err != nil {
			t.Fatalf("failed to encrypt second time: %v", err)
		}

		// Due to random nonce, ciphertexts should be different
		if ciphertext1 == ciphertext2 {
			t.Error("encrypting same plaintext twice should produce different ciphertexts")
		}

		// But both should decrypt to the same plaintext
		decrypted1, _ := service.Decrypt(ciphertext1)
		decrypted2, _ := service.Decrypt(ciphertext2)

		if decrypted1 != plaintext || decrypted2 != plaintext {
			t.Error("both ciphertexts should decrypt to original plaintext")
		}
	})

	t.Run("empty plaintext error", func(t *testing.T) {
		_, err := service.Encrypt("")
		if err == nil {
			t.Error("expected error for empty plaintext")
		}
	})

	t.Run("empty ciphertext error", func(t *testing.T) {
		_, err := service.Decrypt("")
		if err == nil {
			t.Error("expected error for empty ciphertext")
		}
	})

	t.Run("invalid ciphertext error", func(t *testing.T) {
		_, err := service.Decrypt("invalid-base64-!@#$")
		if err == nil {
			t.Error("expected error for invalid ciphertext")
		}
	})
}

func TestNewEncryptionService_MissingKey(t *testing.T) {
	// Ensure ENCRYPTION_KEY is not set
	os.Unsetenv("ENCRYPTION_KEY")

	_, err := NewEncryptionService()
	if err == nil {
		t.Error("expected error when ENCRYPTION_KEY is not set")
	}
}

func TestNewEncryptionService_InvalidKeyLength(t *testing.T) {
	// Set a key with wrong length (16 bytes instead of 32)
	key := make([]byte, 16)
	keyStr := base64.StdEncoding.EncodeToString(key)
	os.Setenv("ENCRYPTION_KEY", keyStr)
	defer os.Unsetenv("ENCRYPTION_KEY")

	_, err := NewEncryptionService()
	if err == nil {
		t.Error("expected error for invalid key length")
	}
}
