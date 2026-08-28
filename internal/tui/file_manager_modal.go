package tui

import (
	"fmt"
	"leitstand/internal/i18n"
	"leitstand/internal/ssh"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

// TransferActionMsg triggers a batch upload, download, or cross-pane move.
type TransferActionMsg struct {
	HostID      int64
	HostName    string
	IsUpload    bool
	IsMove      bool
	SrcPaths    []string
	DestDirPath string
}

// FileTransferProgressMsg updates the live progress bar.
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
	msgChan      chan tea.Msg
}

// LocalFileListMsg carries local directory listing results with fallback path.
type LocalFileListMsg struct {
	HostID  int64
	Path    string
	OldPath string
	Items   []*ssh.FileItem
	Err     error
}

// NavigateRemoteMsg requests remote directory navigation.
type NavigateRemoteMsg struct {
	HostID  int64
	NewPath string
	OldPath string
}

// FileManagerRefreshMsg tells the file manager to refresh directory contents.
type FileManagerRefreshMsg struct {
	HostID        int64
	RefreshLocal  bool
	RefreshRemote bool
}

// FileOpActionMsg triggers file system operations (mkdir, touch, rename, delete, move, copy_same).
type FileOpActionMsg struct {
	HostID      int64
	IsLocal     bool
	OpType      string // "mkdir", "touch", "rename", "delete", "move", "copy_same"
	DirPath     string
	TargetPaths []string
	OldName     string
	NewName     string
}

// FileManagerQuickCmdMsg requests execution of a quick shell command in the current directory.
type FileManagerQuickCmdMsg struct {
	HostID  int64
	IsLocal bool
	DirPath string
	Command string
}

// FileManagerQuickCmdResultMsg delivers the output of the executed command.
type FileManagerQuickCmdResultMsg struct {
	HostID  int64
	IsLocal bool
	Command string
	OldCWD  string
	NewCWD  string
	Output  string
	Err     error
}

// PromptType represents an active modal sub-input.
type PromptType int

const (
	PromptNone PromptType = iota
	PromptMkdir
	PromptTouch
	PromptRename
	PromptDeleteConfirm
	PromptMoveDestination
	PromptQuickCmd
	PromptExitConfirm
)

// FileManagerModal represents the dual-pane SFTP file transfer cockpit.
type FileManagerModal struct {
	HostID         int64
	HostName       string
	FocusLocal     bool // true = Local Pane, false = Remote Pane
	LocalPath      string
	LocalItems     []*ssh.FileItem
	LocalCursor    int
	LocalSelected  map[string]bool
	RemotePath     string
	RemoteItems    []*ssh.FileItem
	RemoteCursor   int
	RemoteSelected map[string]bool
	ShowHidden     bool
	ShowRunbook    bool // true = displaying keyboard runbook guide
	StatusMessage  string

	// File manipulation sub-prompt state
	ActivePrompt     PromptType
	PromptInput      textinput.Model
	PromptTargetItem *ssh.FileItem
	TabMatches       []string
	TabMatchIdx      int

	// Clipboard state (Cut / Copy / Paste)
	ClipboardPaths   []string
	ClipboardIsCut   bool // true = Cut (Move), false = Copy
	ClipboardIsLocal bool // true = cut from Local, false = cut from Remote

	// Quick Command Output Viewer state
	ShowCmdOutput    bool
	CmdOutputTitle   string
	CmdOutputContent string
	CmdOutputScroll  int

	// Live Transfer Progress state
	IsTransferring   bool
	TransferIsUpload bool
	TransferIsMove   bool
	CurrentFileName  string
	FileIndex        int
	FileTotal        int
	CurrentBytes     int64
	CurrentTotal     int64
	BytesPerSec      float64
	TransferDoneMsg  string
	TransferDoneTime time.Time
}

// NewFileManagerModal initializes the Dual-Pane SFTP Manager.
func NewFileManagerModal(hostID int64, hostName string, initialLocalPath, initialRemotePath string) *FileManagerModal {
	if initialLocalPath == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			initialLocalPath = home
		} else {
			initialLocalPath = "."
		}
	}
	if initialRemotePath == "" {
		initialRemotePath = "."
	}

	ti := textinput.New()
	ti.Prompt = "❯ "
	ti.Width = 40

	fm := &FileManagerModal{
		HostID:         hostID,
		HostName:       hostName,
		FocusLocal:     false, // Start focused on Remote Server pane
		LocalPath:      initialLocalPath,
		LocalItems:     make([]*ssh.FileItem, 0),
		LocalCursor:    0,
		LocalSelected:  make(map[string]bool),
		RemotePath:     initialRemotePath,
		RemoteItems:    make([]*ssh.FileItem, 0),
		RemoteCursor:   0,
		RemoteSelected: make(map[string]bool),
		ShowHidden:     false,
		ShowRunbook:    false,
		StatusMessage:  i18n.T("sftp_ready"),
		ActivePrompt:   PromptNone,
		PromptInput:    ti,
	}

	return fm
}

