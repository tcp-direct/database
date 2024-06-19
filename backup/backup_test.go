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
	t.Run("outpath_directory", func(t *testing.T) {
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

		if backup.Path() == "" {
			t.Error("expected a valid file path for the backup")
		}

		if len(backup.Stores) != 1 || backup.Stores[0] != "yeet" {
			t.Errorf("expected stores %v, got %v", stores, backup.Stores)
		}

		if backup.Checksum.Type != "sha256" || backup.Checksum.Value == "" {
			t.Errorf("expected a valid checksum, got %v", backup.Checksum)
		}

		tmp, err := os.ReadFile(backup.Path())
		if err != nil {
			t.Fatalf("error reading backup file: %v", err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(tmp)) != backup.Checksum.Value {
			t.Error("expected checksum to match the file")
		}

		if err = VerifyBackup(backup); err != nil {
			t.Fatalf("error verifying backup: %v", err)
		}

		t.Logf("backup metadata: %v", backup)
		t.Log(spew.Sdump(backup))
	})
	t.Run("outpath_file", func(t *testing.T) {
		inDir := t.TempDir()
		outDir := t.TempDir()
		outPath := filepath.Join(outDir, "backup.tar.gz")

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
		backup, err := NewTarGzBackup(inDir, outPath, stores)
		if err != nil {
			t.Fatalf("error creating tar.gz backup: %v", err)
		}
		if backup.Format() != string(FormatTarGz) {
			t.Errorf("expected format %s, got %s", FormatTarGz, backup.Format())
		}
		if backup.Path() == "" {
			t.Error("expected a valid file path for the backup")
		}
		if backup.Path() != filepath.Join(outDir, "backup.tar.gz") {
			t.Errorf("expected path %s, got %s", outPath, backup.Path())
		}
		if len(backup.Stores) != 1 || backup.Stores[0] != "yeet" {
			t.Errorf("expected stores %v, got %v", stores, backup.Stores)
		}
		if backup.Checksum.Type != "sha256" || backup.Checksum.Value == "" {
			t.Errorf("expected a valid checksum, got %v", backup.Checksum)
		}
		tmp, err := os.ReadFile(backup.Path())
		if err != nil {
			t.Fatalf("error reading backup file: %v", err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(tmp)) != backup.Checksum.Value {
			t.Error("expected checksum to match the file")
		}
		if err = VerifyBackup(backup); err != nil {
			t.Fatalf("error verifying backup: %v", err)
		}
		t.Logf("backup metadata: %v", backup)
	})

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

func TestRestoreTarGzBackup(t *testing.T) {
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

	if err = RestoreTarGzBackup(backup.Path(), outDir); err != nil {
		t.Fatalf("error restoring tar.gz backup: %v", err)
	}

	tmp, err := os.ReadFile(filepath.Join(outDir, "yeet", "sample.txt"))
	if err != nil {
		t.Fatalf("error reading restored file: %v", err)
	}
	if string(tmp) != "yeets" {
		t.Errorf("expected file contents yeets, got %s", tmp)
	}

	tmp, err = os.ReadFile(filepath.Join(outDir, "yeet", "sample2.txt"))
	if err != nil {
		t.Fatalf("error reading restored file: %v", err)
	}
	if string(tmp) != "yeets2" {
		t.Errorf("expected file contents yeets2, got %s", tmp)
	}
}
