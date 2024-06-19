package backup

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"git.tcp.direct/tcp.direct/database/models"
)

type Format string

var _ models.Backup = &BackupMetadata{}

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

func (bm BackupMetadata) MarshalJSON() ([]byte, error) {
	mdat := map[string]interface{}{
		"timestamp": bm.Date,
		"format":    bm.FileFormat,
		"path":      bm.FilePath,
		"stores":    bm.Stores,
		"checksum":  bm.Checksum,
	}
	if bm.Size > 0 {
		mdat["size"] = bm.Size
	}
	return json.Marshal(mdat)
}

func (bm BackupMetadata) Type() string {
	return bm.FileFormat
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

func NewTarGzBackup(inPath string, outPath string, stores []string, extraData ...[]byte) (BackupMetadata, error) {
	stat, err := os.Stat(inPath)
	nilBackup := BackupMetadata{}
	if err != nil {
		return nilBackup, fmt.Errorf("error collecting files to backup: %w", err)
	}
	if !stat.IsDir() {
		return nilBackup, fmt.Errorf("error collecting files to backup, not a directory: %s", stat.Name())
	}
	stat, err = os.Stat(outPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nilBackup, fmt.Errorf("error checking backup path: %w", err)
	}
	if stat != nil && stat.IsDir() {
		outPath = filepath.Join(outPath, filepath.Base(inPath)+".tar.gz")
	}

	tmpTar := outPath + ".tar.tmp"
	f, ferr := os.Create(tmpTar)
	if ferr != nil {
		return nilBackup, fmt.Errorf("error creating temporary tar file: %w", ferr)
	}
	tf := tar.NewWriter(f)
	if err = tf.AddFS(os.DirFS(inPath)); err != nil {
		return nilBackup, fmt.Errorf("error adding files to backup: %w", err)
	}
	if err = tf.Close(); err != nil {
		return nilBackup, fmt.Errorf("error closing backup tar file: %w", err)
	}
	if err = f.Sync(); err != nil {
		return nilBackup, fmt.Errorf("error syncing backup tar file: %w", err)
	}
	if _, err = f.Seek(0, 0); err != nil {
		return nilBackup, fmt.Errorf("error seeking to beginning of tar file: %w", err)
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
		return nilBackup, fmt.Errorf("error verifying backup tar file: %w", err)
	}

	for _, storeName := range stores {
		if !seen[storeName] {
			return nilBackup, fmt.Errorf("store %s not found in backup", storeName)
		}
	}

	buffer := make([]byte, 1024)

	var finalFile *os.File

	if finalFile, err = os.Create(outPath); err != nil {
		return nilBackup, fmt.Errorf("error opening final tar.gz file before writing: %w", err)
	}

	gz := gzip.NewWriter(finalFile)
	gz.Comment = "git.tcp.direct/tcp.direct/database backup archive"
	if len(extraData) > 0 {
		for _, data := range extraData {
			gz.Comment += "\n" + string(data)
		}
	}
	if _, err = f.Seek(0, 0); err != nil {
		return nilBackup, fmt.Errorf("error seeking to beginning of tar file: %w", err)
	}
	if _, err = io.CopyBuffer(gz, f, buffer); err != nil {
		return nilBackup, fmt.Errorf("error writing to final tar.gz file: %w", err)
	}
	if err = gz.Close(); err != nil {
		return nilBackup, fmt.Errorf("error closing final tar.gz file: %w", err)
	}
	if err = f.Close(); err != nil {
		return nilBackup, fmt.Errorf("error closing temporary tar file: %w", err)
	}
	_ = finalFile.Sync()
	if _, err = finalFile.Seek(0, 0); err != nil {
		return nilBackup, fmt.Errorf("error seeking to beginning of final tar.gz file: %w", err)
	}

	summah := sha256.New()
	if _, err = io.CopyBuffer(summah, finalFile, buffer); err != nil {
		return nilBackup, fmt.Errorf("error calculating checksum: %w", err)
	}
	if err = finalFile.Close(); err != nil {
		return nilBackup, fmt.Errorf("error closing final tar.gz file: %w", err)
	}
	checksum := Checksum{
		Type:  "sha256",
		Value: fmt.Sprintf("%x", summah.Sum(nil)),
	}

	if err = os.Remove(f.Name()); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nilBackup, fmt.Errorf("error removing temporary tar file: %w", err)
	}

	tgz := &TarGzBackup{
		path:      outPath,
		stores:    stores,
		timestamp: time.Now(),
		checksum:  checksum,
	}

	return tgz.Metadata(), nil
}

func RestoreTarGzBackup(inPath string, outPath string) error {
	stat, err := os.Stat(inPath)
	if err != nil {
		return fmt.Errorf("error checking backup file: %w", err)
	}
	if stat.IsDir() {
		return fmt.Errorf("error checking backup file, not a file: %s", stat.Name())
	}
	f, ferr := os.Open(inPath)

	defer func() {
		_ = f.Close()
	}()

	if ferr != nil {
		return fmt.Errorf("error opening backup file: %w", ferr)
	}
	gz, gerr := gzip.NewReader(f)
	if gerr != nil {
		return fmt.Errorf("error creating gzip reader: %w", gerr)
	}

	buf := make([]byte, 1024)

	tfr := tar.NewReader(gz)
	var entry *tar.Header

	for {
		entry, err = tfr.Next()
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("error reading tar file: %w", err)
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if entry == nil {
			break
		}
		if !filepath.IsLocal(entry.Name) {
			return fmt.Errorf("tar file contains invalid path: %s", entry.Name)
		}
		switch entry.Typeflag {
		case tar.TypeDir:
			if err = os.MkdirAll(filepath.Join(outPath, entry.Name), 0755); err != nil {
				return fmt.Errorf("error creating directory: %w", err)
			}
		case tar.TypeReg:
			var file *os.File
			dirStat, dirErr := os.Stat(filepath.Dir(filepath.Join(outPath, filepath.Dir(entry.Name))))
			if errors.Is(dirErr, os.ErrNotExist) {
				if err = os.MkdirAll(filepath.Join(outPath, filepath.Dir(entry.Name)), 0755); err != nil {
					return fmt.Errorf("error creating directory: %w", err)
				}
			}
			if !errors.Is(dirErr, os.ErrNotExist) && dirErr != nil {
				return fmt.Errorf("error checking output directory: %w", dirErr)
			}
			if dirStat != nil && !dirStat.IsDir() {
				return fmt.Errorf("directory in backup exists in outpath as a file: %s", filepath.Dir(entry.Name))
			}
			if file, err = os.Create(filepath.Join(outPath, entry.Name)); err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("error creating file %s: %w", entry.Name, err)
				}
				if err = os.MkdirAll(filepath.Dir(filepath.Join(outPath, entry.Name)), 0755); err != nil {
					return fmt.Errorf("error creating directory: %w", err)
				}
				if file, err = os.Create(filepath.Join(outPath, entry.Name)); err != nil {
					return fmt.Errorf("error creating file %s: %w", entry.Name, err)
				}
			}
			if _, err = io.CopyBuffer(file, tfr, buf); err != nil {
				_ = file.Close()
				return fmt.Errorf("error writing file: %w", err)
			}
			if err = file.Close(); err != nil {
				return fmt.Errorf("error closing file (%s): %w", file.Name(), err)
			}
		default:
			return fmt.Errorf("unsupported tar file type: %c", entry.Typeflag)
		}
	}

	return nil
}
