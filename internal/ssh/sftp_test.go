package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"leitstand/internal/storage"
	"golang.org/x/crypto/ssh"
)

func TestSSHPrivateKeyAuth(t *testing.T) {
	rawKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(rawKey)
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}
	publicKey := signer.PublicKey()

	keyDER := x509.MarshalPKCS1PrivateKey(rawKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyDER,
	})

	mockAddr, cleanup := startMockSSHServer(t, "key-user", "", publicKey)
	defer cleanup()

	hostStr, portStr, _ := net.SplitHostPort(mockAddr)
	port, _ := strconv.Atoi(portStr)

	host := &storage.Host{
		ID:        10,
		Name:      "key-server",
		Address:   hostStr,
		Port:      port,
		Username:  "key-user",
		GroupName: "Cloud",
	}

	secret := &storage.HostSecret{
		HostID:     10,
		AuthMethod: "private_key",
	}

	payload := &storage.SecretPayload{
		PrivateKey: string(keyPEM),
	}

	pool := NewPool(5 * time.Second)
	defer pool.CloseAll()

	client, err := pool.GetOrCreateFromPayload(host, secret, payload)
	if err != nil {
		t.Fatalf("failed to connect via private key: %v", err)
	}

	stdout, stderr, err := client.Exec("echo key-auth-success")
	if err != nil {
		t.Fatalf("exec error: %v (stderr: %s)", err, string(stderr))
	}
	if strings.TrimSpace(string(stdout)) != "key-auth-success" {
		t.Errorf("unexpected output: %s", string(stdout))
	}

	wrongHost := &storage.Host{
		ID:        11,
		Name:      "wrong-server",
		Address:   hostStr,
		Port:      port,
		Username:  "wrong-user",
		GroupName: "Cloud",
	}
	_, err = pool.GetOrCreateFromPayload(wrongHost, secret, payload)
	if err == nil {
		t.Error("expected failure for wrong username with private key")
	}
}
