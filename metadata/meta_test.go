package metadata

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"git.tcp.direct/tcp.direct/database/models"
)

func TestNewMetaFile_Success(t *testing.T) {
	path := t.TempDir()
	keeperType := "testType"

	meta, err := NewMetaFile(keeperType, path)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if meta == nil {
		t.Fatal("expected a valid meta file")
	}
	if meta.Type() != keeperType {
		t.Errorf("expected type %s, got %s", keeperType, meta.Type())
	}
	if meta.Created.IsZero() {
		t.Error("expected a valid created time")
	}
	if meta.LastOpened.IsZero() {
		t.Error("expected a valid last opened time")
	}
}

func TestNewMetaFile_Fail_InvalidPath(t *testing.T) {
	path := "/invalid/path"
	keeperType := "testType"

	meta, err := NewMetaFile(keeperType, path)

	if err == nil {
		t.Error("expected an error, got nil")
	}
	if meta != nil {
		t.Error("expected no meta, got one")
	}
}

func TestOpenMetaFile_Success(t *testing.T) {
	path := t.TempDir()
	keeperType := "testType"

	if _, err := NewMetaFile(keeperType, path); err != nil {
		t.Fatalf("error creating meta file: %v", err)
	}
	meta, err := OpenMetaFile(filepath.Join(path, "meta.json"))

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if meta == nil {
		t.Fatal("expected a valid meta file")
	}
	if meta.Type() != keeperType {
		t.Errorf("expected type %s, got %s", keeperType, meta.Type())
	}

	metaDat, readErr := os.ReadFile(filepath.Join(path, "meta.json"))
	if readErr != nil {
		t.Fatalf("error reading meta.json: %v", readErr)
	}
	if meta, err = LoadMeta(metaDat); err != nil {
		t.Fatalf("error loading meta: %v", err)
	}
}

func TestOpenMetaFile_Fail_FileNotFound(t *testing.T) {
	path := "/invalid/path"

	meta, err := OpenMetaFile(path)

	if err == nil {
		t.Error("expected an error, got nil")
	}
	if meta != nil {
		t.Error("expected no meta, got one")
	}
}

func TestMetadata_AddStore(t *testing.T) {
	meta := NewMeta("testType")
	store := "testStore"

	meta.AddStore(store)

	if len(meta.KnownStores) != 1 || meta.KnownStores[0] != store {
		t.Errorf("expected %s in KnownStores, got %v", store, meta.KnownStores)
	}
}

func TestMetadata_RemoveStore(t *testing.T) {
	meta := NewMeta("testType")
	store := "testStore"

	meta.AddStore(store)
	meta.RemoveStore(store)

	if len(meta.KnownStores) != 0 {
		t.Errorf("expected KnownStores to be empty, got %v", meta.KnownStores)
	}
}

func TestMetadata_Ping(t *testing.T) {
	meta := NewMeta("testType")
	timeBeforePing := time.Now()

	meta.Ping()

	if meta.LastOpened.Before(timeBeforePing) {
		t.Errorf("expected LastOpened to be after %v, got %v", timeBeforePing, meta.LastOpened)
	}
}

func TestMetadata_Sync(t *testing.T) {
	path := t.TempDir()
	meta, err := NewMetaFile("testType", path)
	if err != nil {
		t.Fatalf("error creating meta file: %v", err)
	}

	err = meta.Sync()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMetadata_Close(t *testing.T) {
	path := t.TempDir()
	meta, err := NewMetaFile("testType", path)
	if err != nil {
		t.Fatalf("error creating meta file: %v", err)
	}

	err = meta.Close()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

type ClosableSeekableBuffer struct {
	*bytes.Buffer
	closed bool
}

func (cb *ClosableSeekableBuffer) Close() error {
	cb.closed = true
	return nil
}

func (cb *ClosableSeekableBuffer) Seek(offset int64, whence int) (int64, error) {
	if offset != 0 || whence != 0 {
		panic("unexpected seek")
	}
	cb.Buffer.Reset()
	return 0, nil
}

func TestMetadata_WithWriter(t *testing.T) {
	meta := NewMeta("testType")
	buf := new(bytes.Buffer)
	cbuf := &ClosableSeekableBuffer{Buffer: buf}

	meta.WithWriter(cbuf)

	if meta.w != cbuf {
		t.Error("expected writer to be set")
	}
	if cbuf.closed {
		t.Error("expected writer to not be closed")
	}
	if err := meta.Sync(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if cbuf.closed {
		t.Error("expected writer to not be closed after sync")
	}
	if err := meta.Close(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !cbuf.closed {
		t.Error("expected writer to be closed after close")

	}
}

type testBackup struct {
}

func (tb *testBackup) Format() string {
	return "test"
}

func (tb *testBackup) Metadata() models.Metadata {
	m := NewMeta("test")
	return m
}

func (tb *testBackup) Path() string {
	return "test"
}

func (tb *testBackup) Timestamp() time.Time {
	return time.Now()
}

func (tb *testBackup) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Format    string    `json:"format"`
		Timestamp time.Time `json:"timestamp"`
		Path      string    `json:"path"`
	}{
		Format:    tb.Format(),
		Timestamp: tb.Timestamp(),
		Path:      tb.Path(),
	})
}

func TestMetadata_WithBackups(t *testing.T) {
	meta := NewMeta("testType")
	backup := &testBackup{}

	meta.WithBackups(backup)

	if len(meta.Backups) != 1 {
		t.Error("expected one backup")
	}
}

func TestMetadata_WithLastOpened(t *testing.T) {
	meta := NewMeta("testType")
	lastOpened := time.Now()

	meta.WithLastOpened(lastOpened)

	if !meta.LastOpened.Equal(lastOpened) {
		t.Error("expected LastOpened to be set")
	}
}

func TestMetadata_WithCreated(t *testing.T) {
	meta := NewMeta("testType")
	created := time.Now()

	meta.WithCreated(created)

	if !meta.Created.Equal(created) {
		t.Error("expected Created to be set")
	}
}

func TestMetadata_WithStores(t *testing.T) {
	meta := NewMeta("testType")
	stores := []string{"store1", "store2"}

	meta.WithStores(stores...)

	if len(meta.KnownStores) != len(stores) {
		t.Error("expected KnownStores to be set")
	}
}
