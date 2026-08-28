package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {
	// Generate host key for the server
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate RSA host key: %v", err)
	}
	signer, err := ssh.NewSignerFromKey(key)
	if err != nil {
		log.Fatalf("Failed to create signer: %v", err)
	}

	config := &ssh.ServerConfig{
		// Allow any valid public key or password for convenient local testing
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			log.Printf("[MockSSH] 🔑 User '%s' logged in via Password", c.User())
			return nil, nil
		},
		PublicKeyCallback: func(c ssh.ConnMetadata, pubKey ssh.PublicKey) (*ssh.Permissions, error) {
			log.Printf("[MockSSH] 🔑 User '%s' authenticated successfully via Private Key (%s)!", c.User(), pubKey.Type())
			return nil, nil
		},
	}
	config.AddHostKey(signer)

	port := "2222"
	listener, err := net.Listen("tcp", "127.0.0.1:"+port)
	if err != nil {
		log.Fatalf("Failed to listen on 127.0.0.1:%s: %v", port, err)
	}
	defer listener.Close()

	fmt.Println("==========================================================")
	fmt.Printf("🚀 Leitstand Local Test SSH Server running on 127.0.0.1:%s\n", port)
	fmt.Println("👉 You can connect from Leitstand using Private Key (test_key)!")
	fmt.Println("==========================================================")

	for {
		nConn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go handleConn(nConn, config)
	}
}

func handleConn(nConn net.Conn, config *ssh.ServerConfig) {
	conn, chans, reqs, err := ssh.NewServerConn(nConn, config)
	if err != nil {
		log.Printf("[MockSSH] Handshake failed: %v", err)
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
			log.Printf("Could not accept channel: %v", err)
			return
		}

		go func(ch ssh.Channel, in <-chan *ssh.Request) {
			defer ch.Close()
			for req := range in {
				switch req.Type {
				case "exec":
					cmdStr := string(req.Payload[4:])
					req.Reply(true, nil)
					handleExec(ch, cmdStr, conn.User())
					ch.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
					return

				default:
					req.Reply(false, nil)
				}
			}
		}(channel, requests)
	}
}

var tickCounter int64

func handleExec(ch ssh.Channel, cmdStr string, user string) {
	tickCounter++
	now := time.Now()

	// If Leitstand telemetry metric script is requested:
	if strings.Contains(cmdStr, "/proc/stat") || strings.Contains(cmdStr, "LEITSTAND_SPLIT") {
		sinVal := (math.Sin(float64(tickCounter)*0.3) + 1.0) / 2.0 // 0.0 ~ 1.0
		userTicks := int64(1000000 + sinVal*300000)
		idleTicks := int64(5000000 - sinVal*300000)

		response := fmt.Sprintf(`cpu  %d 20000 30000 %d 15000 0 5000 0 0 0
===LEITSTAND_SPLIT===
MemTotal:       16384000 kB
MemFree:         4194304 kB
MemAvailable:    8388608 kB
Buffers:          524288 kB
Cached:          3670016 kB
===LEITSTAND_SPLIT===
Filesystem     1K-blocks      Used Available Use%% Mounted on
/dev/root      102400000  45000000  57400000  44%% /
===LEITSTAND_SPLIT===
Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    eth0: %d 10000    0    0    0     0          0         0 %d 10000    0    0    0     0       0          0
===LEITSTAND_SPLIT===
Linux 6.8.0-test-ed25519 x86_64
PRETTY_NAME="Ubuntu 24.04 LTS (Local Test Server)"
up 5 days, 12:34
8
`, userTicks, idleTicks, tickCounter*102400, tickCounter*40960)

		io.WriteString(ch, response)
		return
	}

	// Normal command executions
	trimmed := strings.TrimSpace(cmdStr)
	log.Printf("[MockSSH] 💻 Executing: %s", trimmed)

	if strings.Contains(trimmed, "whoami") {
		io.WriteString(ch, user+"\n")
	} else if strings.Contains(trimmed, "uname") {
		io.WriteString(ch, "Linux leitstand-test-box 6.8.0-test-ed25519 x86_64 GNU/Linux\n")
	} else if strings.Contains(trimmed, "pwd") {
		io.WriteString(ch, "/home/"+user+"\n")
	} else if strings.Contains(trimmed, "ls") {
		io.WriteString(ch, "app.py\nconfig.yaml\nlogs/\nREADME.md\ndocker-compose.yml\n")
	} else if strings.Contains(trimmed, "docker ps") {
		io.WriteString(ch, "CONTAINER ID   IMAGE          COMMAND                  CREATED         STATUS         PORTS                  NAMES\n"+
			"a1b2c3d4e5f6   nginx:alpine   \"/docker-entrypoint.…\"   2 days ago      Up 2 days      0.0.0.0:80->80/tcp     web-proxy\n"+
			"9f8e7d6c5b4a   redis:alpine   \"docker-entrypoint.s…\"   5 days ago      Up 5 days      0.0.0.0:6379->6379/tcp redis-cache\n")
	} else {
		io.WriteString(ch, fmt.Sprintf("✅ [Local Server Result] Successfully executed: %s (Time: %s)\n", trimmed, now.Format("15:04:05")))
	}
}
