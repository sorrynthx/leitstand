package tui

import (
	"context"
	"leitstand/internal/ssh"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type ActivePanel int

const (
	PanelLocal ActivePanel = iota
	PanelRemote
)

type ActivePrompt int

const (
	PromptNone ActivePrompt = iota
	PromptMkdir
	PromptTouch
	PromptRename
	PromptDeleteConfirm
	PromptQuickCmd
	PromptExitConfirm
)

type SortField int

const (
	SortByName SortField = iota
	SortBySize
	SortByModTime
)

type LocalFileListMsg struct {
	HostID  int64
	Path    string
	OldPath string
	Items   []*ssh.FileItem
	Err     error
}

type NavigateRemoteMsg struct {
	HostID  int64
	NewPath string
	OldPath string
}

type FileTransferProgressMsg struct {
	HostID       int64
	FileName     string
	FileIndex    int
	FileTotal    int
	CurrentBytes int64
	TotalBytes   int64
	BytesPerSec  float64
	IsDone       bool
	IsMove       bool
	Err          error
	msgChan      <-chan tea.Msg
}

type TransferActionMsg struct {
	HostID      int64
	IsUpload    bool
	IsMove      bool
	IsSameHost  bool
	IsLocalOp   bool
	SrcPaths    []string
	DestDirPath string
}

type FileOpActionMsg struct {
	HostID      int64
	IsLocal     bool
	OpType      string
	DirPath     string
	OldName     string
	NewName     string
	TargetPaths []string
}

type FileManagerQuickCmdMsg struct {
	HostID  int64
	IsLocal bool
	DirPath string
	Command string
}

type FileManagerQuickCmdResultMsg struct {
	HostID  int64
	IsLocal bool
	Command string
	OldCWD  string
	NewCWD  string
	Output  string
	Err     error
}

type FileManagerRefreshMsg struct {
	HostID        int64
	RefreshRemote bool
}

type FileManagerModal struct {
	HostID           int64
	HostName         string
	LocalPath        string
	RemotePath       string
	LocalItems       []*ssh.FileItem
	RemoteItems      []*ssh.FileItem
	LocalCursor      int
	RemoteCursor     int
	LocalSelected    map[string]bool
	RemoteSelected   map[string]bool
	ActivePanel      ActivePanel
	ActivePrompt     ActivePrompt
	FocusLocal       bool
	SubInput         textinput.Model
	QuickCmdInput    textinput.Model
	StatusMessage    string
	ShowHidden       bool
	ShowRunbook      bool
	LocalFilter      string
	ClipboardPaths   []string
	ClipboardIsCut   bool
	ClipboardIsLocal bool
	LocalSort        SortField
	RemoteSort       SortField
	LocalSortAsc     bool
	RemoteSortAsc            bool
	IsTransferring           bool
	TransferIsUpload         bool
	CurrentFileName          string
	FileIndex                int
	FileTotal                int
	CurrentBytes             int64
	CurrentTotal             int64
	BytesPerSec              float64
	TransferDoneMsg          string
	TransferDoneTime         time.Time
	TransferCancel           context.CancelFunc
	ShowTransferCancelPrompt bool
	IsTransferBackground     bool
	IsTransferCanceled       bool
	ShowCmdOutput            bool
	CmdOutputTitle   string
	CmdOutputContent string
	CmdOutputScroll  int
	RunbookCursor    int
	Width            int
	Height           int
}

func NewFileManagerModal(hostID int64, hostName, localInit, remoteInit string) *FileManagerModal {
	if localInit == "" {
		home, err := osUserHomeDir()
		if err == nil && home != "" {
			localInit = home
		} else {
			localInit = "."
		}
	}
	absLocal, err := filepath.Abs(localInit)
	if err == nil {
		localInit = absLocal
	}

	if remoteInit == "" {
		remoteInit = "."
	}

	subIn := textinput.New()
	subIn.Width = 35

	cmdIn := textinput.New()
	cmdIn.Placeholder = "Enter inline command (e.g. unzip file.zip, chmod +x app.sh)..."
	cmdIn.Width = 60

	return &FileManagerModal{
		HostID:         hostID,
		HostName:       hostName,
		LocalPath:      localInit,
		RemotePath:     remoteInit,
		LocalSelected:  make(map[string]bool),
		RemoteSelected: make(map[string]bool),
		ActivePanel:    PanelRemote,
		FocusLocal:     false,
		ActivePrompt:   PromptNone,
		SubInput:       subIn,
		QuickCmdInput:  cmdIn,
		ShowHidden:     false,
		LocalSort:      SortByName,
		RemoteSort:     SortByName,
		LocalSortAsc:   true,
		RemoteSortAsc:  true,
	}
}

func osUserHomeDir() (string, error) {
	return filepath.Abs(".")
}
