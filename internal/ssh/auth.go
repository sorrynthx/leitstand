package ssh

import (
	"errors"
	"fmt"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

var (
	ErrUnsupportedAuthMethod = errors.New("unsupported authentication method")
	ErrEmptyCredential       = errors.New("empty credential provided")
)

// BuildClientConfig creates an ssh.ClientConfig based on the host and decrypted secret.
func BuildClientConfig(username string, authMethod string, secretPayload []byte, passphrase []byte, timeout time.Duration) (*ssh.ClientConfig, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	var authMethods []ssh.AuthMethod

	switch authMethod {
	case "password":
		if len(secretPayload) == 0 {
			return nil, ErrEmptyCredential
		}
		authMethods = append(authMethods, ssh.Password(string(secretPayload)))

	case "private_key":
		if len(secretPayload) == 0 {
			return nil, ErrEmptyCredential
		}
		var signer ssh.Signer
		var err error
		if len(passphrase) > 0 {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(secretPayload, passphrase)
		} else {
			signer, err = ssh.ParsePrivateKey(secretPayload)
		}
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))

	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedAuthMethod, authMethod)
	}

	return &ssh.ClientConfig{
		User:            username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: support known_hosts in future phases
		Timeout:         timeout,
		HostKeyAlgorithms: []string{
			ssh.KeyAlgoED25519,
			ssh.KeyAlgoECDSA256,
			ssh.KeyAlgoECDSA384,
			ssh.KeyAlgoECDSA521,
			ssh.KeyAlgoRSASHA256,
			ssh.KeyAlgoRSASHA512,
			ssh.KeyAlgoRSA,
		},
	}, nil
}

// BuildAddress formats the host and port into standard "host:port" address.
func BuildAddress(host string, port int) string {
	if port <= 0 {
		port = 22
	}
	return net.JoinHostPort(host, fmt.Sprintf("%d", port))
}
