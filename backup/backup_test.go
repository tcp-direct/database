package backup

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

func TestNewTarGzBackup(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	sampleDir := filepath.Join(inDir, "yeet")
	sampleFile1 := filepath.Join(sampleDir, "sample.txt")
	sampleFile2 := filepath.Join(sampleDir, "sample2.txt")
	if err := os.Mkdir(sampleDir, 0755); err != nil {
		t.Fatalf("error creating sample directory: %v", err)
	}
	err := os.WriteFile(sampleFile1, []byte("yeets"), 0644)
	if err != nil {
		t.Fatalf("error creating sample file: %v", err)
	}
	err = os.WriteFile(sampleFile2, []byte("yeets2"), 0644)
	if err != nil {
		t.Fatalf("error creating sample file: %v", err)
	}

	stores := []string{"yeet"}

	backup, err := NewTarGzBackup(inDir, outDir, stores)
	if err != nil {
		t.Fatalf("error creating tar.gz backup: %v", err)
	}

	if backup.Format() != string(FormatTarGz) {
		t.Errorf("expected format %s, got %s", FormatTarGz, backup.Format())
	}

	if backup.Metadata().FilePath == "" {
		t.Error("expected a valid file path for the backup")
	}

	if len(backup.Metadata().Stores) != 1 || backup.Metadata().Stores[0] != "yeet" {
		t.Errorf("expected stores %v, got %v", stores, backup.Metadata().Stores)
	}

	if backup.Metadata().Checksum.Type != "sha256" || backup.Metadata().Checksum.Value == "" {
		t.Errorf("expected a valid checksum, got %v", backup.Metadata().Checksum)
	}

	tmp, err := os.ReadFile(backup.Metadata().Path())
	if err != nil {
		t.Fatalf("error reading backup file: %v", err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(tmp)) != backup.Metadata().Checksum.Value {
		t.Error("expected checksum to match the file")
	}

	if err = VerifyBackup(backup.Metadata(), backup.Metadata().Path()); err != nil {
		t.Fatalf("error verifying backup: %v", err)
	}

	t.Logf("backup metadata: %v", backup.Metadata())
	t.Log(spew.Sdump(backup))

}

func TestTarGzBackup_Metadata(t *testing.T) {
	timestamp := time.Now()
	checksum := Checksum{Type: "sha256", Value: "dummychecksum"}
	tgz := &TarGzBackup{
		path:      "dummy/path",
		size:      1024,
		complete:  true,
		stores:    []string{"store1", "store2"},
		checksum:  checksum,
		timestamp: timestamp,
	}

	meta := tgz.Metadata()

	if meta.FileFormat != string(FormatTarGz) {
		t.Errorf("expected format %s, got %s", FormatTarGz, meta.FileFormat)
	}

	if meta.FilePath != "dummy/path" {
		t.Errorf("expected path dummy/path, got %s", meta.FilePath)
	}

	if len(meta.Stores) != 2 || meta.Stores[0] != "store1" || meta.Stores[1] != "store2" {
		t.Errorf("expected stores [store1 store2], got %v", meta.Stores)
	}

	if meta.Checksum != checksum {
		t.Errorf("expected checksum %v, got %v", checksum, meta.Checksum)
	}

	if !meta.Date.Equal(timestamp) {
		t.Errorf("expected timestamp %v, got %v", timestamp, meta.Date)
	}
}
