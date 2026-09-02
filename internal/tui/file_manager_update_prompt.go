package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (fm *FileManagerModal) UpdatePrompt(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.String() {
	case "esc":
		fm.ActivePrompt = PromptNone
		fm.SubInput.SetValue("")
		fm.QuickCmdInput.SetValue("")
		return false, nil

	case "enter":
		switch fm.ActivePrompt {
		case PromptMkdir:
			dirName := strings.TrimSpace(fm.SubInput.Value())
			if dirName != "" {
				fm.ActivePrompt = PromptNone
				fm.SubInput.SetValue("")
				targetDir := fm.RemotePath
				if fm.FocusLocal {
					targetDir = fm.LocalPath
				}
				return false, func() tea.Msg {
					return FileOpActionMsg{
						HostID:  fm.HostID,
						IsLocal: fm.FocusLocal,
						OpType:  "mkdir",
						DirPath: targetDir,
						NewName: dirName,
					}
				}
			}

		case PromptTouch:
			fileName := strings.TrimSpace(fm.SubInput.Value())
			if fileName != "" {
				fm.ActivePrompt = PromptNone
				fm.SubInput.SetValue("")
				targetDir := fm.RemotePath
				if fm.FocusLocal {
					targetDir = fm.LocalPath
				}
				return false, func() tea.Msg {
					return FileOpActionMsg{
						HostID:  fm.HostID,
						IsLocal: fm.FocusLocal,
						OpType:  "touch",
						DirPath: targetDir,
						NewName: fileName,
					}
				}
			}

		case PromptRename:
			newName := strings.TrimSpace(fm.SubInput.Value())
			if newName != "" {
				fm.ActivePrompt = PromptNone
				fm.SubInput.SetValue("")
				targetDir := fm.RemotePath
				if fm.FocusLocal {
					targetDir = fm.LocalPath
				}
				items := fm.GetActiveItems()
				cursor := fm.GetActiveCursor()
				oldName := ""
				if cursor >= 0 && cursor < len(items) {
					oldName = items[cursor].Name
				}
				if oldName != "" {
					return false, func() tea.Msg {
						return FileOpActionMsg{
							HostID:  fm.HostID,
							IsLocal: fm.FocusLocal,
							OpType:  "rename",
							DirPath: targetDir,
							OldName: oldName,
							NewName: newName,
						}
					}
				}
			}

		case PromptDeleteConfirm:
			fm.ActivePrompt = PromptNone
			pathsToDelete := fm.GetSelectedPaths()
			fm.ClearSelections()

			if len(pathsToDelete) > 0 {
				targetDir := fm.RemotePath
				if fm.FocusLocal {
					targetDir = fm.LocalPath
				}
				return false, func() tea.Msg {
					return FileOpActionMsg{
						HostID:      fm.HostID,
						IsLocal:     fm.FocusLocal,
						OpType:      "delete",
						DirPath:     targetDir,
						TargetPaths: pathsToDelete,
					}
				}
			}

		case PromptExitConfirm:
			fm.ActivePrompt = PromptNone
			return true, nil

		case PromptQuickCmd:
			cmdStr := strings.TrimSpace(fm.QuickCmdInput.Value())
			if cmdStr != "" {
				fm.ActivePrompt = PromptNone
				fm.QuickCmdInput.SetValue("")

				fields := strings.Fields(cmdStr)
				if len(fields) > 0 {
					firstWord := strings.ToLower(fields[0])
					interactiveCmds := map[string]bool{
						"vi": true, "vim": true, "nano": true, "emacs": true,
						"top": true, "htop": true, "less": true, "more": true,
						"man": true, "bash": true, "sh": true, "zsh": true, "su": true,
					}
					if interactiveCmds[firstWord] {
						fm.StatusMessage = fmt.Sprintf("⚠️ '%s' 대화형 프로그램은 내장 에디터(F4)나 터미널([t])을 이용해주세요! (추천 인라인: ls, df, du, chmod, tar, tail, grep)", firstWord)
						return false, nil
					}
				}

				dirPath := fm.RemotePath
				if fm.FocusLocal {
					dirPath = fm.LocalPath
				}
				return false, func() tea.Msg {
					return FileManagerQuickCmdMsg{
						HostID:  fm.HostID,
						IsLocal: fm.FocusLocal,
						DirPath: dirPath,
						Command: cmdStr,
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	if fm.ActivePrompt == PromptQuickCmd {
		fm.QuickCmdInput, cmd = fm.QuickCmdInput.Update(msg)
	} else {
		fm.SubInput, cmd = fm.SubInput.Update(msg)
	}
	return false, cmd
}

func filepathBase(p string) string {
	return filepath.Base(p)
}
