package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

const (
	NonceLength = 12 // Standard 96-bit nonce for GCM
)

var (
	ErrInvalidCiphertext = errors.New("invalid ciphertext or authentication tag mismatch")
	ErrInvalidKeySize    = errors.New("invalid key size: AES-256 requires exactly 32 bytes")
)

// Encrypt encrypts plaintext using AES-256-GCM with a fresh random nonce.
func Encrypt(plaintext []byte, key []byte) (nonce []byte, ciphertext []byte, err error) {
	if len(key) != 32 {
		return nil, nil, ErrInvalidKeySize
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate random nonce: %w", err)
	}

	// Seal appends ciphertext and GCM authentication tag
	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}

// Decrypt authenticates and decrypts ciphertext using AES-256-GCM.
func Decrypt(nonce []byte, ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKeySize
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	if len(nonce) != gcm.NonceSize() {
		return nil, errors.New("invalid nonce length")
	}

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	return plaintext, nil
}

// ZeroBytes overwrites a byte slice with zeros to clear sensitive data from memory.
func ZeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
