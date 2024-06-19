package bitcask

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"git.tcp.direct/tcp.direct/database/backup"
	"git.tcp.direct/tcp.direct/database/models"
)

func (db *DB) BackupAll(archivePath string) (models.Backup, error) {
	// calling write lock should stop any other operations on the stores while we backup. shouldn't need to close.
	db.mu.Lock()
	defer db.mu.Unlock()

	var (
		storeNames []string
		discovErr  error
	)

	if storeNames, discovErr = db.discover(); discovErr != nil {
		return nil, discovErr
	}

	if err := db.SyncAll(); err != nil {
		return nil, err
	}

	if err := db.closeAll(); err != nil {
		return nil, err
	}

	for name, store := range db.store {
		if !slices.Contains(storeNames, name) {
			println("WARN: store", name, "not found in discovered stores... appending but this is unexpected behavior.")
			if err := store.Sync(); err != nil {
				return nil, err
			}
			if err := store.Close(); err != nil {
				return nil, err
			}
			storeNames = append(storeNames, name)
		}
	}

	bu, err := backup.NewTarGzBackup(db.path, archivePath, storeNames)
	if err != nil {
		return nil, err
	}
	db.meta.Backups[bu.FilePath] = bu
	err = db.meta.Sync()
	return bu, err
}

func (db *DB) RestoreAll(archivePath string) error {
	var preBu models.Backup

	if err := db.SyncAndCloseAll(); err != nil && !errors.Is(err, ErrNoStores) {
		return err
	}

	if len(db.store) > 0 {
		var err error
		if preBu, err = db.BackupAll(filepath.Join(os.TempDir(), "pre-restore-"+time.Now().Format(time.RFC3339)+".tar.gz")); err != nil {
			return fmt.Errorf("failed to create pre-restore backup: %w", err)
		}
		for name := range db.store {
			if err = db.Destroy(name); err != nil {
				return fmt.Errorf("failed to destroy existing store %s (backup: %s): %w", preBu.Path(), name, err)
			}
		}
	}

	preBackupPath := ""

	if preBu != nil {
		preBackupPath = fmt.Sprintf(" (backup: %s)", preBu.Path())
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	db.initialized.Store(false)

	if err := backup.RestoreTarGzBackup(archivePath, db.path); err != nil {
		return err
	}

	if err := db._init(); err != nil {
		return fmt.Errorf("failed to re-init db after restore%s: %w", preBackupPath, err)
	}

	_, err := db.discover()
	if err != nil {
		return fmt.Errorf("failed during discover call after restore%s: %w", preBackupPath, err)
	}

	if err = db.meta.Sync(); err != nil {
		return fmt.Errorf("failed to sync meta after restore%s: %w", preBackupPath, err)
	}

	return nil
}
