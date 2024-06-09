package backup

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type Format string

const (
	FormatTarGz Format = "tar.gz"
	FormatTar   Format = "tar"
	FormatZip   Format = "zip"
)

type Checksum struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type BackupMetadata struct {
	Date       time.Time `json:"timestamp"`
	FileFormat string    `json:"format"`
	FilePath   string    `json:"path"`
	Stores     []string  `json:"stores,omitempty"`
	Checksum   Checksum  `json:"checksum,omitempty"`
	Size       int64     `json:"size,omitempty"`
}

func (bm BackupMetadata) Timestamp() time.Time {
	return bm.Date
}

func (bm BackupMetadata) Format() string {
	return bm.FileFormat
}

func (bm BackupMetadata) Path() string {
	return bm.FilePath
}

type TarGzBackup struct {
	path      string
	size      int64
	complete  bool
	stores    []string
	checksum  Checksum
	timestamp time.Time
}

func (tgz *TarGzBackup) Format() string {
	return string(FormatTarGz)
}

func (tgz *TarGzBackup) Metadata() BackupMetadata {
	return BackupMetadata{
		FileFormat: tgz.Format(),
		FilePath:   tgz.path,
		Stores:     tgz.stores,
		Checksum:   tgz.checksum,
		Size:       tgz.size,
		Date:       tgz.timestamp,
	}
}

func NewTarGzBackup(inPath string, outPath string, stores []string, extraData ...[]byte) (*TarGzBackup, error) {
	stat, err := os.Stat(inPath)
	if err != nil {
		return nil, fmt.Errorf("error collecting files to backup: %w", err)
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("error collecting files to backup, not a directory: %s", stat.Name())
	}
	stat, err = os.Stat(outPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("error checking backup path: %w", err)
	}
	if err == nil && !stat.IsDir() {
		return nil, fmt.Errorf("error checking backup path, not a directory and file exists: %s", stat.Name())
	}
	if errors.Is(err, os.ErrNotExist) {
		if err = os.MkdirAll(outPath, 0755); err != nil {
			return nil, fmt.Errorf("error creating backup directory: %w", err)
		}
		stat, err = os.Stat(outPath)
		if err != nil {
			return nil, fmt.Errorf("error creating backup directory: %w", err)
		}
	}
	if stat.IsDir() {
		outPath = filepath.Join(filepath.Dir(outPath), filepath.Base(inPath)+".tar.gz")
	}
	tmpTar := outPath + ".tar.tmp"
	f, ferr := os.Create(tmpTar)
	if ferr != nil {
		return nil, fmt.Errorf("error creating backup file: %w", ferr)
	}
	tf := tar.NewWriter(f)
	if err = tf.AddFS(os.DirFS(inPath)); err != nil {
		return nil, fmt.Errorf("error adding files to backup: %w", err)
	}
	if err = tf.Close(); err != nil {
		return nil, fmt.Errorf("error closing backup tar file: %w", err)
	}
	if err = f.Sync(); err != nil {
		return nil, fmt.Errorf("error syncing backup tar file: %w", err)
	}
	if _, err = f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("error seeking to beginning of tar file: %w", err)
	}

	tfr := tar.NewReader(f)
	var entry *tar.Header

	var seen = make(map[string]bool)
	for _, storeName := range stores {
		seen[storeName] = false
	}

	for entry, err = tfr.Next(); err == nil; entry, err = tfr.Next() {
		if s, ok := seen[filepath.Dir(entry.Name)]; ok {
			if s {
				continue
			}
			seen[filepath.Dir(entry.Name)] = true
		}
	}

	if !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("error verifying backup tar file: %w", err)
	}

	for _, storeName := range stores {
		if !seen[storeName] {
			return nil, fmt.Errorf("store %s not found in backup", storeName)
		}
	}

	var finalFile *os.File

	if finalFile, err = os.Create(outPath); err != nil {
		return nil, fmt.Errorf("error opening final tar.gz file before writing: %w", err)
	}

	gz := gzip.NewWriter(finalFile)
	gz.Comment = "git.tcp.direct/tcp.direct/database backup archive"
	if len(extraData) > 0 {
		gz.Extra = append(gz.Extra, extraData[0]...)
	}
	if _, err = f.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("error seeking to beginning of tar file: %w", err)
	}
	if _, err = io.Copy(gz, f); err != nil {
		return nil, fmt.Errorf("error writing to final tar.gz file: %w", err)
	}
	if err = gz.Close(); err != nil {
		return nil, fmt.Errorf("error closing final tar.gz file: %w", err)
	}
	if err = f.Close(); err != nil {
		return nil, fmt.Errorf("error closing temporary tar file: %w", err)
	}
	_ = finalFile.Sync()
	if _, err = finalFile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("error seeking to beginning of final tar.gz file: %w", err)
	}

	summah := sha256.New()
	if _, err = io.Copy(summah, finalFile); err != nil {
		return nil, fmt.Errorf("error calculating checksum: %w", err)
	}
	if err = finalFile.Close(); err != nil {
		return nil, fmt.Errorf("error closing final tar.gz file: %w", err)
	}
	checksum := Checksum{
		Type:  "sha256",
		Value: fmt.Sprintf("%x", summah.Sum(nil)),
	}

	if err = os.Remove(tmpTar); err != nil {
		return nil, fmt.Errorf("error removing temporary tar file: %w", err)
	}

	return &TarGzBackup{
		path:      outPath,
		stores:    stores,
		timestamp: time.Now(),
		checksum:  checksum,
	}, nil

}
