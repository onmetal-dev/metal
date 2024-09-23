package ignorewalk

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"text/tabwriter"
)

func TestWalk(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "ignorewalk_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create files and directories
	files := []string{
		"file1.txt",
		"file2.log",
		"dir1/file3.txt",
		"dir1/file4.log",
		"dir1/dir2/file5.txt",
		"dir1/dir2/file6.log",
		"dir1/dir2/include.txt",
	}
	for _, file := range files {
		path := filepath.Join(tempDir, file)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		if _, err := os.Create(path); err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
	}

	// Create .gitignore files
	gitignoreRoot := `
*.log
dir1/file3.txt
`
	if err := os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitignoreRoot), 0644); err != nil {
		t.Fatalf("Failed to write .gitignore: %v", err)
	}

	gitignoreDir1 := `
file4.log
dir2/file5.txt
`
	if err := os.WriteFile(filepath.Join(tempDir, "dir1", ".gitignore"), []byte(gitignoreDir1), 0644); err != nil {
		t.Fatalf("Failed to write .gitignore: %v", err)
	}

	var walkedFiles []string
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(tempDir, path)
			if err != nil {
				return err
			}
			walkedFiles = append(walkedFiles, relPath)
		}
		return nil
	}

	if err := Walk(tempDir, walkFn, WithIgnoreFiles([]string{".gitignore"})); err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	expectedFiles := []string{
		".gitignore",
		"file1.txt",
		"dir1/.gitignore",
		"dir1/dir2/include.txt",
	}

	sort.Strings(walkedFiles)
	sort.Strings(expectedFiles)
	if len(walkedFiles) != len(expectedFiles) {
		var errorTable bytes.Buffer
		writer := tabwriter.NewWriter(&errorTable, 0, 0, 1, ' ', tabwriter.Debug)
		fmt.Fprintln(writer, "expected\tgot")
		for i := 0; i < int(math.Max(float64(len(expectedFiles)), float64(len(walkedFiles)))); i++ {
			if i < len(expectedFiles) {
				fmt.Fprintf(writer, "%s\t", expectedFiles[i])
			} else {
				fmt.Fprint(writer, "\t")
			}
			if i < len(walkedFiles) {
				fmt.Fprintf(writer, "%s\n", walkedFiles[i])
			} else {
				fmt.Fprint(writer, "\n")
			}
		}
		writer.Flush()
		t.Errorf("Expected %d files, but got %d\n%s", len(expectedFiles), len(walkedFiles), errorTable.String())
	}
}
