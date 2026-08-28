package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"leitstand/internal/storage"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

// startMockSSHServer launches an in-process SSH server supporting password and public key auth.
func startMockSSHServer(t *testing.T, expectedUser, expectedPass string, expectedPublicKey ssh.PublicKey) (string, func()) {
	// Generate host key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa host key: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}

	serverConfig := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			if expectedPass != "" && conn.User() == expectedUser && string(password) == expectedPass {
				return nil, nil
			}
			return nil, fmt.Errorf("password rejected for user %s", conn.User())
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if expectedPublicKey != nil && conn.User() == expectedUser {
				if string(key.Marshal()) == string(expectedPublicKey.Marshal()) {
					return nil, nil
				}
			}
			return nil, fmt.Errorf("public key rejected for user %s", conn.User())
		},
	}
	serverConfig.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}

	addr := listener.Addr().String()

	stopCh := make(chan struct{})

	go func() {
		for {
			tcpConn, err := listener.Accept()
			if err != nil {
				select {
				case <-stopCh:
					return
				default:
					continue
				}
			}

			go handleMockConn(tcpConn, serverConfig)
		}
	}()

	cleanup := func() {
		close(stopCh)
		listener.Close()
	}

	return addr, cleanup
}

func handleMockConn(nConn net.Conn, config *ssh.ServerConfig) {
	conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		return
	}
	defer conn.Close()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			return
		}

		go func(ch ssh.Channel, in <-chan *ssh.Request) {
			defer ch.Close()
			for req := range in {
				switch req.Type {
				case "exec":
					// Parse command payload
					cmdStr := string(req.Payload[4:])
					req.Reply(true, nil)

					if strings.HasPrefix(cmdStr, "echo ") {
						val := strings.TrimPrefix(cmdStr, "echo ")
						io.WriteString(ch, val+"\n")
					} else {
						io.WriteString(ch, "mock response for: "+cmdStr+"\n")
					}

					// Send exit status 0
					ch.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
					return

				default:
					req.Reply(false, nil)
				}
			}
		}(channel, requests)
	}
}

func TestSSHPoolAndExec(t *testing.T) {
	mockAddr, cleanup := startMockSSHServer(t, "leitstand-user", "secret-pass", nil)
	defer cleanup()

	hostStr, portStr, _ := net.SplitHostPort(mockAddr)
	port, _ := strconv.Atoi(portStr)

	host := &storage.Host{
		ID:        1,
		Name:      "test-box",
		Address:   hostStr,
		Port:      port,
		Username:  "leitstand-user",
		GroupName: "Dev",
	}

	secret := &storage.HostSecret{
		HostID:     1,
		AuthMethod: "password",
	}

	pool := NewPool(5 * time.Second)
	defer pool.CloseAll()

	// 1. Successful connection and pool insertion
	client1, err := pool.GetOrCreate(host, secret, []byte("secret-pass"), nil)
	if err != nil {
		t.Fatalf("failed to connect via pool: %v", err)
	}

	if pool.ActiveCount() != 1 {
		t.Errorf("expected 1 active client in pool, got %d", pool.ActiveCount())
	}

	// 2. Command execution
	stdout, stderr, err := client1.Exec("echo hello-leitstand")
	if err != nil {
		t.Fatalf("exec error: %v, stderr: %s", err, string(stderr))
	}
	if strings.TrimSpace(string(stdout)) != "hello-leitstand" {
		t.Errorf("unexpected stdout: %s", string(stdout))
	}

	// 3. Pool reuse check (should return identical client pointer)
	client2, err := pool.GetOrCreate(host, secret, []byte("secret-pass"), nil)
	if err != nil {
		t.Fatalf("failed to get existing client: %v", err)
	}
	if client1 != client2 {
		t.Error("pool should reuse existing client connection")
	}

	// 4. Failed authentication check
	badHost := &storage.Host{
		ID:        2,
		Name:      "bad-box",
		Address:   hostStr,
		Port:      port,
		Username:  "leitstand-user",
		GroupName: "Dev",
	}
	_, err = pool.GetOrCreate(badHost, secret, []byte("wrong-password"), nil)
	if err == nil {
		t.Error("expected authentication failure for wrong password")
	}

	// 5. Close host
	pool.CloseHost(1)
	if pool.ActiveCount() != 0 {
		t.Errorf("expected 0 active clients after CloseHost, got %d", pool.ActiveCount())
	}
}

func TestSSHPrivateKeyAuth(t *testing.T) {
	// Generate RSA key pair for testing
	rawKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}

	signer, err := ssh.NewSignerFromKey(rawKey)
	if err != nil {
		t.Fatalf("failed to create signer: %v", err)
	}
	publicKey := signer.PublicKey()

	// Encode private key to PEM format
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

	// Connect using GetOrCreateFromPayload
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

	// Check reject wrong user or wrong key
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
