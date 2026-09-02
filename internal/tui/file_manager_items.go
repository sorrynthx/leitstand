package tui

import (
	"leitstand/internal/ssh"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m *FileManagerModal) RefreshLocalCmd() tea.Cmd {
	return func() tea.Msg {
		items, err := ssh.ListLocalDirectory(m.LocalPath, m.ShowHidden)
		return LocalFileListMsg{
			HostID:  m.HostID,
			Path:    m.LocalPath,
			OldPath: m.LocalPath,
			Items:   items,
			Err:     err,
		}
	}
}

func (m *FileManagerModal) SetFocusLocal(local bool) {
	m.FocusLocal = local
	if local {
		m.ActivePanel = PanelLocal
	} else {
		m.ActivePanel = PanelRemote
	}
}

func (m *FileManagerModal) GetVisibleItems(isLocal bool) []*ssh.FileItem {
	if isLocal {
		return m.GetSortedLocalItems()
	}
	return m.GetSortedRemoteItems()
}

func (m *FileManagerModal) GetActiveItems() []*ssh.FileItem {
	if m.ActivePanel == PanelLocal {
		return m.GetSortedLocalItems()
	}
	return m.GetSortedRemoteItems()
}

func (m *FileManagerModal) GetActiveCursor() int {
	if m.ActivePanel == PanelLocal {
		return m.LocalCursor
	}
	return m.RemoteCursor
}

func (m *FileManagerModal) GetSortedLocalItems() []*ssh.FileItem {
	var items []*ssh.FileItem
	for _, it := range m.LocalItems {
		if it.Name == ".." || m.LocalFilter == "" || strings.Contains(strings.ToLower(it.Name), strings.ToLower(m.LocalFilter)) {
			items = append(items, it)
		}
	}
	sortFileItems(items, m.LocalSort, m.LocalSortAsc)
	return items
}

func (m *FileManagerModal) GetSortedRemoteItems() []*ssh.FileItem {
	var items []*ssh.FileItem
	for _, it := range m.RemoteItems {
		items = append(items, it)
	}
	sortFileItems(items, m.RemoteSort, m.RemoteSortAsc)
	return items
}

func sortFileItems(items []*ssh.FileItem, field SortField, asc bool) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].IsDir != items[j].IsDir {
			return items[i].IsDir
		}
		if items[i].Name == ".." {
			return true
		}
		if items[j].Name == ".." {
			return false
		}

		var cmp bool
		switch field {
		case SortBySize:
			cmp = items[i].Size < items[j].Size
		case SortByModTime:
			cmp = items[i].ModTime.Before(items[j].ModTime)
		default:
			cmp = strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
		}

		if !asc {
			return !cmp
		}
		return cmp
	})
}

func (m *FileManagerModal) ToggleSelection() {
	items := m.GetActiveItems()
	cursor := m.GetActiveCursor()
	if cursor < 0 || cursor >= len(items) {
		return
	}
	item := items[cursor]
	if item.Name == ".." {
		return
	}

	if m.ActivePanel == PanelLocal {
		m.LocalSelected[item.Path] = !m.LocalSelected[item.Path]
		if !m.LocalSelected[item.Path] {
			delete(m.LocalSelected, item.Path)
		}
		if m.LocalCursor < len(items)-1 {
			m.LocalCursor++
		}
	} else {
		m.RemoteSelected[item.Path] = !m.RemoteSelected[item.Path]
		if !m.RemoteSelected[item.Path] {
			delete(m.RemoteSelected, item.Path)
		}
		if m.RemoteCursor < len(items)-1 {
			m.RemoteCursor++
		}
	}
}

func (m *FileManagerModal) SelectAll() {
	items := m.GetActiveItems()
	selMap := m.RemoteSelected
	if m.ActivePanel == PanelLocal {
		selMap = m.LocalSelected
	}

	allSelected := true
	for _, it := range items {
		if it.Name == ".." {
			continue
		}
		if !selMap[it.Path] {
			allSelected = false
			break
		}
	}

	if allSelected {
		for _, it := range items {
			delete(selMap, it.Path)
		}
	} else {
		for _, it := range items {
			if it.Name != ".." {
				selMap[it.Path] = true
			}
		}
	}
}

func (m *FileManagerModal) GetSelectedPaths(isLocal ...bool) []string {
	wantLocal := (m.ActivePanel == PanelLocal)
	if len(isLocal) > 0 {
		wantLocal = isLocal[0]
	}

	selMap := m.RemoteSelected
	if wantLocal {
		selMap = m.LocalSelected
	}

	var paths []string
	for p, checked := range selMap {
		if checked {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)

	if len(paths) == 0 {
		items := m.GetActiveItems()
		cursor := m.GetActiveCursor()
		if cursor >= 0 && cursor < len(items) {
			it := items[cursor]
			if it.Name != ".." {
				paths = append(paths, it.Path)
			}
		}
	}

	return paths
}

func (m *FileManagerModal) ClearSelections() {
	m.LocalSelected = make(map[string]bool)
	m.RemoteSelected = make(map[string]bool)
}
