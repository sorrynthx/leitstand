package storage

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"leitstand/internal/vault"
)

const verificationSentinel = "leitstand:vault:verification:ok"

var (
	ErrVaultNotInitialized = errors.New("vault is not initialized")
	ErrVaultAlreadyInit    = errors.New("vault is already initialized")
)

// VaultMeta holds the persistent vault salt and verification ciphertext.
type VaultMeta struct {
	Salt                   []byte
	VerificationNonce      []byte
	VerificationCiphertext []byte
}

// IsVaultInitialized checks if the master password and salt have been configured.
func (s *Storage) IsVaultInitialized() (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM vault_meta WHERE id = 1;").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check vault initialization: %w", err)
	}
	return count > 0, nil
}

// InitVault initializes the vault with a new salt and derives keys to encrypt the sentinel.
func (s *Storage) InitVault(v *vault.Vault, masterPassword string) error {
	isInit, err := s.IsVaultInitialized()
	if err != nil {
		return err
	}
	if isInit {
		return ErrVaultAlreadyInit
	}

	salt, err := vault.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	if err := v.Unlock(masterPassword, salt); err != nil {
		return fmt.Errorf("failed to unlock new vault: %w", err)
	}

	nonce, ciphertext, err := v.Encrypt([]byte(verificationSentinel))
	if err != nil {
		return fmt.Errorf("failed to encrypt verification sentinel: %w", err)
	}

	query := `
		INSERT INTO vault_meta (id, salt, verification_nonce, verification_ciphertext)
		VALUES (1, ?, ?, ?);
	`
	_, err = s.db.Exec(query, salt, nonce, ciphertext)
	if err != nil {
		return fmt.Errorf("failed to persist vault metadata: %w", err)
	}

	return nil
}

// UnlockVault reads the salt and attempts to decrypt the sentinel to unlock the Vault.
func (s *Storage) UnlockVault(v *vault.Vault, masterPassword string) error {
	var meta VaultMeta
	query := `
		SELECT salt, verification_nonce, verification_ciphertext
		FROM vault_meta
		WHERE id = 1;
	`
	err := s.db.QueryRow(query).Scan(&meta.Salt, &meta.VerificationNonce, &meta.VerificationCiphertext)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrVaultNotInitialized
		}
		return fmt.Errorf("failed to read vault metadata: %w", err)
	}

	if err := v.Unlock(masterPassword, meta.Salt); err != nil {
		return err
	}

	// Verify by decrypting sentinel
	decrypted, err := v.Decrypt(meta.VerificationNonce, meta.VerificationCiphertext)
	if err != nil || !bytes.Equal(decrypted, []byte(verificationSentinel)) {
		v.Lock() // Clear master key from memory immediately on failure
		return vault.ErrInvalidPassword
	}

	return nil
}

// RekeyVault decrypts all host secrets using oldVault and re-encrypts them with newPassword into newVault.
func (s *Storage) RekeyVault(oldVault *vault.Vault, newVault *vault.Vault, newPassword string) error {
	if !oldVault.IsUnlocked() {
		return vault.ErrVaultLocked
	}

	hosts, err := s.ListHosts()
	if err != nil {
		return err
	}

	type DecryptedSecret struct {
		HostID     int64
		AuthMethod string
		RawSecret  []byte
	}
	var decryptedSecrets []DecryptedSecret

	for _, h := range hosts {
		sec, err := s.GetHostSecret(h.ID)
		if err != nil || sec == nil {
			continue
		}
		raw, err := oldVault.Decrypt(sec.Nonce, sec.Ciphertext)
		if err != nil {
			return fmt.Errorf("failed to decrypt secret for host %d: %w", h.ID, err)
		}
		decryptedSecrets = append(decryptedSecrets, DecryptedSecret{
			HostID:     sec.HostID,
			AuthMethod: sec.AuthMethod,
			RawSecret:  raw,
		})
	}
	defer func() {
		for _, ds := range decryptedSecrets {
			vault.ZeroBytes(ds.RawSecret)
		}
	}()

	newSalt, err := vault.GenerateSalt()
	if err != nil {
		return fmt.Errorf("failed to generate new salt: %w", err)
	}

	if err := newVault.Unlock(newPassword, newSalt); err != nil {
		return fmt.Errorf("failed to unlock new vault: %w", err)
	}

	newNonce, newCiphertext, err := newVault.Encrypt([]byte(verificationSentinel))
	if err != nil {
		return fmt.Errorf("failed to encrypt new sentinel: %w", err)
	}

	_, err = s.db.Exec("UPDATE vault_meta SET salt = ?, verification_nonce = ?, verification_ciphertext = ? WHERE id = 1;", newSalt, newNonce, newCiphertext)
	if err != nil {
		return fmt.Errorf("failed to update vault meta: %w", err)
	}

	for _, ds := range decryptedSecrets {
		nonce, ciphertext, err := newVault.Encrypt(ds.RawSecret)
		if err != nil {
			return fmt.Errorf("failed to re-encrypt secret for host %d: %w", ds.HostID, err)
		}
		err = s.SaveHostSecret(&HostSecret{
			HostID:     ds.HostID,
			AuthMethod: ds.AuthMethod,
			Nonce:      nonce,
			Ciphertext: ciphertext,
		})
		if err != nil {
			return fmt.Errorf("failed to save re-encrypted secret: %w", err)
		}
	}

	return nil
}
