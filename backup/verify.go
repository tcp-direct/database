package backup

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
)

func VerifyBackup(metadata BackupMetadata, path string) error {
	switch metadata.Format() {
	case "tar.gz":
		//
	default:
		return errors.New("unsupported backup format")
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening backup file: %w", err)
	}

	var hasher hash.Hash

	defer file.Close()
	switch metadata.Checksum.Type {
	case "sha256":
		hasher = sha256.New()
	case "sha512":
		hasher = sha512.New()
	case "md5":
		hasher = md5.New()
	case "sha1":
		hasher = sha1.New()
	default:
		return fmt.Errorf("unsupported checksum type: %s", metadata.Checksum.Type)
	}

	if _, err = io.Copy(hasher, file); err != nil {
		return fmt.Errorf("error hashing backup file: %w", err)
	}

	if metadata.Checksum.Value != fmt.Sprintf("%x", hasher.Sum(nil)) {
		return fmt.Errorf("checksums do not match")
	}

	return nil
}
