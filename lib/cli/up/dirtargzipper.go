package up

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"io"
	"iter"
	"math"
	"os"
	"path/filepath"

	"github.com/onmetal-dev/metal/lib/ignorewalk"
)

// DirTargzipper targz's a directory and respects ignore files.
// By default, it will respect ignore patterns present in any .gitignore and .dockerignore files.
type DirTargzipper struct {
	sourcePath  string
	writer      io.Writer
	ignoreFiles []string
	total       int64
}

// DirTargzipperOption configures a DirTargzipper.
type DirTargzipperOption func(*DirTargzipper)

// WithIgnoreFiles configures which ignore files to respect during compression.
func WithIgnoreFiles(ignoreFiles []string) DirTargzipperOption {
	return func(dt *DirTargzipper) {
		dt.ignoreFiles = ignoreFiles
	}
}

// NewDirTargzipper creates a new targzipper that compresses a directory.
func NewDirTargzipper(sourcePath string, writer io.Writer, opts ...DirTargzipperOption) (*DirTargzipper, error) {
	dt := &DirTargzipper{
		sourcePath:  sourcePath,
		writer:      writer,
		ignoreFiles: []string{".gitignore", ".dockerignore"}, // Default ignore files
	}
	for _, opt := range opts {
		opt(dt)
	}
	total, err := calculateTotalSize(dt.sourcePath, dt.ignoreFiles)
	if err != nil {
		return nil, err
	}
	dt.total = total
	return dt, nil
}

// Progress of the compression.
type Progress struct {
	// Percentage of the operation that is complete.
	Percentage float64
	// Processed bytes thus far.
	Processed int64
	// Total number of bytes to process.
	Total int64
	// Filename being processed. Could be empty if this is the begining or end of the compression operation.
	Filename string
	// Done is true when compression is complete.
	Done bool
}

// Run the compression operation. The iterator returned yields information about the progress of the operation.
func (pt *DirTargzipper) Run() iter.Seq2[Progress, error] {
	gzipWriter := gzip.NewWriter(pt.writer)
	tarWriter := tar.NewWriter(gzipWriter)
	processed := int64(0)
	return func(yield func(Progress, error) bool) {
		defer gzipWriter.Close()
		defer tarWriter.Close()
		if err := ignorewalk.Walk(pt.sourcePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			var link string
			if info.Mode()&os.ModeSymlink != 0 {
				if link, err = os.Readlink(path); err != nil {
					return err
				}
			}

			header, err := tar.FileInfoHeader(info, link)
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(pt.sourcePath, path)
			if err != nil {
				return err
			}
			header.Name = relPath

			if err := tarWriter.WriteHeader(header); err != nil {
				return err
			}

			if !info.Mode().IsRegular() {
				return nil
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if (!yield(Progress{
				Percentage: math.Min(1.0, float64(processed)/float64(pt.total)),
				Processed:  processed,
				Total:      pt.total,
				Filename:   file.Name(),
			}, nil)) {
				return errors.New("iterator stopped")
			}

			n, err := io.Copy(tarWriter, file)
			if err != nil {
				return err
			}

			processed += n
			return nil
		}, ignorewalk.WithIgnoreFiles(pt.ignoreFiles)); err != nil {
			yield(Progress{}, err)
			return
		}
		yield(Progress{
			Percentage: 1.0,
			Processed:  processed,
			Total:      pt.total,
			Done:       true,
		}, nil)
	}
}

func calculateTotalSize(sourcePath string, ignoreFiles []string) (int64, error) {
	var total int64
	return total, ignorewalk.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		total += info.Size()
		return nil
	}, ignorewalk.WithIgnoreFiles(ignoreFiles))
}