// Update handles key navigation, selection, clipboard, runbook, and prompts inside the dual-pane file manager.
func (m *FileManagerModal) Update(msg tea.Msg) (done bool, cmd tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyStr := msg.String()

		// 1. Handle Active Quick Command Output Overlay
		if m.ShowCmdOutput {
			switch keyStr {
			case "esc", "q", "enter", "space", ":":
				m.ShowCmdOutput = false
				m.CmdOutputScroll = 0
				return false, nil
			case "up", "k":
				if m.CmdOutputScroll > 0 {
					m.CmdOutputScroll--
				}
				return false, nil
			case "down", "j":
				m.CmdOutputScroll++
				return false, nil
			default:
				return false, nil
			}
		}

		// 2. Handle Active Runbook Overlay
		if m.ShowRunbook {
			switch keyStr {
			case "esc", "q", "?", "f1", "enter", "space":
				m.ShowRunbook = false
				return false, nil
			default:
				return false, nil
			}
		}

		// 3. Handle Active Input Prompt (mkdir, touch, rename, delete, move, quickcmd)
		if m.ActivePrompt != PromptNone {
			switch keyStr {
			case "esc":
				m.ActivePrompt = PromptNone
				m.PromptInput.Blur()
				m.PromptInput.SetValue("")
				m.StatusMessage = i18n.T("sftp_ready")
				return false, nil

			case "enter":
				val := strings.TrimSpace(m.PromptInput.Value())
				prompt := m.ActivePrompt
				m.ActivePrompt = PromptNone
				m.PromptInput.Blur()
				m.PromptInput.SetValue("")

				targetDir := m.LocalPath
				if !m.FocusLocal {
					targetDir = m.RemotePath
				}

				switch prompt {
				case PromptMkdir:
					if val != "" {
						return false, func() tea.Msg {
							return FileOpActionMsg{
								HostID:  m.HostID,
								IsLocal: m.FocusLocal,
								OpType:  "mkdir",
								DirPath: targetDir,
								NewName: val,
							}
						}
					}

				case PromptTouch:
					if val != "" {
						return false, func() tea.Msg {
							return FileOpActionMsg{
								HostID:  m.HostID,
								IsLocal: m.FocusLocal,
								OpType:  "touch",
								DirPath: targetDir,
								NewName: val,
							}
						}
					}

				case PromptRename:
					if val != "" && m.PromptTargetItem != nil {
						targetItem := m.PromptTargetItem
						m.PromptTargetItem = nil
						return false, func() tea.Msg {
							return FileOpActionMsg{
								HostID:  m.HostID,
								IsLocal: m.FocusLocal,
								OpType:  "rename",
								DirPath: targetDir,
								OldName: targetItem.Name,
								NewName: val,
							}
						}
					}

				case PromptDeleteConfirm:
					selected := m.GetSelectedPaths(m.FocusLocal)
					if len(selected) == 0 && m.PromptTargetItem != nil {
						selected = []string{m.PromptTargetItem.Path}
					}
					m.PromptTargetItem = nil
					if len(selected) > 0 {
						return false, func() tea.Msg {
							return FileOpActionMsg{
								HostID:      m.HostID,
								IsLocal:     m.FocusLocal,
								OpType:      "delete",
								DirPath:     targetDir,
								TargetPaths: selected,
							}
						}
					}

				case PromptMoveDestination:
					selected := m.GetSelectedPaths(m.FocusLocal)
					if len(selected) == 0 && m.PromptTargetItem != nil {
						selected = []string{m.PromptTargetItem.Path}
					}
					m.PromptTargetItem = nil
					if len(selected) > 0 && val != "" {
						destDir := val
						if m.FocusLocal {
							if !filepath.IsAbs(destDir) {
								destDir = filepath.Clean(filepath.Join(m.LocalPath, destDir))
							}
						} else {
							if !strings.HasPrefix(destDir, "/") {
								destDir = path.Clean(path.Join(m.RemotePath, destDir))
							}
						}
						return false, func() tea.Msg {
							return FileOpActionMsg{
								HostID:      m.HostID,
								IsLocal:     m.FocusLocal,
								OpType:      "move",
								DirPath:     targetDir,
								TargetPaths: selected,
								NewName:     destDir,
							}
						}
					}

				case PromptQuickCmd:
					if val != "" {
						return false, func() tea.Msg {
							return FileManagerQuickCmdMsg{
								HostID:  m.HostID,
								IsLocal: m.FocusLocal,
								DirPath: targetDir,
								Command: val,
							}
						}
					}

				case PromptExitConfirm:
					m.ActivePrompt = PromptNone
					return true, nil // Close modal
				}
				return false, nil

			case "y", "Y":
				if m.ActivePrompt == PromptDeleteConfirm {
					m.ActivePrompt = PromptNone
					selected := m.GetSelectedPaths(m.FocusLocal)
					if len(selected) == 0 && m.PromptTargetItem != nil {
						selected = []string{m.PromptTargetItem.Path}
					}
					m.PromptTargetItem = nil
					targetDir := m.LocalPath
					if !m.FocusLocal {
						targetDir = m.RemotePath
					}
					if len(selected) > 0 {
						return false, func() tea.Msg {
							return FileOpActionMsg{
								HostID:      m.HostID,
								IsLocal:     m.FocusLocal,
								OpType:      "delete",
								DirPath:     targetDir,
								TargetPaths: selected,
							}
						}
					}
					return false, nil
				} else if m.ActivePrompt == PromptExitConfirm {
					m.ActivePrompt = PromptNone
					return true, nil // Close modal
				}
				var tiCmd tea.Cmd
				m.PromptInput, tiCmd = m.PromptInput.Update(msg)
				return false, tiCmd

			case "n", "N":
				if m.ActivePrompt == PromptDeleteConfirm || m.ActivePrompt == PromptExitConfirm {
					m.ActivePrompt = PromptNone
					m.PromptTargetItem = nil
					m.StatusMessage = i18n.T("sftp_ready")
					return false, nil
				}
				var tiCmd tea.Cmd
				m.PromptInput, tiCmd = m.PromptInput.Update(msg)
				return false, tiCmd

			case "tab":
				if m.ActivePrompt == PromptMoveDestination {
					curVal := strings.TrimSpace(m.PromptInput.Value())
					prefix := strings.TrimPrefix(curVal, "./")

					if len(m.TabMatches) > 0 {
						m.TabMatchIdx = (m.TabMatchIdx + 1) % len(m.TabMatches)
						m.PromptInput.SetValue(m.TabMatches[m.TabMatchIdx])
						m.PromptInput.CursorEnd()
						return false, nil
					}

					items := m.LocalItems
					if !m.FocusLocal {
						items = m.RemoteItems
					}

					var matches []string
					for _, it := range items {
						if it.IsDir && it.Name != ".." {
							if strings.HasPrefix(strings.ToLower(it.Name), strings.ToLower(prefix)) {
								matches = append(matches, it.Name+"/")
							}
						}
					}
					if strings.HasPrefix("..", prefix) {
						matches = append(matches, "../")
					}

					if len(matches) > 0 {
						m.TabMatches = matches
						m.TabMatchIdx = 0
						m.PromptInput.SetValue(matches[0])
						m.PromptInput.CursorEnd()
					}
					return false, nil
				}
				var tiCmd tea.Cmd
				m.PromptInput, tiCmd = m.PromptInput.Update(msg)
				return false, tiCmd

			default:
				m.TabMatches = nil
				m.TabMatchIdx = 0
				var tiCmd tea.Cmd
				m.PromptInput, tiCmd = m.PromptInput.Update(msg)
				return false, tiCmd
			}
		}

		// 4. Standard Key Navigation & Operations
		switch keyStr {
		case "esc":
			if len(m.ClipboardPaths) > 0 {
				m.ClipboardPaths = nil
				m.StatusMessage = i18n.T("sftp_clipboard_cleared")
				return false, nil
			}
			m.ActivePrompt = PromptExitConfirm
			return false, nil

		case "q", "f", "F6":
			m.ActivePrompt = PromptExitConfirm
			return false, nil

		case "?", "f1":
			m.ShowRunbook = true
			return false, nil

		case ":", "!":
			// Open Quick Shell Command prompt
			m.ActivePrompt = PromptQuickCmd
			m.PromptInput.SetValue("")
			m.PromptInput.Focus()
			return false, nil

		case "tab", "left", "right":
			m.FocusLocal = !m.FocusLocal
			return false, nil

		case "up", "k":
			if m.FocusLocal {
				if m.LocalCursor > 0 {
					m.LocalCursor--
				}
			} else {
				if m.RemoteCursor > 0 {
					m.RemoteCursor--
				}
			}
			return false, nil

		case "down", "j":
			if m.FocusLocal {
				if m.LocalCursor < len(m.LocalItems)-1 {
					m.LocalCursor++
				}
			} else {
				if m.RemoteCursor < len(m.RemoteItems)-1 {
					m.RemoteCursor++
				}
			}
			return false, nil

		case " ":
			// Toggle selection
			if m.FocusLocal {
				if len(m.LocalItems) > 0 && m.LocalCursor >= 0 && m.LocalCursor < len(m.LocalItems) {
					item := m.LocalItems[m.LocalCursor]
					if item.Name != ".." {
						m.LocalSelected[item.Path] = !m.LocalSelected[item.Path]
					}
				}
			} else {
				if len(m.RemoteItems) > 0 && m.RemoteCursor >= 0 && m.RemoteCursor < len(m.RemoteItems) {
					item := m.RemoteItems[m.RemoteCursor]
					if item.Name != ".." {
						m.RemoteSelected[item.Path] = !m.RemoteSelected[item.Path]
					}
				}
			}
			return false, nil

		case "a":
			// Toggle select all
			if m.FocusLocal {
				allSelected := true
				for _, item := range m.LocalItems {
					if item.Name != ".." && !m.LocalSelected[item.Path] {
						allSelected = false
						break
					}
				}
				for _, item := range m.LocalItems {
					if item.Name != ".." {
						m.LocalSelected[item.Path] = !allSelected
					}
				}
			} else {
				allSelected := true
				for _, item := range m.RemoteItems {
					if item.Name != ".." && !m.RemoteSelected[item.Path] {
						allSelected = false
						break
					}
				}
				for _, item := range m.RemoteItems {
					if item.Name != ".." {
						m.RemoteSelected[item.Path] = !allSelected
					}
				}
			}
			return false, nil

		case "x", "ctrl+x":
			// Cut (잘라내기 / 이동 대기)
			selected := m.GetSelectedPaths(m.FocusLocal)
			curItem := m.CurrentCursorItem()
			if len(selected) == 0 && curItem != nil && curItem.Name != ".." {
				selected = []string{curItem.Path}
			}
			if len(selected) > 0 {
				m.ClipboardPaths = selected
				m.ClipboardIsCut = true
				m.ClipboardIsLocal = m.FocusLocal
				m.LocalSelected = make(map[string]bool)
				m.RemoteSelected = make(map[string]bool)
				m.StatusMessage = fmt.Sprintf(i18n.T("sftp_cut_done"), len(selected))
			}
			return false, nil

		case "c", "ctrl+c":
			// Copy (복사 대기)
			selected := m.GetSelectedPaths(m.FocusLocal)
			curItem := m.CurrentCursorItem()
			if len(selected) == 0 && curItem != nil && curItem.Name != ".." {
				selected = []string{curItem.Path}
			}
			if len(selected) > 0 {
				m.ClipboardPaths = selected
				m.ClipboardIsCut = false
				m.ClipboardIsLocal = m.FocusLocal
				m.LocalSelected = make(map[string]bool)
				m.RemoteSelected = make(map[string]bool)
				m.StatusMessage = fmt.Sprintf(i18n.T("sftp_copy_done"), len(selected))
			}
			return false, nil

		case "p", "v", "ctrl+v":
			// Paste (붙여넣기 / 현재 폴더로 이동 또는 복사 실행)
			if len(m.ClipboardPaths) == 0 {
				m.StatusMessage = i18n.T("sftp_paste_empty")
				return false, nil
			}

			destDir := m.RemotePath
			if m.FocusLocal {
				destDir = m.LocalPath
			}

			pathsToPaste := m.ClipboardPaths
			isCut := m.ClipboardIsCut
			isSrcLocal := m.ClipboardIsLocal

			m.ClipboardPaths = nil // Unconditionally clear clipboard upon paste

			// Same side paste (Remote -> Remote or Local -> Local)
			if isSrcLocal == m.FocusLocal {
				opType := "copy_same"
				if isCut {
					opType = "move"
				}
				return false, func() tea.Msg {
					return FileOpActionMsg{
						HostID:      m.HostID,
						IsLocal:     m.FocusLocal,
						OpType:      opType,
						DirPath:     destDir,
						TargetPaths: pathsToPaste,
						NewName:     destDir,
					}
				}
			}

			// Cross side paste (Upload or Download)
			isUpload := isSrcLocal
			return false, func() tea.Msg {
				return TransferActionMsg{
					HostID:      m.HostID,
					HostName:    m.HostName,
					IsUpload:    isUpload,
					IsMove:      isCut,
					SrcPaths:    pathsToPaste,
					DestDirPath: destDir,
				}
			}

		case "enter":
			// Enter directory with rollback protection
			if m.FocusLocal {
				if len(m.LocalItems) > 0 && m.LocalCursor >= 0 && m.LocalCursor < len(m.LocalItems) {
					item := m.LocalItems[m.LocalCursor]
					if item.IsDir {
						m.LocalCursor = 0
						m.LocalSelected = make(map[string]bool)
						return false, m.NavigateLocalCmd(item.Path, m.LocalPath)
					}
				}
			} else {
				if len(m.RemoteItems) > 0 && m.RemoteCursor >= 0 && m.RemoteCursor < len(m.RemoteItems) {
					item := m.RemoteItems[m.RemoteCursor]
					if item.IsDir {
						m.RemoteCursor = 0
						m.RemoteSelected = make(map[string]bool)
						return false, m.NavigateRemoteCmd(item.Path, m.RemotePath)
					}
				}
			}
			return false, nil

		case "backspace":
			// Go to parent directory with rollback protection
			if m.FocusLocal {
				parent := filepath.Dir(m.LocalPath)
				m.LocalCursor = 0
				return false, m.NavigateLocalCmd(parent, m.LocalPath)
			} else {
				parent := path.Dir(m.RemotePath)
				if parent == "" {
					parent = "/"
				}
				m.RemoteCursor = 0
				return false, m.NavigateRemoteCmd(parent, m.RemotePath)
			}

		case "u", "U":
			// Upload: Local -> Remote
			if m.IsTransferring {
				return false, nil
			}
			selected := m.GetSelectedPaths(true)
			if len(selected) == 0 && len(m.LocalItems) > 0 && m.LocalCursor >= 0 && m.LocalCursor < len(m.LocalItems) {
				item := m.LocalItems[m.LocalCursor]
				if item.Name != ".." {
					selected = []string{item.Path}
				}
			}
			if len(selected) > 0 {
				m.IsTransferring = true
				m.TransferIsUpload = true
				m.TransferIsMove = false
				m.FileIndex = 1
				m.FileTotal = len(selected)
				m.CurrentFileName = filepath.Base(selected[0])
				m.TransferDoneMsg = ""
				m.StatusMessage = fmt.Sprintf("⏳ Uploading %d file(s) to %s...", len(selected), m.RemotePath)
				return false, func() tea.Msg {
					return TransferActionMsg{
						HostID:      m.HostID,
						HostName:    m.HostName,
						IsUpload:    true,
						IsMove:      false,
						SrcPaths:    selected,
						DestDirPath: m.RemotePath,
					}
				}
			}
			return false, nil

		case "d", "D":
			// Download: Remote -> Local
			if m.IsTransferring {
				return false, nil
			}
			selected := m.GetSelectedPaths(false)
			if len(selected) == 0 && len(m.RemoteItems) > 0 && m.RemoteCursor >= 0 && m.RemoteCursor < len(m.RemoteItems) {
				item := m.RemoteItems[m.RemoteCursor]
				if item.Name != ".." {
					selected = []string{item.Path}
				}
			}
			if len(selected) > 0 {
				m.IsTransferring = true
				m.TransferIsUpload = false
				m.TransferIsMove = false
				m.FileIndex = 1
				m.FileTotal = len(selected)
				m.CurrentFileName = filepath.Base(selected[0])
				m.TransferDoneMsg = ""
				m.StatusMessage = fmt.Sprintf("⏳ Downloading %d file(s) to %s...", len(selected), m.LocalPath)
				return false, func() tea.Msg {
					return TransferActionMsg{
						HostID:      m.HostID,
						HostName:    m.HostName,
						IsUpload:    false,
						IsMove:      false,
						SrcPaths:    selected,
						DestDirPath: m.LocalPath,
					}
				}
			}
			return false, nil

		case "m", "M":
			// Move prompt: moves selected files on current pane to destination directory
			selected := m.GetSelectedPaths(m.FocusLocal)
			curItem := m.CurrentCursorItem()
			if len(selected) == 0 && curItem != nil && curItem.Name != ".." {
				selected = []string{curItem.Path}
			}
			if len(selected) > 0 {
				m.ActivePrompt = PromptMoveDestination
				m.PromptTargetItem = curItem
				m.PromptInput.SetValue("./")
				m.PromptInput.Focus()
				return false, nil
			}
			return false, nil

		case "n":
			// New Folder Prompt
			m.ActivePrompt = PromptMkdir
			m.PromptInput.SetValue("")
			m.PromptInput.Focus()
			return false, nil

		case "N", "ctrl+n":
			// New File Prompt
			m.ActivePrompt = PromptTouch
			m.PromptInput.SetValue("")
			m.PromptInput.Focus()
			return false, nil

		case "r", "R":
			// Rename Prompt
			curItem := m.CurrentCursorItem()
			if curItem != nil && curItem.Name != ".." {
				m.ActivePrompt = PromptRename
				m.PromptTargetItem = curItem
				m.PromptInput.SetValue(curItem.Name)
				m.PromptInput.Focus()
				return false, nil
			}
			return false, tea.Batch(m.RefreshLocalCmd(), m.RefreshRemoteCmd())

		case "delete", "shift+x", "X":
			// Delete Prompt
			curItem := m.CurrentCursorItem()
			selected := m.GetSelectedPaths(m.FocusLocal)
			if len(selected) > 0 || (curItem != nil && curItem.Name != "..") {
				m.ActivePrompt = PromptDeleteConfirm
				m.PromptTargetItem = curItem
				return false, nil
			}
			return false, nil

		case ".":
			m.ShowHidden = !m.ShowHidden
			return false, tea.Batch(m.RefreshLocalCmd(), m.RefreshRemoteCmd())

		case "f5":
			return false, tea.Batch(m.RefreshLocalCmd(), m.RefreshRemoteCmd())
		}
	}

	return false, nil
}

