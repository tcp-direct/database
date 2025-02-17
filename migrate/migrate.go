// Package migrate implements the migration of data from one type of Keeper to another.
package migrate

import (
	"context"
	"errors"
	"sync"

	"github.com/tcp-direct/database"
)

var (
	ErrNoStores = errors.New("no stores found in source keeper")
	ErrDupKeys  = errors.New(
		"duplicate keys found in destination stores, enable skipping or clobbering of existing data to continue migration",
	)
)

type ErrDuplicateKeys struct {
	// map[store][]keys
	Duplicates map[string][][]byte
}

func (e ErrDuplicateKeys) Unwrap() error {
	return ErrDupKeys
}

func (e ErrDuplicateKeys) Error() string {
	return "duplicate keys found in destination stores, enable skipping or clobbering of existing data to continue migration"
}

func NewDuplicateKeysErr(duplicates map[string][][]byte) *ErrDuplicateKeys {
	return &ErrDuplicateKeys{Duplicates: duplicates}
}

type Migrator struct {
	From database.Keeper
	To   database.Keeper

	duplicateKeys map[string]map[string]struct{}

	clobber      bool
	skipExisting bool

	mu sync.Mutex
}

func mapMaptoMapSlice(m map[string]map[string]struct{}) map[string][][]byte {
	out := make(map[string][][]byte)
	for store, keys := range m {
		for key := range keys {
			out[store] = append(out[store], []byte(key))
		}
	}
	return out
}

func NewMigrator(from, to database.Keeper) (*Migrator, error) {
	if _, err := from.Discover(); err != nil {
		return nil, err
	}
	if _, err := to.Discover(); err != nil {
		return nil, err
	}
	return &Migrator{
		From:         from,
		To:           to,
		clobber:      false,
		skipExisting: false,
	}, nil
}

// WithClobber sets the clobber flag on the Migrator, allowing it to overwrite existing data in the destination Keeper.
func (m *Migrator) WithClobber() *Migrator {
	m.mu.Lock()
	m.clobber = true
	m.mu.Unlock()
	return m
}

// WithSkipExisting sets the skipExisting flag on the Migrator, allowing it to skip existing data in the destination Keeper.
func (m *Migrator) WithSkipExisting() *Migrator {
	m.mu.Lock()
	m.skipExisting = true
	m.mu.Unlock()
	return m
}

func (m *Migrator) CheckDupes() error {
	fromStores := m.From.AllStores()
	toStores := m.To.AllStores()

	if len(fromStores) == 0 {
		return ErrNoStores
	}

	if m.duplicateKeys == nil {
		m.duplicateKeys = make(map[string]map[string]struct{})
	}

	wg := &sync.WaitGroup{}

	for storeName, store := range fromStores {
		existingStore, ok := toStores[storeName]
		if !ok {
			continue
		}
		if existingStore.Len() == 0 {
			continue
		}
		wg.Add(1)
		go func(storeName string, store, existingStore database.Filer) {
			defer wg.Done()
			keys := existingStore.Keys()
			for _, key := range keys {
				if store.Has(key) {
					m.mu.Lock()
					if _, exists := m.duplicateKeys[storeName]; !exists {
						m.duplicateKeys[storeName] = make(map[string]struct{})
					}
					m.duplicateKeys[storeName][string(key)] = struct{}{}
					m.mu.Unlock()
				}
			}
		}(storeName, store, existingStore)
	}

	wg.Wait()

	if len(m.duplicateKeys) == 0 || m.skipExisting || m.clobber {
		return nil
	}

	m.mu.Lock()
	mslice := mapMaptoMapSlice(m.duplicateKeys)
	m.mu.Unlock()

	return NewDuplicateKeysErr(mslice)
}

func (m *Migrator) Migrate() error {
	fromStores := m.From.AllStores()

	if len(fromStores) == 0 {
		return ErrNoStores
	}

	if err := m.CheckDupes(); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	errCh := make(chan error, len(fromStores))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg := &sync.WaitGroup{}
	for srcStoreName, srcStore := range fromStores {
		if srcStore.Len() == 0 {
			continue
		}
		wg.Add(1)
		go func(storeName string, store database.Filer) {
			defer wg.Done()
			keys := store.Keys()
			for _, key := range keys {
				select {
				case <-ctx.Done():
					return
				default:
				}
				srcVal, err := m.From.With(storeName).Get(key)
				if err != nil {
					errCh <- err
					return
				}
				if _, exists := m.duplicateKeys[storeName][string(key)]; exists {
					if m.skipExisting {
						continue
					}
					if !m.clobber {
						errCh <- NewDuplicateKeysErr(mapMaptoMapSlice(m.duplicateKeys))
						return
					}
					if err = m.To.With(storeName).Put(key, srcVal); err != nil {
						errCh <- err
						return
					}
					continue
				}
				if err = m.To.WithNew(storeName).Put(key, srcVal); err != nil {
					return
				}
			}
		}(srcStoreName, srcStore)
	}

	wgCh := make(chan struct{})

	go func() {
		wg.Wait()
		close(wgCh)
	}()

	select {
	case <-wgCh:
	case err := <-errCh:
		return err
	}

	fStores := m.From.AllStores()
	tStores := m.To.AllStores()

	if len(fStores) != len(tStores) {
		return errors.New("number of stores in source and destination keepers do not match")
	}

	syncErrs := make([]error, 0, 2)
	syncErrs = append(syncErrs, m.From.SyncAll(), m.To.SyncAll())
	return errors.Join(syncErrs...)
}
