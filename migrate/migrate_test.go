package migrate

import (
	"errors"
	"testing"

	"git.tcp.direct/tcp.direct/database/test"
)

func TestMigrator_WithClobber(t *testing.T) {
	from := database.NewMockKeeper("yeeeties")
	to := database.NewMockKeeper("yooties")

	migrator, err := NewMigrator(from, to)
	if err != nil {
		t.Fatalf("error creating migrator: %v", err)
	}
	migrator = migrator.WithClobber()

	if !migrator.clobber {
		t.Error("expected clobber to be true")
	}
}

func TestMigrator_WithSkipExisting(t *testing.T) {
	from := database.NewMockKeeper("yeeeties")
	to := database.NewMockKeeper("yooties")

	migrator, err := NewMigrator(from, to)
	if err != nil {
		t.Fatalf("error creating migrator: %v", err)
	}
	migrator = migrator.WithSkipExisting()

	if !migrator.skipExisting {
		t.Error("expected skipExisting to be true")
	}
}

func TestMigrator_CheckDupes_NoStores(t *testing.T) {
	from := database.NewMockKeeper("yeeeties")
	to := database.NewMockKeeper("yooties")

	migrator, err := NewMigrator(from, to)
	if err != nil {
		t.Fatalf("error creating migrator: %v", err)
	}

	err = migrator.CheckDupes()

	if !errors.Is(err, ErrNoStores) {
		t.Error("expected ErrNoStores error")
	}
}

func TestMigrator_CheckDupes_DuplicateKeys(t *testing.T) {
	from := database.NewMockKeeper("yeeeties")
	to := database.NewMockKeeper("yooties")

	if err := from.WithNew("store1").Put([]byte("key1"), []byte("value1")); err != nil {
		t.Fatalf("error putting key1: %v", err)
	}
	if err := to.WithNew("store1").Put([]byte("key1"), []byte("value1")); err != nil {
		t.Fatalf("error putting key1: %v", err)
	}

	migrator, err := NewMigrator(from, to)
	if err != nil {
		t.Fatalf("error creating migrator: %v", err)
	}

	err = migrator.CheckDupes()

	if !errors.Is(err, ErrDupKeys) {
		t.Error("expected ErrDuplicateKeys error")
	}
}

func TestMigrator_Success(t *testing.T) {
	from := database.NewMockKeeper("yeeeties")
	to := database.NewMockKeeper("yooties")

	if err := from.WithNew("store1").Put([]byte("key1"), []byte("value1")); err != nil {
		t.Fatalf("error putting key1: %v", err)
	}

	migrator, err := NewMigrator(from, to)
	if err != nil {
		t.Fatalf("error creating migrator: %v", err)
	}
	migrator = migrator.WithClobber()

	err = migrator.Migrate()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !to.With("store1").Has([]byte("key1")) {
		t.Error("expected key1 to be  to destination keeper")
	}
}