// CurrentCursorItem returns the item under the cursor on the focused side.
func (m *FileManagerModal) CurrentCursorItem() *ssh.FileItem {
	if m.FocusLocal {
		if len(m.LocalItems) > 0 && m.LocalCursor >= 0 && m.LocalCursor < len(m.LocalItems) {
			return m.LocalItems[m.LocalCursor]
		}
	} else {
		if len(m.RemoteItems) > 0 && m.RemoteCursor >= 0 && m.RemoteCursor < len(m.RemoteItems) {
			return m.RemoteItems[m.RemoteCursor]
		}
	}
	return nil
}

// GetSelectedPaths returns list of selected paths for either local or remote.
func (m *FileManagerModal) GetSelectedPaths(isLocal bool) []string {
	var paths []string
	targetMap := m.LocalSelected
	if !isLocal {
		targetMap = m.RemoteSelected
	}
	for path, selected := range targetMap {
		if selected {
			paths = append(paths, path)
		}
	}
	return paths
}

// NavigateLocalCmd scans local filesystem and handles permission rollback.
func (m *FileManagerModal) NavigateLocalCmd(newPath, oldPath string) tea.Cmd {
	return func() tea.Msg {
		items, err := ssh.ListLocalDirectory(newPath, m.ShowHidden)
		return LocalFileListMsg{
			HostID:  m.HostID,
			Path:    newPath,
			OldPath: oldPath,
			Items:   items,
			Err:     err,
		}
	}
}

