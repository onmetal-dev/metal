package ignorewalk

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/plumbing/format/gitignore"
)

type walker struct {
	ignoreFiles []string
	patterns    []gitignore.Pattern
}

func (w *walker) walk(root string, fn filepath.WalkFunc) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if w.shouldIgnore(path, info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// If it's a directory, load its ignore file(s)
		if info.IsDir() {
			for _, ignoreFile := range w.ignoreFiles {
				ps, err := readIgnoreFile(path, ignoreFile)
				if err != nil {
					return err
				} else if len(ps) > 0 {
					w.patterns = append(w.patterns, ps...)
				}
			}
		}

		return fn(path, info, err)
	})
}

func (w *walker) shouldIgnore(path string, isDir bool) bool {
	segments := strings.Split(path, string(os.PathSeparator))
	for _, pattern := range w.patterns {
		match := pattern.Match(segments, isDir)
		if match != gitignore.NoMatch {
			return match == gitignore.Exclude
		}
	}
	return false
}

type WalkOption func(w *walker) error

// WithIgnoreFiles specifies filename to treat as ignore files. E.g., [".gitignore", ".dockerignore"].
func WithIgnoreFiles(ignoreFiles []string) WalkOption {
	return func(w *walker) error {
		w.ignoreFiles = ignoreFiles
		return nil
	}
}

func Walk(root string, fn filepath.WalkFunc, walkOpt ...WalkOption) error {
	w := &walker{}
	for _, opt := range walkOpt {
		if err := opt(w); err != nil {
			return err
		}
	}

	// Load the root ignore file(s)
	for _, ignoreFile := range w.ignoreFiles {
		ps, err := readIgnoreFile(root, ignoreFile)
		if err != nil {
			return err
		} else if len(ps) > 0 {
			w.patterns = append(w.patterns, ps...)
		}
	}
	return w.walk(root, fn)
}

const commentPrefix = "#"

// readIgnoreFile tries to read a specific ignore file.
// It returns nil, nil if the file does not exist.
func readIgnoreFile(path string, ignoreFile string) (ps []gitignore.Pattern, err error) {
	f, err := os.Open(filepath.Join(path, ignoreFile))
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		return nil, nil
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		s := scanner.Text()
		if !strings.HasPrefix(s, commentPrefix) && len(strings.TrimSpace(s)) > 0 {
			ps = append(ps, gitignore.ParsePattern(s, strings.Split(path, string(os.PathSeparator))))
		}
	}
	return
}
