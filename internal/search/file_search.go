package search

import (
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

type FileItem struct {
	Path    string
	Starred bool
}

func (f FileItem) Title() string       { return filepath.Base(f.Path) }
func (f FileItem) Description() string { return "" }
func (f FileItem) FilterValue() string { return f.Path }

type FileSearch struct {
	Files    []string
	Query    string
	Selected []string
}

func NewFileSearch() (FileSearch, error) {
	out, err := exec.Command("fd", "--type", "f", "--strip-cwd-prefix").Output()
	if err != nil {
		return FileSearch{}, err
	}
	files := strings.Split(strings.TrimSpace(string(out)), "\n")
	return FileSearch{Files: files}, nil
}

func (fs *FileSearch) ToggleSelection(path string) {
	if slices.Contains(fs.Selected, path) {
		fs.Selected = slices.Delete(fs.Selected, slices.Index(fs.Selected, path), slices.Index(fs.Selected, path)+1)
	} else {
		fs.Selected = append(fs.Selected, path)
	}
}

func (fs *FileSearch) BuildItems() []FileItem {
	q := strings.ToLower(fs.Query)
	var items []FileItem

	for _, s := range fs.Selected {
		items = append(items, FileItem{Path: s, Starred: true})
	}

	for _, f := range fs.Files {
		if q == "" || strings.Contains(strings.ToLower(f), q) {
			if !slices.Contains(fs.Selected, f) {
				items = append(items, FileItem{Path: f})
			}
		}
	}
	return items
}

func (fs *FileSearch) RemoveSelection(path string) {
	if idx := slices.Index(fs.Selected, path); idx != -1 {
		fs.Selected = slices.Delete(fs.Selected, idx, idx+1)
	}
}
