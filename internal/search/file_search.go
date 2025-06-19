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

func SearchByContent(query string) ([]string, error) {
	if query == "" {
		return nil, nil
	}

	cmd := exec.Command("rg", "--files-with-matches", "--no-heading", "--smart-case", query)
	out, err := cmd.Output()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []string{}, nil
		}
		return nil, err
	}

	files := strings.Split(strings.TrimSpace(string(out)), "\n")
	return files, nil
}

func (fs *FileSearch) ToggleSelection(path string) {
	if slices.Contains(fs.Selected, path) {
		fs.Selected = slices.Delete(fs.Selected, slices.Index(fs.Selected, path), slices.Index(fs.Selected, path)+1)
	} else {
		fs.Selected = append(fs.Selected, path)
	}
}

func (fs *FileSearch) BuildItems(mode SearchMode) ([]FileItem, error) {
	q := strings.ToLower(fs.Query)
	var items []FileItem

	for _, s := range fs.Selected {
		items = append(items, FileItem{Path: s, Starred: true})
	}

	switch mode {
	case Filename:
		for _, f := range fs.Files {
			if q == "" || strings.Contains(strings.ToLower(f), q) {
				if !slices.Contains(fs.Selected, f) {
					items = append(items, FileItem{Path: f})
				}
			}
		}
	case Content:
		files, err := SearchByContent(fs.Query)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			if !slices.Contains(fs.Selected, f) {
				items = append(items, FileItem{Path: f})
			}
		}
	}

	return items, nil
}

func (fs *FileSearch) RemoveSelection(path string) {
	if idx := slices.Index(fs.Selected, path); idx != -1 {
		fs.Selected = slices.Delete(fs.Selected, idx, idx+1)
	}
}