// NavigateRemoteCmd scans remote server and handles permission rollback.
func (m *FileManagerModal) NavigateRemoteCmd(newPath, oldPath string) tea.Cmd {
	return func() tea.Msg {
		return NavigateRemoteMsg{
			HostID:  m.HostID,
			NewPath: newPath,
			OldPath: oldPath,
		}
	}
}

// RefreshLocalCmd scans local filesystem.
func (m *FileManagerModal) RefreshLocalCmd() tea.Cmd {
	return m.NavigateLocalCmd(m.LocalPath, m.LocalPath)
}

// RefreshRemoteCmd scans remote server filesystem.
func (m *FileManagerModal) RefreshRemoteCmd() tea.Cmd {
	return m.NavigateRemoteCmd(m.RemotePath, m.RemotePath)
}

// View renders the dual-pane file manager, output overlay, or runbook guide.
func (m *FileManagerModal) View(totalWidth, totalHeight int) string {
	if totalWidth < 60 {
		totalWidth = 60
	}
	if totalHeight < 14 {
		totalHeight = 14
	}

	// 1. If Quick Command Output overlay is active, render it
	if m.ShowCmdOutput {
		return m.renderCmdOutput(totalWidth, totalHeight)
	}

	// 2. If Runbook guide is active, render it
	if m.ShowRunbook {
		return m.renderRunbook(totalWidth, totalHeight)
	}

	// 3. Header Bar (1 line)
	titleText := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("📂 " + i18n.T("sftp_title"))
	serverInfo := lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf(i18n.T("sftp_server"), m.HostName))

	gapWidth := totalWidth - lipgloss.Width(titleText) - lipgloss.Width(serverInfo) - 2
	if gapWidth < 1 {
		gapWidth = 1
	}
	headerBar := lipgloss.JoinHorizontal(lipgloss.Center, titleText, strings.Repeat(" ", gapWidth), serverInfo)

	// 4. Dual Pane Dimensions
	availHeight := totalHeight - 4
	if availHeight < 8 {
		availHeight = 8
	}

	leftWidth := (totalWidth - 2) / 2
	rightWidth := totalWidth - leftWidth - 2

	localPane := m.renderPane(true, leftWidth, availHeight)
	remotePane := m.renderPane(false, rightWidth, availHeight)

	dualPanes := lipgloss.JoinHorizontal(lipgloss.Top, localPane, " ", remotePane)

	// 5. Bottom Transfer / Prompt Deck (1 line)
	transferBar := m.renderTransferDeck(totalWidth - 2)

	// 6. Bottom Legend Bar (1 line)
	legendText := lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("sftp_legend"))

	mainLayout := lipgloss.JoinVertical(lipgloss.Left,
		headerBar,
		dualPanes,
		transferBar,
		legendText,
	)

	return lipgloss.Place(totalWidth, totalHeight, lipgloss.Left, lipgloss.Top, mainLayout)
}

