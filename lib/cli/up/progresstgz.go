package up

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"

	"github.com/onmetal-dev/metal/lib/ignorewalk"
)

type progressTargzipper struct {
	sourcePath string
	writer     io.Writer
	total      int64
	processed  int64
	onProgress func(float64)
	gzipWriter *gzip.Writer
	tarWriter  *tar.Writer
}

func NewProgressTargzipper(sourcePath string, writer io.Writer, onProgress func(float64)) (*progressTargzipper, error) {
	total, err := calculateTotalSize(sourcePath)
	if err != nil {
		return nil, err
	}

	gzipWriter := gzip.NewWriter(writer)
	tarWriter := tar.NewWriter(gzipWriter)

	return &progressTargzipper{
		sourcePath: sourcePath,
		writer:     writer,
		total:      total,
		onProgress: onProgress,
		gzipWriter: gzipWriter,
		tarWriter:  tarWriter,
	}, nil
}

func (pt *progressTargzipper) Start() error {
	defer pt.gzipWriter.Close()
	defer pt.tarWriter.Close()

	pt.onProgress(0.0)
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

		if err := pt.tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !info.Mode().IsRegular() {
			return nil
		}

		//if !info.IsDir() {
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		n, err := io.Copy(pt.tarWriter, file)
		if err != nil {
			return err
		}

		pt.processed += n
		if pt.onProgress != nil {
			pt.onProgress(float64(pt.processed) / float64(pt.total))
		}
		//}

		return nil
	}, ignorewalk.WithIgnoreFiles([]string{".gitignore", ".dockerignore"})); err != nil {
		return err
	}
	pt.onProgress(1.0)
	return nil
}

func calculateTotalSize(sourcePath string) (int64, error) {
	var total int64

	err := filepath.Walk(sourcePath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})

	return total, err
}
