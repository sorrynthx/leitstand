package ssh

import (
	"fmt"
	"io"
	"leitstand/internal/logger"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// InteractiveSession runs a full PTY interactive SSH session using robust goroutine pipes.
func (c *Client) InteractiveSession(in io.Reader, out io.Writer, errOut io.Writer) error {
	c.mu.Lock()
	if c.rawClient == nil {
		c.mu.Unlock()
		return fmt.Errorf("ssh client is closed")
	}
	client := c.rawClient
	c.mu.Unlock()

	logger.Infof("InteractiveSession: creating new SSH session for %s", c.address)

	session, err := client.NewSession()
	if err != nil {
		logger.Errorf("InteractiveSession: failed to create session: %v", err)
		return fmt.Errorf("failed to create ssh session: %w", err)
	}
	defer session.Close()

	// Setup stdin/stdout/stderr pipes
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		logger.Errorf("InteractiveSession: StdinPipe error: %v", err)
		return fmt.Errorf("failed to open stdin pipe: %w", err)
	}

	stdoutPipe, err := session.StdoutPipe()
	if err != nil {
		logger.Errorf("InteractiveSession: StdoutPipe error: %v", err)
		return fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	stderrPipe, err := session.StderrPipe()
	if err != nil {
		logger.Errorf("InteractiveSession: StderrPipe error: %v", err)
		return fmt.Errorf("failed to open stderr pipe: %w", err)
	}

	// Terminal sizing
	width, height := 120, 30
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		if w, h, err := term.GetSize(fd); err == nil && w > 0 && h > 0 {
			width = w
			height = h
		}
	}

	termType := os.Getenv("TERM")
	if termType == "" {
		termType = "xterm-256color"
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty(termType, height, width, modes); err != nil {
		logger.Errorf("InteractiveSession: RequestPty error: %v", err)
		return fmt.Errorf("request for pseudo terminal failed: %w", err)
	}

	// Switch local terminal to raw mode
	if term.IsTerminal(fd) {
		oldState, err := term.MakeRaw(fd)
		if err == nil {
			defer term.Restore(fd, oldState)
		}
	}

	// Start remote shell
	if err := session.Shell(); err != nil {
		logger.Errorf("InteractiveSession: Shell start error: %v", err)
		return fmt.Errorf("failed to start remote shell: %w", err)
	}
	logger.Infof("InteractiveSession: remote shell started successfully on %s", c.address)

	// Stream I/O via goroutines
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		_, _ = io.Copy(out, stdoutPipe)
	}()

	go func() {
		defer wg.Done()
		_, _ = io.Copy(errOut, stderrPipe)
	}()

	go func() {
		_, _ = io.Copy(stdinPipe, in)
		_ = stdinPipe.Close()
	}()

	// Wait for remote shell to exit (e.g. exit command)
	waitErr := session.Wait()
	wg.Wait()

	logger.Infof("InteractiveSession: session ended for %s (err=%v)", c.address, waitErr)
	return waitErr
}

// ShellRunner implements tea.ExecCommand to be launched seamlessly within Bubbletea.
type ShellRunner struct {
	client *Client
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

// NewShellRunner creates a tea.ExecCommand compatible runner for a given Client.
func NewShellRunner(client *Client) *ShellRunner {
	return &ShellRunner{
		client: client,
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func (r *ShellRunner) SetStdin(in io.Reader) {
	if in != nil {
		r.stdin = in
	}
}

func (r *ShellRunner) SetStdout(out io.Writer) {
	if out != nil {
		r.stdout = out
	}
}

func (r *ShellRunner) SetStderr(errOut io.Writer) {
	if errOut != nil {
		r.stderr = errOut
	}
}

func (r *ShellRunner) Run() error {
	if r.client == nil {
		return fmt.Errorf("no active ssh client")
	}
	return r.client.InteractiveSession(r.stdin, r.stdout, r.stderr)
}

// DemoShellRunner provides a simulated interactive shell for Demo Mode.
type DemoShellRunner struct {
	HostName string
	Address  string
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
}

func NewDemoShellRunner(hostname, address string) *DemoShellRunner {
	return &DemoShellRunner{
		HostName: hostname,
		Address:  address,
		stdin:    os.Stdin,
		stdout:   os.Stdout,
		stderr:   os.Stderr,
	}
}

func (r *DemoShellRunner) SetStdin(in io.Reader)   { r.stdin = in }
func (r *DemoShellRunner) SetStdout(out io.Writer) { r.stdout = out }
func (r *DemoShellRunner) SetStderr(err io.Writer) { r.stderr = err }

func (r *DemoShellRunner) Run() error {
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		oldState, err := term.MakeRaw(fd)
		if err == nil {
			defer term.Restore(fd, oldState)
		}
	}

	screen := term.NewTerminal(struct {
		io.Reader
		io.Writer
	}{r.stdin, r.stdout}, fmt.Sprintf("ubuntu@%s:~$ ", r.HostName))

	screen.Write([]byte(fmt.Sprintf("\r\n\x1b[1;32mWelcome to %s (%s) - Ubuntu 24.04 LTS (Demo Session)\x1b[0m\r\n", r.HostName, r.Address)))
	screen.Write([]byte("Type standard Linux commands (e.g. 'ls', 'pwd', 'uname -a', 'free -m', 'exit').\r\n\r\n"))

	for {
		line, err := screen.ReadLine()
		if err != nil {
			break
		}
		cmd := line
		if cmd == "exit" || cmd == "logout" || cmd == "quit" {
			screen.Write([]byte("Connection to host closed.\r\n"))
			time.Sleep(200 * time.Millisecond)
			break
		} else if cmd == "clear" || cmd == "cls" {
			screen.Write([]byte("\x1b[2J\x1b[H"))
		} else if cmd == "pwd" {
			screen.Write([]byte("/home/ubuntu\r\n"))
		} else if cmd == "ls" || cmd == "ls -la" {
			screen.Write([]byte("total 32\r\ndrwxr-xr-x 4 ubuntu ubuntu 4096 Aug 26 12:00 .\r\ndrwxr-xr-x 3 root   root   4096 Aug 20 09:00 ..\r\n-rw------- 1 ubuntu ubuntu  820 Aug 26 11:30 .bash_history\r\n-rw-r--r-- 1 ubuntu ubuntu  220 Jan  7  2024 .bash_logout\r\n-rw-r--r-- 1 ubuntu ubuntu 3771 Jan  7  2024 .bashrc\r\ndrwx------ 2 ubuntu ubuntu 4096 Aug 20 09:15 .ssh\r\ndrwxr-xr-x 2 ubuntu ubuntu 4096 Aug 26 10:00 app\r\n"))
		} else if cmd == "uname -a" {
			screen.Write([]byte(fmt.Sprintf("Linux %s 6.8.0-40-generic #40-Ubuntu SMP PREEMPT_DYNAMIC x86_64 GNU/Linux\r\n", r.HostName)))
		} else if cmd == "" {
			continue
		} else {
			screen.Write([]byte(fmt.Sprintf("[Demo Shell] Executed: %s\r\n", cmd)))
		}
	}
	return nil
}
