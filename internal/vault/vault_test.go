package vault

import (
	"bytes"
	"testing"
)

func TestArgon2idAndAESGCM(t *testing.T) {
	password := []byte("SuperSecretMasterPass123!")
	salt, err := GenerateSalt()
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}

	key := DeriveKey(password, salt)
	if len(key) != 32 {
		t.Fatalf("expected 32-byte key, got %d", len(key))
	}

	plaintext := []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC...")

	// Encrypt
	nonce1, ciphertext1, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Encrypt again (should produce different nonce and ciphertext)
	nonce2, ciphertext2, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	if bytes.Equal(nonce1, nonce2) {
		t.Error("nonces must be unique per encryption")
	}
	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("ciphertexts must differ due to unique nonces")
	}

	// Decrypt
	decrypted, err := Decrypt(nonce1, ciphertext1, key)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Errorf("decrypted text mismatch: got %s", string(decrypted))
	}

	// Decrypt with wrong key (must fail with auth tag error)
	wrongKey := DeriveKey([]byte("WrongPassword"), salt)
	_, err = Decrypt(nonce1, ciphertext1, wrongKey)
	if err != ErrInvalidCiphertext {
		t.Errorf("expected ErrInvalidCiphertext, got: %v", err)
	}

	// Decrypt with corrupted ciphertext (must fail)
	corrupted := append([]byte(nil), ciphertext1...)
	corrupted[0] ^= 0xFF
	_, err = Decrypt(nonce1, corrupted, key)
	if err != ErrInvalidCiphertext {
		t.Errorf("expected ErrInvalidCiphertext for corrupted payload, got: %v", err)
	}
}

func TestVaultLifecycle(t *testing.T) {
	v := New()
	if v.IsUnlocked() {
		t.Fatal("new vault should be locked")
	}

	_, _, err := v.Encrypt([]byte("secret"))
	if err != ErrVaultLocked {
		t.Errorf("expected ErrVaultLocked, got: %v", err)
	}

	salt, _ := GenerateSalt()
	if err := v.Unlock("master-pass", salt); err != nil {
		t.Fatalf("unlock failed: %v", err)
	}
	if !v.IsUnlocked() {
		t.Fatal("vault should be unlocked")
	}

	nonce, ct, err := v.Encrypt([]byte("my-secret-ssh-key"))
	if err != nil {
		t.Fatalf("vault encrypt failed: %v", err)
	}

	pt, err := v.Decrypt(nonce, ct)
	if err != nil {
		t.Fatalf("vault decrypt failed: %v", err)
	}
	if string(pt) != "my-secret-ssh-key" {
		t.Errorf("mismatch: %s", string(pt))
	}

	// Lock
	v.Lock()
	if v.IsUnlocked() {
		t.Fatal("vault should be locked after Lock()")
	}

	_, _, err = v.Encrypt([]byte("secret"))
	if err != ErrVaultLocked {
		t.Errorf("expected ErrVaultLocked after Lock(), got: %v", err)
	}
}
