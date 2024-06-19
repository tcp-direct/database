package database_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"slices"
	"testing"

	"git.tcp.direct/kayos/common/entropy"

	"git.tcp.direct/tcp.direct/database"
	"git.tcp.direct/tcp.direct/database/backup"
	_ "git.tcp.direct/tcp.direct/database/bitcask" // register bitcask
	"git.tcp.direct/tcp.direct/database/kv"
	"git.tcp.direct/tcp.direct/database/models"
	_ "git.tcp.direct/tcp.direct/database/pogreb" // register pogreb
	"git.tcp.direct/tcp.direct/database/registry"
)

func TestAllKeepers(t *testing.T) {
	keepers := registry.AllKeepers()
	if len(keepers) != 2 {
		t.Errorf("expected 2 keepers, got %d", len(keepers))
	}
	if !slices.Contains(keepers, "bitcask") {
		t.Error("expected 'bitcask' keeper")
	}
	if !slices.Contains(keepers, "pogreb") {
		t.Error("expected 'pogreb' keeper")
	}
	t.Logf("keepers: %v", keepers)
}

func TestImplementationsBasic(t *testing.T) {
	testKey := []byte("yeeterson")
	testValue := []byte("mcgeeterson")
	for _, name := range registry.AllKeepers() {
		t.Run(name+"_basic", func(t *testing.T) {
			tpath := filepath.Join(t.TempDir(), name)
			keeper := registry.GetKeeper(name)
			if keeper == nil {
				t.Fatalf("expected keeper for %q, got nil", name)
			}
			instance, err := keeper(tpath)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if instance == nil {
				t.Fatalf("expected keeper instance, got nil")
			}
			storeName := name + "_test"

			// Test that we can't access a store that hasn't been initialized,
			// but that was initialized before when using a different keeper.
			//
			// additionally we'll test the implementation's destroy method here,
			// allowing us to use the store name for the rest of the parent test.
			t.Run("sanity_check", func(t *testing.T) {
				if shouldntExist := instance.With(storeName); shouldntExist != nil {
					t.Fatalf("got a store that shouldn't exist: %v", shouldntExist)
				}
				if err = instance.Init(storeName); err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				defer func() {
					t.Run("destroy", func(t *testing.T) {
						if err = instance.Destroy(storeName); err != nil {
							t.Fatalf("expected no error, got %v", err)
						}
					})
				}()
				if nonexistentValue, shouldError := instance.With(storeName).Get(testKey); shouldError == nil || nonexistentValue != nil {
					if shouldError == nil {
						t.Fatalf("expected error, got nil and got %v", nonexistentValue)
					}
					if !kv.IsNonExistentKey(shouldError) {
						t.Fatalf("expected NonExistentKeyError, got %v", shouldError)
					}
					if nonexistentValue != nil {
						t.Fatalf("expected nil value, got %v", nonexistentValue)
					}
				}
			})

			if err = instance.Init(storeName); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if err = instance.With(storeName).Put(testKey, testValue); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			var ret []byte
			if ret, err = instance.With(storeName).Get(testKey); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if string(ret) != string(testValue) {
				t.Errorf("expected %q, got %q", testValue, ret)
			}
			if err = instance.With(storeName).Delete(testKey); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if ret, err = instance.With(storeName).Get(testKey); err == nil {
				t.Errorf("expected error when fetching key we just deleted, got nil")
			}
			neErr := &kv.NonExistentKeyError{}
			if !errors.As(err, &neErr) {
				t.Errorf("expected NonExistentKeyError, got %v", err)
			} else {
				if !kv.IsNonExistentKey(err) {
					t.Errorf("expected IsNonExistentKey to return true, got false")
				}
			}
			t.Run("close_and_reopen", func(t *testing.T) {
				if err = instance.With(storeName).Put(testKey, testValue); err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if ret, err = instance.With(storeName).Get(testKey); err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if string(ret) != string(testValue) {
					t.Errorf("expected %q, got %q", testValue, ret)
				}
				if err = instance.Close(storeName); err != nil {
					t.Fatalf("expected no error during close, got %v", err)
				}
				if nonExistentStore := instance.With(storeName); nonExistentStore != nil {
					t.Fatalf("expected nil store after close, got %v", nonExistentStore)
				}
				if ret, err = instance.WithNew(storeName).Get(testKey); err != nil {
					t.Fatalf("expected no error getting a key we wrote before closing, got %v", err)
				}
				if string(ret) != string(testValue) {
					t.Errorf("expected %q, got %q", testValue, ret)
				}
			})
			if err = instance.SyncAndCloseAll(); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if nonExistentStore := instance.With(storeName); nonExistentStore != nil {
				t.Fatalf("expected nil store after close, got %v", nonExistentStore)
			}
		})
	}
}

func insertGarbo(t *testing.T, db database.Keeper) map[string][]kv.KeyValue {
	t.Helper()
	inserted := make(map[string][]kv.KeyValue)
	for i := 0; i < 10; i++ {
		newName := entropy.RandStrWithUpper(5)
		if err := db.Init(newName); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		inserted[newName] = make([]kv.KeyValue, 0, 100)
		for j := 0; j < 100; j++ {
			key := []byte(entropy.RandStrWithUpper(10))
			value := []byte(entropy.RandStrWithUpper(10))
			if err := db.With(newName).Put(key, value); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			inserted[newName] = append(inserted[newName], kv.NewKeyValueFromBytes(key, value))
		}
		if db.With(newName).Len() != 100 {
			t.Fatalf("expected 100 keys in store, got %d", db.With(newName).Len())
		}
	}
	return inserted
}

func TestImplementationsBackup(t *testing.T) {
	for _, name := range registry.AllKeepers() {
		t.Run(name+"_backup", func(t *testing.T) {
			// t.Parallel()
			tpath := filepath.Join(t.TempDir(), name)
			keeper := registry.GetKeeper(name)
			if keeper == nil {
				t.Fatalf("expected keeper for %q, got nil", name)
			}
			instance, err := keeper(tpath)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if instance == nil {
				t.Fatalf("expected keeper instance, got nil")
			}
			garbo := insertGarbo(t, instance)
			if err = instance.SyncAll(); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			var bu models.Backup

			newBackup := filepath.Join(t.TempDir(), "backup.tar.gz")

			t.Logf("creating backup: %v", newBackup)

			if bu, err = instance.BackupAll(newBackup); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if bu == nil {
				t.Fatalf("expected backup, got nil")
			}

			t.Logf("backup creation result: (%T) %v", bu, bu)

			t.Logf("verifying backup: %v", bu.Path())

			if vErr := backup.VerifyBackup(bu.(backup.BackupMetadata)); vErr != nil {
				t.Fatalf("expected no error, got %v", vErr)
			}

			t.Logf("restoring backup: %v", bu.Path())
			if err = instance.RestoreAll(bu.Path()); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			t.Logf("backup restored: %v", bu.Path())
			t.Run("verify_restored_data", func(t *testing.T) {
				for storeName, kvs := range garbo {
					t.Run("verify_"+storeName, func(t *testing.T) {
						for _, kvTuple := range kvs {
							// t.Logf("checking key: %s", kvTuple.Key.String())
							var ret []byte
							var getErr error
							if ret, getErr = instance.With(storeName).Get(kvTuple.Key.Bytes()); getErr != nil {
								t.Fatalf("expected no error, got %v", getErr)
							}
							if !bytes.Equal(kvTuple.Value.Bytes(), ret) {
								t.Errorf("expected %q, got %q", kvTuple.Value.String(), ret)
							}
						}
					})
				}
			})
			if err = instance.SyncAndCloseAll(); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
