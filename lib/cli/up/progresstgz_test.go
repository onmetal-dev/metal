package up

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProgressTargzipper(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "progresstgz_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test file content
	fileContent := bytes.Repeat([]byte("test data"), 1000) // 9000 bytes

	// Create test directory structure and files
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

	// Create a buffer to capture the tar.gz data
	var buf bytes.Buffer

	// Create progress callback
	var progressValues []float64
	onProgress := func(progress float64) {
		progressValues = append(progressValues, progress)
	}

	// Create and start the progressTargzipper
	targzipper, err := NewProgressTargzipper(tempDir, &buf, onProgress)
	require.NoError(t, err)

	err = targzipper.Start()
	require.NoError(t, err)

	// Check progress values
	assert.GreaterOrEqual(t, len(progressValues), 5, "Expected at least 5 progress updates")
	assert.InDelta(t, 0.0, progressValues[0], 0.1, "First progress should be close to 0.0")
	assert.InDelta(t, 0.25, progressValues[len(progressValues)/4], 0.1, "Progress should be close to 0.25")
	assert.InDelta(t, 0.5, progressValues[len(progressValues)/2], 0.1, "Progress should be close to 0.50")
	assert.InDelta(t, 0.75, progressValues[len(progressValues)*3/4], 0.1, "Progress should be close to 0.75")
	assert.InDelta(t, 1.0, progressValues[len(progressValues)-1], 0.1, "Last progress should be close to 1.0")

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
		assert.Equal(t, fileContent, content, "Extracted file content should match original")
	}
}
