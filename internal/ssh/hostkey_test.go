package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/crypto/ssh"
)

func generateTestHostKey(t *testing.T) ssh.PublicKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}
	pub, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("failed to create ssh public key: %v", err)
	}
	return pub
}

func TestHostKeyVerification_TOFU(t *testing.T) {
	tempDir := t.TempDir()
	khPath := filepath.Join(tempDir, "known_hosts")

	key1 := generateTestHostKey(t)
	key2 := generateTestHostKey(t)

	callback, err := GetHostKeyCallback(khPath, "accept-new")
	if err != nil {
		t.Fatalf("GetHostKeyCallback failed: %v", err)
	}

	dummyAddr, _ := net.ResolveTCPAddr("tcp", "192.168.1.100:22")

	// 1. First connection: Host is new. Policy "accept-new" should register it.
	err = callback("192.168.1.100:22", dummyAddr, key1)
	if err != nil {
		t.Fatalf("First connection with accept-new should succeed, got: %v", err)
	}

	// Verify known_hosts was written
	content, err := os.ReadFile(khPath)
	if err != nil || len(content) == 0 {
		t.Fatalf("known_hosts file should not be empty")
	}

	// 2. Second connection with SAME key: Should verify successfully.
	err = callback("192.168.1.100:22", dummyAddr, key1)
	if err != nil {
		t.Fatalf("Subsequent connection with matching key should succeed, got: %v", err)
	}

	// 3. Third connection with DIFFERENT key (MITM simulation): MUST FAIL!
	err = callback("192.168.1.100:22", dummyAddr, key2)
	if err == nil {
		t.Fatalf("Connection with changed key MUST fail (MITM attack defense)")
	}
	if !strings.Contains(err.Error(), "MISMATCH DETECTED") {
		t.Errorf("Error should mention MISMATCH DETECTED, got: %v", err)
	}
}

func TestHostKeyVerification_Insecure(t *testing.T) {
	callback, err := GetHostKeyCallback("", "insecure")
	if err != nil {
		t.Fatalf("GetHostKeyCallback failed: %v", err)
	}

	key := generateTestHostKey(t)
	err = callback("any-host:22", nil, key)
	if err != nil {
		t.Fatalf("Insecure policy should always succeed, got: %v", err)
	}
}