func (m *FileManagerModal) renderPane(isLocal bool, width, height int) string {
	var lines []string

	isFocused := (isLocal == m.FocusLocal)
	borderColor := ColorBorder
	headerColor := ColorMuted
	title := i18n.T("sftp_local_title")
	currentPath := m.LocalPath
	items := m.LocalItems
	cursor := m.LocalCursor
	selectedMap := m.LocalSelected

	if !isLocal {
		title = fmt.Sprintf(i18n.T("sftp_remote_title"), m.HostName)
		currentPath = m.RemotePath
		items = m.RemoteItems
		cursor = m.RemoteCursor
		selectedMap = m.RemoteSelected
	}

	if isFocused {
		borderColor = ColorPrimary
		headerColor = ColorPrimary
	}

	if cursor >= len(items) {
		if len(items) > 0 {
			cursor = len(items) - 1
		} else {
			cursor = 0
		}
	}
	if cursor < 0 {
		cursor = 0
	}

	innerWidth := width - 4
	if innerWidth < 15 {
		innerWidth = 15
	}

	// 1. Pane Header
	pathColor := lipgloss.Color("#80D8FF") // Bright Light Cyan
	if !isLocal {
		pathColor = lipgloss.Color("#FFE082") // Bright Light Amber
	}

	focusBadge := " [Tab전환]"
	if isFocused {
		focusBadge = " [● 활성/ACTIVE]"
	}
	fullTitle := title + focusBadge

	titlePadded := padToWidth(runewidth.Truncate(fullTitle, innerWidth, "…"), innerWidth)
	pathPadded := padToWidth(runewidth.Truncate("📍 "+currentPath, innerWidth, "…"), innerWidth)

	lines = append(lines,
		lipgloss.NewStyle().Bold(true).Foreground(headerColor).Render(titlePadded),
		lipgloss.NewStyle().Bold(true).Foreground(pathColor).Render(pathPadded),
		lipgloss.NewStyle().Foreground(borderColor).Render(strings.Repeat("─", innerWidth)),
	)

	// 2. Visible File Rows
	maxRows := height - 5
	if maxRows < 3 {
		maxRows = 3
	}

	clipboardMap := make(map[string]bool)
	for _, cp := range m.ClipboardPaths {
		if m.ClipboardIsLocal == isLocal {
			clipboardMap[cp] = true
		}
	}

	if len(items) == 0 {
		emptyMsg := padToWidth(i18n.T("sftp_empty"), innerWidth)
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorMuted).Render(emptyMsg))
		for len(lines) < maxRows+3 {
			lines = append(lines, strings.Repeat(" ", innerWidth))
		}
	} else {
		startIdx := 0
		if cursor >= maxRows {
			startIdx = cursor - maxRows + 1
		}
		endIdx := startIdx + maxRows
		if endIdx > len(items) {
			endIdx = len(items)
		}

		for i := startIdx; i < endIdx; i++ {
			item := items[i]
			isCur := (i == cursor && isFocused)
			isSelected := selectedMap[item.Path]
			isClipboard := clipboardMap[item.Path]

			icon := "📄"
			if item.IsDir {
				icon = "📁"
			}

			check := "[ ]"
			if isSelected {
				check = "[*]"
			}
			if isClipboard {
				if m.ClipboardIsCut {
					check = "[✂]"
				} else {
					check = "[📋]"
				}
			}
			if item.Name == ".." {
				check = "   "
			}

			sizeStr := fmt.Sprintf("%8s", item.FormatSize())

			availNameW := innerWidth - runewidth.StringWidth(check) - runewidth.StringWidth(icon) - len(sizeStr) - 3
			if availNameW < 4 {
				availNameW = 4
			}
			cleanName := runewidth.Truncate(item.Name, availNameW, "…")

			leftPart := fmt.Sprintf("%s %s %s", check, icon, cleanName)
			rowPadded := padToWidth(leftPart, innerWidth-len(sizeStr)) + sizeStr

			var renderedRow string
			if isCur {
				renderedRow = lipgloss.NewStyle().Bold(true).Foreground(ColorBg).Background(ColorPrimary).Render(rowPadded)
			} else if isClipboard {
				renderedRow = lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(rowPadded)
			} else if isSelected {
				renderedRow = lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(rowPadded)
			} else {
				renderedRow = lipgloss.NewStyle().Foreground(ColorText).Render(rowPadded)
			}

			lines = append(lines, renderedRow)
		}

		// Fill remaining rows so both panes have the exact same line height
		for len(lines) < maxRows+3 {
			lines = append(lines, strings.Repeat(" ", innerWidth))
		}
	}

	paneBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(width - 2).
		Height(height - 2)

	return paneBox.Render(strings.Join(lines, "\n"))
}

