package database_test

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"

	_ "git.tcp.direct/tcp.direct/database/bitcask" // register bitcask
	"git.tcp.direct/tcp.direct/database/kv"
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
		t.Run(name, func(t *testing.T) {
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
