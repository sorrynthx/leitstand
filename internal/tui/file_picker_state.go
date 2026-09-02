package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type FileItem struct {
	Name  string
	IsDir bool
	IsKey bool
	Size  int64
	Path  string
}

type FilePickerModal struct {
	currentDir    string
	items         []FileItem
	selectedIndex int
	width         int
	height        int
	errMessage    string
}

func NewFilePickerModal(initDir string, termWidth, termHeight int) *FilePickerModal {
	if initDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			initDir = filepath.Join(home, ".ssh")
			if _, err := os.Stat(initDir); os.IsNotExist(err) {
				initDir = home
			}
		} else {
			initDir = "."
		}
	}

	absDir, err := filepath.Abs(initDir)
	if err == nil {
		initDir = absDir
	}

	fp := &FilePickerModal{
		currentDir: initDir,
		width:      termWidth,
		height:     termHeight,
	}
	fp.loadDirectory(initDir)
	return fp
}

func (fp *FilePickerModal) loadDirectory(dirPath string) {
	fp.errMessage = ""
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		fp.errMessage = err.Error()
		fp.items = []FileItem{}
		fp.selectedIndex = 0
		return
	}

	var items []FileItem

	if parent := filepath.Dir(dirPath); parent != dirPath {
		items = append(items, FileItem{
			Name:  "..",
			IsDir: true,
			Path:  parent,
		})
	}

	var dirs []FileItem
	var files []FileItem

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") && name != ".ssh" {
			continue
		}

		fullPath := filepath.Join(dirPath, name)
		info, err := entry.Info()
		var size int64
		if err == nil {
			size = info.Size()
		}

		if entry.IsDir() {
			dirs = append(dirs, FileItem{
				Name:  name,
				IsDir: true,
				Size:  size,
				Path:  fullPath,
			})
		} else {
			isKey := isLikelyPrivateKey(name, fullPath)
			files = append(files, FileItem{
				Name:  name,
				IsDir: false,
				IsKey: isKey,
				Size:  size,
				Path:  fullPath,
			})
		}
	}

	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})

	sort.Slice(files, func(i, j int) bool {
		if files[i].IsKey != files[j].IsKey {
			return files[i].IsKey
		}
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	items = append(items, dirs...)
	items = append(items, files...)

	fp.items = items
	fp.currentDir = dirPath
	fp.selectedIndex = 0
}

func isLikelyPrivateKey(name, path string) bool {
	lower := strings.ToLower(name)
	if strings.HasSuffix(lower, ".pub") || strings.HasSuffix(lower, ".known_hosts") {
		return false
	}
	if strings.Contains(lower, "id_rsa") || strings.Contains(lower, "id_ed25519") ||
		strings.Contains(lower, "id_ecdsa") || strings.HasSuffix(lower, ".pem") ||
		strings.HasSuffix(lower, ".key") {
		return true
	}

	content, err := os.ReadFile(path)
	if err == nil {
		s := string(content)
		if strings.Contains(s, "PRIVATE KEY-----") {
			return true
		}
	}
	return false
}
