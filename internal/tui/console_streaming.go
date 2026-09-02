package tui

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"leitstand/internal/storage"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	gossh "golang.org/x/crypto/ssh"
)

func IsStreamingCommand(cmd string) bool {
	trimmed := strings.TrimSpace(strings.ToLower(cmd))
	if strings.Contains(trimmed, "logs -f") || strings.Contains(trimmed, "logs --follow") || strings.Contains(trimmed, "tail -f") || strings.Contains(trimmed, "tail -F") {
		return true
	}
	streamingPrefixes := []string{
		"top", "htop", "btop", "iotop", "iftop",
		"watch ", "tail -n", "ping ", "journalctl -f", "journalctl -u",
	}
	for _, p := range streamingPrefixes {
		if trimmed == p || strings.HasPrefix(trimmed, p) {
			return true
		}
	}
	return false
}

type StreamChunkMsg struct {
	HostID  int64
	TabID   string
	Chunk   string
	msgChan <-chan tea.Msg
}

type StreamFinishedMsg struct {
	HostID int64
	TabID  string
	Err    error
}

func (m *Model) execStreamingCmdInTab(host *storage.Host, tab *ConsoleTab, cmdText string) tea.Cmd {
	hostID := host.ID
	tabID := tab.ID
	cwd := tab.CWD
	if cwd == "" {
		cwd = "~"
	}

	tab.IsStreaming = true
	tab.StreamCmd = cmdText

	ctx, cancel := context.WithCancel(context.Background())
	tab.StreamCancel = cancel

	trimmed := strings.TrimSpace(cmdText)
	if trimmed == "top" || strings.HasPrefix(trimmed, "top ") || strings.HasPrefix(trimmed, "htop") || strings.HasPrefix(trimmed, "btop") || strings.HasPrefix(trimmed, "watch") {
		tab.IsScreenApp = true
	} else {
		tab.IsScreenApp = false
	}

	msgChan := make(chan tea.Msg, 50)

	go func() {
		defer close(msgChan)

		if m.isDemo {
			for i := 1; i <= 5; i++ {
				select {
				case <-ctx.Done():
					return
				case <-time.After(500 * time.Millisecond):
					chunk := fmt.Sprintf("[%s Streaming Demo] Tick #%d - %s\nCPU: 25.4%% | MEM: 4.2GB / 16GB | DISK: 45%%", cmdText, i, time.Now().Format("15:04:05"))
					msgChan <- StreamChunkMsg{HostID: hostID, TabID: tabID, Chunk: chunk, msgChan: msgChan}
				}
			}
			msgChan <- StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: nil}
			return
		}

		client, err := m.getSSHClient(host)
		if err != nil {
			msgChan <- StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: err}
			return
		}

		rawClient := client.RawClient()
		if rawClient == nil {
			msgChan <- StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: fmt.Errorf("ssh connection not available")}
			return
		}

		session, err := rawClient.NewSession()
		if err != nil {
			msgChan <- StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: err}
			return
		}
		defer session.Close()

		stdoutPipe, err := session.StdoutPipe()
		if err != nil {
			msgChan <- StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: err}
			return
		}

		stderrPipe, err := session.StderrPipe()
		if err != nil {
			msgChan <- StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: err}
			return
		}

		var safeCdTarget string
		if cwd == "~" || cwd == "" {
			safeCdTarget = "$HOME"
		} else {
			safeCdTarget = fmt.Sprintf("\"%s\"", cwd)
		}

		actualCmd := cmdText
		if tab.IsScreenApp {
			trimmed := strings.TrimSpace(cmdText)
			if trimmed == "top" || strings.HasPrefix(trimmed, "top ") {
				actualCmd = "LINES=120 COLUMNS=512 top -b -d 1 -c -w 512"
			} else {
				actualCmd = fmt.Sprintf("LINES=120 COLUMNS=512 %s", trimmed)
			}
		}

		wrappedCmd := fmt.Sprintf("cd %s 2>/dev/null || cd \"$HOME\" ; %s", safeCdTarget, actualCmd)

		if err := session.Start(wrappedCmd); err != nil {
			msgChan <- StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: err}
			return
		}

		go func() {
			<-ctx.Done()
			_ = session.Signal(gossh.SIGKILL)
			_ = session.Close()
		}()

		reader := io.MultiReader(stdoutPipe, stderrPipe)
		scanner := bufio.NewScanner(reader)

		if tab.IsScreenApp {
			var frameLines []string
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "top - ") && len(frameLines) > 0 {
					frame := strings.Join(frameLines, "\n")
					frameLines = []string{line}
					select {
					case <-ctx.Done():
						return
					case msgChan <- StreamChunkMsg{HostID: hostID, TabID: tabID, Chunk: frame, msgChan: msgChan}:
					}
				} else {
					frameLines = append(frameLines, line)
				}
			}
			if len(frameLines) > 0 {
				msgChan <- StreamChunkMsg{HostID: hostID, TabID: tabID, Chunk: strings.Join(frameLines, "\n"), msgChan: msgChan}
			}
		} else {
			for scanner.Scan() {
				line := scanner.Text()
				select {
				case <-ctx.Done():
					break
				case msgChan <- StreamChunkMsg{HostID: hostID, TabID: tabID, Chunk: line, msgChan: msgChan}:
				}
			}
		}

		_ = session.Wait()
		msgChan <- StreamFinishedMsg{HostID: hostID, TabID: tabID, Err: nil}
	}()

	return listenStreamCmd(msgChan)
}