func padToWidth(s string, targetWidth int) string {
	w := runewidth.StringWidth(s)
	if w >= targetWidth {
		return runewidth.Truncate(s, targetWidth, "")
	}
	return s + strings.Repeat(" ", targetWidth-w)
}

func (m *FileManagerModal) renderTransferDeck(width int) string {
	// 1. Active Sub-prompt Input
	if m.ActivePrompt != PromptNone {
		var promptLabel string
		switch m.ActivePrompt {
		case PromptMkdir:
			promptLabel = i18n.T("sftp_mkdir_prompt")
		case PromptTouch:
			promptLabel = i18n.T("sftp_touch_prompt")
		case PromptRename:
			promptLabel = i18n.T("sftp_rename_prompt")
		case PromptDeleteConfirm:
			selected := m.GetSelectedPaths(m.FocusLocal)
			var promptLabel string
			if len(selected) == 0 && m.PromptTargetItem != nil {
				typeName := "파일"
				if m.PromptTargetItem.IsDir {
					typeName = "폴더"
				}
				switch i18n.GetLang() {
				case i18n.LangKO:
					promptLabel = fmt.Sprintf("🗑️ '%s' (%s)을(를) 정말 영구 삭제하시겠습니까? [y/Enter: 삭제, n/Esc: 취소]", m.PromptTargetItem.Name, typeName)
				case i18n.LangDE:
					deType := "Datei"
					if m.PromptTargetItem.IsDir {
						deType = "Ordner"
					}
					promptLabel = fmt.Sprintf("🗑️ Möchten Sie '%s' (%s) wirklich dauerhaft löschen? [y/n]", m.PromptTargetItem.Name, deType)
				default:
					enType := "file"
					if m.PromptTargetItem.IsDir {
						enType = "folder"
					}
					promptLabel = fmt.Sprintf("🗑️ Permanently delete '%s' (%s)? [y/n]", m.PromptTargetItem.Name, enType)
				}
			} else if len(selected) == 1 {
				baseName := filepath.Base(selected[0])
				switch i18n.GetLang() {
				case i18n.LangKO:
					promptLabel = fmt.Sprintf("🗑️ '%s' 항목을 정말 영구 삭제하시겠습니까? [y/Enter: 삭제, n/Esc: 취소]", baseName)
				case i18n.LangDE:
					promptLabel = fmt.Sprintf("🗑️ Möchten Sie '%s' wirklich dauerhaft löschen? [y/n]", baseName)
				default:
					promptLabel = fmt.Sprintf("🗑️ Permanently delete '%s'? [y/n]", baseName)
				}
			} else if len(selected) == 2 {
				b1 := filepath.Base(selected[0])
				b2 := filepath.Base(selected[1])
				switch i18n.GetLang() {
				case i18n.LangKO:
					promptLabel = fmt.Sprintf("🗑️ 선택한 2개 항목('%s', '%s')을 정말 삭제하시겠습니까? [y/Enter: 삭제, n/Esc: 취소]", b1, b2)
				case i18n.LangDE:
					promptLabel = fmt.Sprintf("🗑️ Ausgewählte 2 Einträge ('%s', '%s') wirklich löschen? [y/n]", b1, b2)
				default:
					promptLabel = fmt.Sprintf("🗑️ Delete selected 2 items ('%s', '%s')? [y/n]", b1, b2)
				}
			} else if len(selected) > 2 {
				b1 := filepath.Base(selected[0])
				b2 := filepath.Base(selected[1])
				remaining := len(selected) - 2
				switch i18n.GetLang() {
				case i18n.LangKO:
					promptLabel = fmt.Sprintf("🗑️ 선택한 %d개 항목('%s', '%s' 외 %d개)을 정말 삭제하시겠습니까? [y/Enter: 삭제, n/Esc: 취소]", len(selected), b1, b2, remaining)
				case i18n.LangDE:
					promptLabel = fmt.Sprintf("🗑️ Ausgewählte %d Einträge ('%s', '%s' u. %d weitere) wirklich löschen? [y/n]", len(selected), b1, b2, remaining)
				default:
					promptLabel = fmt.Sprintf("🗑️ Delete selected %d items ('%s', '%s' and %d more)? [y/n]", len(selected), b1, b2, remaining)
				}
			} else {
				promptLabel = "🗑️ " + i18n.T("sftp_delete_confirm")
			}
			return lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(padToWidth(promptLabel, width))
		case PromptMoveDestination:
			targetSide := "원격"
			if m.FocusLocal {
				targetSide = "로컬"
			}
			promptLabel = fmt.Sprintf("📦 [%s 파일이동] 대상 폴더명 입력 (Tab 자동완성):", targetSide)
		case PromptExitConfirm:
			exitMsg := fmt.Sprintf("🚪 %s  [y/Enter: 종료, n/Esc: 취소]", i18n.T("sftp_exit_confirm"))
			return lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(padToWidth(exitMsg, width))
		case PromptQuickCmd:
			targetSide := "원격"
			if m.FocusLocal {
				targetSide = "로컬"
			}
			promptLabel = fmt.Sprintf("💻 [%s 즉석쉘] 실행할 리눅스 명령어 ($):", targetSide)
		}
		promptLine := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render(promptLabel+" ") + m.PromptInput.View()
		return padToWidth(promptLine, width)
	}

	// 2. Active Transfer in Progress
	if m.IsTransferring {
		actionStr := i18n.T("sftp_uploading")
		if !m.TransferIsUpload {
			actionStr = i18n.T("sftp_downloading")
		}
		if m.TransferIsMove {
			actionStr = "📦 MOVING"
		}

		percent := 0.0
		if m.CurrentTotal > 0 {
			percent = float64(m.CurrentBytes) / float64(m.CurrentTotal) * 100.0
		}
		if percent > 100.0 {
			percent = 100.0
		}

		barWidth := width - 45
		if barWidth < 10 {
			barWidth = 10
		}
		filled := int((percent / 100.0) * float64(barWidth))
		if filled > barWidth {
			filled = barWidth
		}

		bar := lipgloss.NewStyle().Foreground(ColorSuccess).Render(strings.Repeat("█", filled)) +
			lipgloss.NewStyle().Foreground(ColorMuted).Render(strings.Repeat("░", barWidth-filled))

		curMB := float64(m.CurrentBytes) / (1024 * 1024)
		totMB := float64(m.CurrentTotal) / (1024 * 1024)
		speedStr := fmt.Sprintf("%.1f MB/s", m.BytesPerSec/(1024*1024))
		if m.BytesPerSec == 0 {
			speedStr = "calculating..."
		}

		transferLine := fmt.Sprintf("%s (%d/%d): %s  [%s] %.0f%% (%.1fMB/%.1fMB) • %s",
			lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(actionStr),
			m.FileIndex, m.FileTotal,
			lipgloss.NewStyle().Bold(true).Render(runewidth.Truncate(m.CurrentFileName, 16, "…")),
			bar, percent, curMB, totMB, speedStr,
		)

		return padToWidth(transferLine, width)
	}

	// 3. Active Clipboard Status Banner
	if len(m.ClipboardPaths) > 0 {
		clipType := "이동(잘라내기)"
		if !m.ClipboardIsCut {
			clipType = "복사"
		}
		clipBanner := fmt.Sprintf("✂️ [%s 대기] %d개 파일 선택됨 • 원하는 폴더로 자유 이동 후 [p] 를 눌러 투하하세요 (취소: [Esc])", clipType, len(m.ClipboardPaths))
		return lipgloss.NewStyle().Bold(true).Foreground(ColorWarning).Render(padToWidth(clipBanner, width))
	}

	// 4. Transfer Complete Banner
	if m.TransferDoneMsg != "" {
		return lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(padToWidth(m.TransferDoneMsg, width))
	}

	// 5. Default Status Message
	msg := padToWidth(m.StatusMessage, width)
	if strings.HasPrefix(m.StatusMessage, "⚠️") || strings.HasPrefix(m.StatusMessage, "❌") {
		return lipgloss.NewStyle().Bold(true).Foreground(ColorDanger).Render(msg)
	}
	if strings.HasPrefix(m.StatusMessage, "✨") || strings.HasPrefix(m.StatusMessage, "🎉") {
		return lipgloss.NewStyle().Bold(true).Foreground(ColorSuccess).Render(msg)
	}
	return lipgloss.NewStyle().Foreground(ColorMuted).Render(msg)
}

