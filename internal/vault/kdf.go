package vault

import (
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2id recommended parameters from Plan.md
	Argon2Time    = 3         // 3 iterations
	Argon2Memory  = 64 * 1024 // 64 MB
	Argon2Threads = 4         // 4 parallel threads
	Argon2KeyLen  = 32        // 256-bit key for AES-256
	SaltLength    = 16        // 128-bit salt
)

// GenerateSalt creates a cryptographically secure random salt.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltLength)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate random salt: %w", err)
	}
	return salt, nil
}

// DeriveKey derives a 256-bit key from a master password and salt using Argon2id.
func DeriveKey(password []byte, salt []byte) []byte {
	return argon2.IDKey(password, salt, Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLen)
}
