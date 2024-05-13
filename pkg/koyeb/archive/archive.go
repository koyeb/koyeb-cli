// This package provides the ability to compress a directory into a tarball. It
// is used by the CLI to upload archives to Koyeb.

package archive

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type tarball struct {
	File *os.File
}

func (t *tarball) Close() error {
	log.Debugf("Remove temporary archive file %s", t.File.Name())
	if err := os.Remove(t.File.Name()); err != nil {
		return err
	}
	return t.File.Close()
}

var ignoredDirectories = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
}

// Archive compresses a directory into a tarball and returns the path to this tarball.
// Some directories are ignored by default (e.g. .git, node_modules, vendor).
// This is not yet configurable but could be in the future.
func Archive(path string) (*tarball, error) {
	basePath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	tempFile, err := os.CreateTemp("", "koyeb-archive-*.tar.gz")
	if err != nil {
		return nil, err
	}
	log.Debugf("Create temporary archive file %s", tempFile.Name())

	tarball := tarball{tempFile}

	// Compress the tarball
	gzipWriter := gzip.NewWriter(tempFile)
	defer gzipWriter.Close()

	// Create a new tar archive
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	if err := filepath.Walk(basePath, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignore directories that should not be included in the archive. We
		// only match the base name of the directory, there is no yet support
		// for regex or more complex patterns, that could be useful to ignore
		// all the files from .gitignore or .dockerignore for example
		if fi.IsDir() && ignoredDirectories[filepath.Base(file)] {
			log.Debugf("Archive: skip %s", file)
			return filepath.SkipDir
		}

		log.Debugf("Archive %s", file)

		relativePath, err := filepath.Rel(basePath, file)
		if err != nil {
			return fmt.Errorf("Unable to get relative path for file '%s': %w", file, err)
		}

		// Create header
		header, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return fmt.Errorf("Unable to create header for file '%s': %w", file, err)
		}

		header.Name = filepath.ToSlash(relativePath)

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("Unable to write header for file '%s': %w", file, err)
		}

		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("Unable to open file '%s': %w", file, err)
			}
			defer data.Close()

			if _, err := io.Copy(tarWriter, data); err != nil {
				return fmt.Errorf("Unable to copy file '%s' into the tarball: %w", file, err)
			}
		}
		return nil
	}); err != nil {
		tarball.Close() // Remove the temporary file in case of error
		return nil, err
	}
	return &tarball, nil
}
