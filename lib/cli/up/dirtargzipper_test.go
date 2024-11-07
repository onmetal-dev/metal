package up

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
)

func TestDirTargzipper(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "progresstgz_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// fill test directory with some files and directories
	fileContent := bytes.Repeat([]byte("test data"), 1000) // 9000 bytes
	testFiles := []string{
		"file1.txt",
		"subdir1/file2.txt",
		"subdir2/file3.txt",
		"subdir2/subdir3/file4.txt",
	}
	for _, filePath := range testFiles {
		fullPath := filepath.Join(tempDir, filePath)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, fileContent, 0644)
		require.NoError(t, err)
	}

	var buf bytes.Buffer
	targzipper, err := NewDirTargzipper(tempDir, &buf)
	require.NoError(t, err)
	tgzIter := targzipper.Run()

	progressValues := []float64{}
	for progress, err := range tgzIter {
		require.NoError(t, err)
		progressValues = append(progressValues, progress.Percentage)
	}

	// Check progress values
	require.Equal(t, len(progressValues), 5, "Expected 5 progress updates: %v", progressValues)
	require.InDeltaSlice(t, []float64{0.0, 0.25, 0.50, 0.75, 1.0}, lo.Map(progressValues, func(p float64, _ int) float64 {
		return p
	}), 0.05)

	// Create a temporary directory for extraction
	extractDir, err := os.MkdirTemp("", "progresstgz_extract")
	require.NoError(t, err)
	defer os.RemoveAll(extractDir)

	// Extract the tar.gz data
	gzr, err := gzip.NewReader(bytes.NewReader(buf.Bytes()))
	require.NoError(t, err)
	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)

		target := filepath.Join(extractDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			err = os.MkdirAll(target, 0755)
			require.NoError(t, err)
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			require.NoError(t, err)
			_, err = io.Copy(f, tr)
			require.NoError(t, err)
			f.Close()
		}
	}

	// Verify extracted files
	for _, filePath := range testFiles {
		extractedPath := filepath.Join(extractDir, filePath)
		content, err := os.ReadFile(extractedPath)
		require.NoError(t, err)
		require.Equal(t, fileContent, content, "Extracted file content should match original")
	}
}