// renderCmdOutput renders the floating output box of a quick shell command.
func (m *FileManagerModal) renderCmdOutput(width, height int) string {
	boxWidth := width - 6
	if boxWidth > 110 {
		boxWidth = 110
	}
	if boxWidth < 50 {
		boxWidth = 50
	}
	boxHeight := height - 4
	if boxHeight < 14 {
		boxHeight = 14
	}

	innerWidth := boxWidth - 4
	availBodyHeight := boxHeight - 5
	if availBodyHeight < 5 {
		availBodyHeight = 5
	}

	titleText := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("💻 " + m.CmdOutputTitle)
	closeHint := lipgloss.NewStyle().Foreground(ColorMuted).Render("[Esc/Enter] 닫기  •  [↑/↓] 스크롤")

	rawLines := strings.Split(m.CmdOutputContent, "\n")
	if len(rawLines) == 0 || (len(rawLines) == 1 && rawLines[0] == "") {
		rawLines = []string{"(명령어가 성공적으로 실행되었습니다. 출력 결과 없음)"}
	}

	// Clamp scroll
	maxScroll := len(rawLines) - availBodyHeight
	if maxScroll < 0 {
		maxScroll = 0
	}
	if m.CmdOutputScroll > maxScroll {
		m.CmdOutputScroll = maxScroll
	}
	if m.CmdOutputScroll < 0 {
		m.CmdOutputScroll = 0
	}

	var visibleLines []string
	startIdx := m.CmdOutputScroll
	endIdx := startIdx + availBodyHeight
	if endIdx > len(rawLines) {
		endIdx = len(rawLines)
	}

	for i := startIdx; i < endIdx; i++ {
		line := rawLines[i]
		lineTrunc := padToWidth(runewidth.Truncate(line, innerWidth, "…"), innerWidth)
		visibleLines = append(visibleLines, lipgloss.NewStyle().Foreground(ColorText).Render(lineTrunc))
	}

	// Fill empty lines
	for len(visibleLines) < availBodyHeight {
		visibleLines = append(visibleLines, strings.Repeat(" ", innerWidth))
	}

	divider := lipgloss.NewStyle().Foreground(ColorBorder).Render(strings.Repeat("─", innerWidth))

	contentBox := lipgloss.JoinVertical(lipgloss.Left,
		titleText,
		divider,
		strings.Join(visibleLines, "\n"),
		divider,
		closeHint,
	)

	dialogBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorPrimary).
		Padding(0, 1).
		Width(boxWidth).
		Height(boxHeight).
		Render(contentBox)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, dialogBox)
}

