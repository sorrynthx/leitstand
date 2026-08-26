package telemetry

import (
	"crypto/rand"
	"crypto/rsa"
	"io"
	"leitstand/internal/ssh"
	"leitstand/internal/storage"
	"leitstand/internal/vault"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	golang_ssh "golang.org/x/crypto/ssh"
)

func TestParseProcStat(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "testdata", "proc_stat.txt"))
	if err != nil {
		t.Fatalf("failed to read testdata/proc_stat.txt: %v", err)
	}

	snap, err := ParseProcStat(string(content))
	if err != nil {
		t.Fatalf("ParseProcStat failed: %v", err)
	}

	if snap.User != 101345 || snap.Idle != 8901234 {
		t.Errorf("unexpected cpu jiffies: user=%d, idle=%d", snap.User, snap.Idle)
	}
}

func TestCalculateCPUPercent(t *testing.T) {
	prev := &CPUTickSnapshot{
		User: 100, Nice: 0, System: 100, Idle: 800,
	}
	curr := &CPUTickSnapshot{
		User: 200, Nice: 0, System: 200, Idle: 1600,
	}

	// Total delta = 1000, Idle delta = 800, Busy delta = 200 => 20%
	pct := CalculateCPUPercent(prev, curr)
	if pct < 19.9 || pct > 20.1 {
		t.Errorf("expected 20.0%% cpu, got %f", pct)
	}
}

func TestParseMeminfo(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "testdata", "meminfo.txt"))
	if err != nil {
		t.Fatalf("failed to read testdata/meminfo.txt: %v", err)
	}

	mem, err := ParseMeminfo(string(content))
	if err != nil {
		t.Fatalf("ParseMeminfo failed: %v", err)
	}

	expectedTotal := uint64(16384000 * 1024)
	if mem.Total != expectedTotal {
		t.Errorf("expected total %d bytes, got %d", expectedTotal, mem.Total)
	}

	expectedAvail := uint64(8192000 * 1024)
	if mem.Available != expectedAvail {
		t.Errorf("expected available %d bytes, got %d", expectedAvail, mem.Available)
	}

	expectedUsed := expectedTotal - expectedAvail
	if mem.Used != expectedUsed {
		t.Errorf("expected used %d bytes, got %d", expectedUsed, mem.Used)
	}
}

func TestParseDF(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "testdata", "df.txt"))
	if err != nil {
		t.Fatalf("failed to read testdata/df.txt: %v", err)
	}

	disk, err := ParseDF(string(content))
	if err != nil {
		t.Fatalf("ParseDF failed: %v", err)
	}

	expectedTotal := uint64(102400000 * 1024)
	if disk.Total != expectedTotal {
		t.Errorf("expected total disk %d, got %d", expectedTotal, disk.Total)
	}
}

func TestParseProcNetDev(t *testing.T) {
	content, err := os.ReadFile(filepath.Join("..", "..", "testdata", "net_dev.txt"))
	if err != nil {
		t.Fatalf("failed to read testdata/net_dev.txt: %v", err)
	}

	netStats, err := ParseProcNetDev(string(content))
	if err != nil {
		t.Fatalf("ParseProcNetDev failed: %v", err)
	}

	// eth0 only (lo ignored) => rx=50000000, tx=25000000
	if netStats.RxBytes != 50000000 || netStats.TxBytes != 25000000 {
		t.Errorf("expected rx=50M, tx=25M, got rx=%d, tx=%d", netStats.RxBytes, netStats.TxBytes)
	}
}

func startTelemetryMockSSHServer(t *testing.T, payload string) (string, func()) {
	key, _ := rsa.GenerateKey(rand.Reader, 2048)
	signer, _ := golang_ssh.NewSignerFromKey(key)

	cfg := &golang_ssh.ServerConfig{
		PasswordCallback: func(conn golang_ssh.ConnMetadata, password []byte) (*golang_ssh.Permissions, error) {
			return nil, nil
		},
	}
	cfg.AddHostKey(signer)

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	stopCh := make(chan struct{})

	go func() {
		for {
			tcpConn, err := l.Accept()
			if err != nil {
				select {
				case <-stopCh:
					return
				default:
					continue
				}
			}

			go func(c net.Conn) {
				_, chans, reqs, err := golang_ssh.NewServerConn(c, cfg)
				if err != nil {
					return
				}
				go golang_ssh.DiscardRequests(reqs)

				for newChannel := range chans {
					ch, in, _ := newChannel.Accept()
					go func(ch golang_ssh.Channel, in <-chan *golang_ssh.Request) {
						defer ch.Close()
						for req := range in {
							if req.Type == "exec" {
								req.Reply(true, nil)
								io.WriteString(ch, payload)
								ch.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
								return
							}
						}
					}(ch, in)
				}
			}(tcpConn)
		}
	}()

	return l.Addr().String(), func() {
		close(stopCh)
		l.Close()
	}
}

func TestCollectorWithMockSSH(t *testing.T) {
	procStat, _ := os.ReadFile(filepath.Join("..", "..", "testdata", "proc_stat.txt"))
	meminfo, _ := os.ReadFile(filepath.Join("..", "..", "testdata", "meminfo.txt"))
	df, _ := os.ReadFile(filepath.Join("..", "..", "testdata", "df.txt"))
	netDev, _ := os.ReadFile(filepath.Join("..", "..", "testdata", "net_dev.txt"))

	mockOutput := string(procStat) + "\n" + MetricSplitDelimiter + "\n" +
		string(meminfo) + "\n" + MetricSplitDelimiter + "\n" +
		string(df) + "\n" + MetricSplitDelimiter + "\n" +
		string(netDev) + "\n" + MetricSplitDelimiter + "\n" +
		"Linux test-node 6.8.0 x86_64\nPRETTY_NAME=\"Ubuntu 22.04.4 LTS\"\nup 5 days, 12 hours\n4"

	addr, cleanup := startTelemetryMockSSHServer(t, mockOutput)
	defer cleanup()

	hostStr, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	// Set up storage & vault
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "collector_test.db")
	store, err := storage.Open(dbPath)
	if err != nil {
		t.Fatalf("storage open failed: %v", err)
	}
	defer store.Close()

	v := vault.New()
	_ = store.InitVault(v, "masterpassword")

	// Create host
	h := &storage.Host{
		Name:      "test-target",
		Address:   hostStr,
		Port:      port,
		Username:  "root",
		GroupName: "Test",
	}
	hostID, err := store.CreateHost(h)
	if err != nil {
		t.Fatalf("create host failed: %v", err)
	}

	// Save secret
	nonce, ct, _ := v.Encrypt([]byte("dummy-pass"))
	_ = store.SaveHostSecret(&storage.HostSecret{
		HostID:     hostID,
		AuthMethod: "password",
		Nonce:      nonce,
		Ciphertext: ct,
	})

	pool := ssh.NewPool(5 * time.Second)
	defer pool.CloseAll()

	collector := NewCollector(store, pool, v)

	// 1st collection (establishes baseline tick)
	rec1, err := collector.CollectHost(h)
	if err != nil {
		t.Fatalf("collect host 1 failed: %v", err)
	}
	if rec1.MemoryTotal == 0 || rec1.DiskTotal == 0 {
		t.Errorf("expected non-zero memory and disk totals: %+v", rec1)
	}

	// Verify persistence in SQLite
	latest, err := store.GetLatestMetric(hostID)
	if err != nil || latest == nil {
		t.Fatalf("failed to retrieve latest metric from DB: %v", err)
	}
	if latest.MemoryTotal != rec1.MemoryTotal {
		t.Errorf("persisted metric mismatch")
	}
}
