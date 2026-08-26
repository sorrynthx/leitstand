package vault

import (
	"errors"
	"sync"
)

var (
	ErrVaultLocked      = errors.New("vault is locked: master password required")
	ErrVaultAlreadyOpen = errors.New("vault is already unlocked")
	ErrInvalidPassword  = errors.New("invalid master password")
)

// Vault manages the in-memory cryptographic key lifecycle.
type Vault struct {
	mu         sync.RWMutex
	masterKey  []byte
	isUnlocked bool
}

// New creates an initially locked Vault.
func New() *Vault {
	return &Vault{}
}

// Unlock derives the master key using Argon2id and transitions the vault to unlocked state.
func (v *Vault) Unlock(password string, salt []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.isUnlocked {
		return ErrVaultAlreadyOpen
	}

	passBytes := []byte(password)
	defer ZeroBytes(passBytes)

	v.masterKey = DeriveKey(passBytes, salt)
	v.isUnlocked = true
	return nil
}

// Lock clears the master key from memory and locks the vault.
func (v *Vault) Lock() {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.masterKey != nil {
		ZeroBytes(v.masterKey)
		v.masterKey = nil
	}
	v.isUnlocked = false
}

// IsUnlocked checks if the vault is currently accessible.
func (v *Vault) IsUnlocked() bool {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.isUnlocked
}

// Encrypt encrypts sensitive data using the active master key.
func (v *Vault) Encrypt(plaintext []byte) (nonce []byte, ciphertext []byte, err error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if !v.isUnlocked || v.masterKey == nil {
		return nil, nil, ErrVaultLocked
	}

	return Encrypt(plaintext, v.masterKey)
}

// Decrypt decrypts encrypted data using the active master key.
func (v *Vault) Decrypt(nonce []byte, ciphertext []byte) ([]byte, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	if !v.isUnlocked || v.masterKey == nil {
		return nil, ErrVaultLocked
	}

	return Decrypt(nonce, ciphertext, v.masterKey)
}