// renderRunbook renders the dedicated File Manager shortcut & operations guide.
func (m *FileManagerModal) renderRunbook(width, height int) string {
	boxWidth := width - 4
	if boxWidth > 90 {
		boxWidth = 90
	}
	boxHeight := height - 4
	if boxHeight < 16 {
		boxHeight = 16
	}

	var lines []string

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorWarning)
	keyStyle := lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary)
	descStyle := lipgloss.NewStyle().Foreground(ColorText)

	lines = append(lines, titleStyle.Render(i18n.T("sftp_runbook_title")), "")

	type item struct {
		key  string
		desc string
	}

	type section struct {
		title string
		items []item
	}

	var sections []section

	switch i18n.GetLang() {
	case i18n.LangKO:
		sections = []section{
			{
				title: "🧭 탐색 & 선택 (Navigation & Selection)",
				items: []item{
					{"[Tab], [◄/►]", "내 PC(로컬) ↔ 원격 서버 패널 전환"},
					{"[↑/↓], [j/k]", "파일/폴더 목록 커서 이동"},
					{"[Enter]", "선택한 폴더 진입 (권한 부족 시 자동 롤백)"},
					{"[Backspace]", "상위 디렉토리(..)로 이동"},
					{"[Space]", "파일/폴더 다중 선택 토글 ([*] 녹색 배지)"},
					{"[a]", "현재 디렉토리 내 전체 선택 / 해제"},
					{"[.]", "숨김 파일(.env, .* 등) 표시 / 숨김 토글"},
				},
			},
			{
				title: "✂️ 클립보드 & 자유 파일 이동 (Cut, Copy, Paste)",
				items: []item{
					{"[x], [Ctrl+X]", "잘라내기(이동 대기): 파일 선택 후 어디든 탐색 가능 ([✂] 배지)"},
					{"[c], [Ctrl+C]", "복사: 파일 선택 후 원하는 폴더로 복제 대기 ([📋] 배지)"},
					{"[p], [Ctrl+V]", "붙여넣기(투하): 현재 열려있는 폴더로 즉시 이동/복사 완료!"},
					{"[Esc]", "클립보드(잘라내기/복사 대기) 취소"},
				},
			},
			{
				title: "🚀 파일 전송 (Upload & Download)",
				items: []item{
					{"[u]", "업로드: 로컬 PC → 원격 서버로 선택 파일 복사"},
					{"[d]", "다운로드: 원격 서버 → 로컬 PC로 선택 파일 복사"},
				},
			},
			{
				title: "🛠️ 파일 조작 & 즉석 쉘 (Operations & Shell)",
				items: []item{
					{"[:], [!]", "현재 경로에서 즉석 쉘 명령어 실행 (ls -la, cd .., df 등)"},
					{"[n]", "새 폴더 생성 (mkdir 인라인 입력)"},
					{"[N], [Ctrl+N]", "새 빈 파일 생성 (touch 인라인 입력)"},
					{"[r]", "선택한 파일/폴더 이름 변경 (mv)"},
					{"[Delete], [X]", "선택한 파일/폴더 일괄 삭제 (rm -rf 확인 후)"},
					{"[F5]", "현재 디렉토리 파일 목록 새로고침"},
				},
			},
			{
				title: "🚪 창 닫기 & 도움말",
				items: []item{
					{"[?], [F1]", "런북(단축키 가이드) 열기 / 닫기"},
					{"[Esc], [q], [f]", "파일 매니저 닫기 (종료 확인 프롬프트)"},
				},
			},
		}

	case i18n.LangDE:
		sections = []section{
			{
				title: "🧭 Navigation & Auswahl",
				items: []item{
					{"[Tab], [◄/►]", "Fenster wechseln (Lokaler PC ↔ Remote-Server)"},
					{"[↑/↓], [j/k]", "In Dateiliste navigieren"},
					{"[Enter]", "Ordner öffnen (Rollback bei fehlenden Rechten)"},
					{"[Backspace]", "Übergeordnetes Verzeichnis (..)"},
					{"[Space]", "Mehrfachauswahl umschalten ([*] Grün)"},
					{"[a]", "Alle Dateien auswählen / abwählen"},
					{"[.]", "Versteckte Dateien ein-/ausblenden"},
				},
			},
			{
				title: "✂️ Zwischenablage (Cut, Copy, Paste)",
				items: []item{
					{"[x], [Ctrl+X]", "Ausschneiden (Verschiebung vormerken)"},
					{"[c], [Ctrl+C]", "Kopieren (Kopieren vormerken)"},
					{"[p], [Ctrl+V]", "Einfügen (In aktuellen Ordner übertragen)"},
					{"[Esc]", "Zwischenablage leeren / abbrechen"},
				},
			},
			{
				title: "🚀 Dateiübertragung (Upload & Download)",
				items: []item{
					{"[u]", "Hochladen: Lokaler PC → Remote-Server"},
					{"[d]", "Herunterladen: Remote-Server → Lokaler PC"},
				},
			},
			{
				title: "🛠️ Dateioperationen & Shell",
				items: []item{
					{"[:], [!]", "Shell-Befehl im aktuellen Pfad ausführen (ls -la, cd .., df)"},
					{"[n]", "Neuen Ordner erstellen (mkdir)"},
					{"[N], [Ctrl+N]", "Neue Datei erstellen (touch)"},
					{"[r]", "Datei/Ordner umbenennen (mv)"},
					{"[Delete], [X]", "Ausgewählte Dateien löschen (rm -rf)"},
					{"[F5]", "Dateiliste aktualisieren"},
				},
			},
			{
				title: "🚪 Beenden & Hilfe",
				items: []item{
					{"[?], [F1]", "Runbook / Tastenkürzel ein-/ausblenden"},
					{"[Esc], [q], [f]", "Dateimanager schließen (mit Bestätigung)"},
				},
			},
		}

	default: // EN
		sections = []section{
			{
				title: "🧭 Navigation & Selection",
				items: []item{
					{"[Tab], [◄/►]", "Switch pane focus (Local PC ↔ Remote Server)"},
					{"[↑/↓], [j/k]", "Navigate file & folder list"},
					{"[Enter]", "Open directory (Protected with auto-rollback)"},
					{"[Backspace]", "Go to parent directory (..)"},
					{"[Space]", "Toggle multi-selection badge ([*])"},
					{"[a]", "Select / Deselect all files in current folder"},
					{"[.]", "Toggle hidden files (.env, .*)"},
				},
			},
			{
				title: "✂️ Clipboard Operations (Cut, Copy, Paste)",
				items: []item{
					{"[x], [Ctrl+X]", "Cut (Stage for Move): Browse freely anywhere ([✂])"},
					{"[c], [Ctrl+C]", "Copy (Stage for Duplicate): ([📋])"},
					{"[p], [Ctrl+V]", "Paste: Move/Copy files into currently viewed folder!"},
					{"[Esc]", "Clear staged clipboard"},
				},
			},
			{
				title: "🚀 File Transfer (Upload & Download)",
				items: []item{
					{"[u]", "Upload: Copy selected files to Remote Server"},
					{"[d]", "Download: Copy selected files to Local PC"},
				},
			},
			{
				title: "🛠️ File Operations & Quick Shell",
				items: []item{
					{"[:], [!]", "Execute instant shell command (ls -la, cd .., df, etc.)"},
					{"[n]", "Create new folder (mkdir)"},
					{"[N], [Ctrl+N]", "Create new empty file (touch)"},
					{"[r]", "Rename selected file or folder (mv)"},
					{"[Delete], [X]", "Delete selected file(s)/folder(s) (rm -rf)"},
					{"[F5]", "Refresh directory listing"},
				},
			},
			{
				title: "🚪 Exit & Help",
				items: []item{
					{"[?], [F1]", "Open / Close this Runbook guide"},
					{"[Esc], [q], [f]", "Close File Manager (with confirmation)"},
				},
			},
		}
	}

	for _, sec := range sections {
		lines = append(lines, sectionStyle.Render(sec.title))
		for _, it := range sec.items {
			keyPadded := padToWidth(it.key, 16)
			row := fmt.Sprintf("  %s : %s", keyStyle.Render(keyPadded), descStyle.Render(it.desc))
			lines = append(lines, row)
		}
		lines = append(lines, "")
	}

	lines = append(lines, lipgloss.NewStyle().Foreground(ColorMuted).Render(i18n.T("sftp_runbook_close")))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2).
		Width(boxWidth).
		Height(boxHeight)

	contentStr := strings.Join(lines, "\n")
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, boxStyle.Render(contentStr))
}
