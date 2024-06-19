package bitcask

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"git.tcp.direct/Mirrors/bitcask-mirror"

	"git.tcp.direct/tcp.direct/database"
	"git.tcp.direct/tcp.direct/database/kv"
	"git.tcp.direct/tcp.direct/database/metadata"
	"git.tcp.direct/tcp.direct/database/models"
)

// Store is an implmentation of a Filer and a Searcher using Bitcask.
type Store struct {
	*bitcask.Bitcask
	database.Searcher
	closed *atomic.Bool
}

// Get is a wrapper around the bitcask Get function for error regularization.
func (s *Store) Get(key []byte) ([]byte, error) {
	if s.closed.Load() {
		return nil, fs.ErrClosed
	}
	ret, err := s.Bitcask.Get(key)
	err = kv.RegularizeKVError(key, ret, err)
	return ret, err
}

// Close is a wrapper around the bitcask Close function.
func (s *Store) Close() error {
	if s.closed.Load() {
		return fs.ErrClosed
	}
	s.closed.Store(true)
	return s.Bitcask.Close()
}

// Backend returns the underlying bitcask instance.
func (s *Store) Backend() any {
	return s.Bitcask
}

// DB is a mapper of a Filer and Searcher implementation using Bitcask.
type DB struct {
	store       map[string]*Store
	path        string
	mu          *sync.RWMutex
	meta        *metadata.Metadata
	initialized *atomic.Bool
}

// Meta returns the [models.Metadata] implementation of the bitcask keeper.
func (db *DB) Meta() models.Metadata {
	db.mu.RLock()
	m := db.meta
	db.mu.RUnlock()
	return m
}

// AllStores returns a map of the names of all bitcask datastores and the corresponding Filers.
func (db *DB) AllStores() map[string]database.Filer {
	if err := db.init(); err != nil {
		panic(err)
	}
	db.mu.RLock()
	defer db.mu.RUnlock()
	var stores = make(map[string]database.Filer)
	for n, s := range db.store {
		stores[n] = s
	}
	return stores
}

func (db *DB) init() error {
	var err error
	if db.initialized.Load() {
		return nil
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, err = os.Stat(db.path); os.IsNotExist(err) {
		err = os.MkdirAll(db.path, 0700)
		if err != nil {
			return fmt.Errorf("error creating bitcask directory: %w", err)
		}
	}
	stat, err := os.Stat(filepath.Join(db.path, "meta.json"))
	if err == nil && stat.IsDir() {
		return errors.New("meta.json is a directory")

	}
	if err == nil && !stat.IsDir() {
		if db.meta, err = metadata.OpenMetaFile(filepath.Join(db.path, "meta.json")); err != nil {
			return fmt.Errorf("error opening meta file: %w", err)
		}
		if db.meta.Type() != db.Type() {
			return fmt.Errorf("meta.json is not a bitcask meta file")
		}
		db.initialized.Store(true)
		return nil
	}

	if errors.Is(err, os.ErrNotExist) {
		db.meta, err = metadata.NewMetaFile(db.Type(), filepath.Join(db.path, "meta.json"))
		if err != nil {
			return fmt.Errorf("error creating meta file: %w", err)
		}
		db.initialized.Store(true)
		return nil
	}

	return err
}

// OpenDB will either open an existing set of bitcask datastores at the given directory, or it will create a new one.
func OpenDB(path string) *DB {
	ainit := &atomic.Bool{}
	ainit.Store(false)
	return &DB{
		store:       make(map[string]*Store),
		path:        path,
		mu:          &sync.RWMutex{},
		meta:        nil,
		initialized: ainit,
	}
}

// Discover will discover and initialize all existing bitcask stores at the path opened by [OpenDB].
func (db *DB) Discover() ([]string, error) {
	if err := db.init(); err != nil {
		return nil, err
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	stores := make([]string, 0)
	errs := make([]error, 0)
	if db.store == nil {
		db.store = make(map[string]*Store)
	}

	entries, err := fs.ReadDir(os.DirFS(db.path), ".")
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if _, ok := db.store[name]; ok {
			continue
		}
		recoverOnce := &sync.Once{}
	openUp:
		c, e := bitcask.Open(filepath.Join(db.path, name), defaultBitcaskOptions...)
		if e != nil {
			retry := false
			recoverOnce.Do(func() {
				metaErr := new(bitcask.ErrBadMetadata)
				if !errors.As(e, &metaErr) {
					return
				}
				if !strings.Contains(metaErr.Error(), "unexpected end of JSON input") {
					return
				}
				if c != nil {
					_ = c.Close()
				}
				println("WARN: bitcask store", name, "has bad metadata, attempting to repair")
				oldMeta := filepath.Join(db.path, name, "meta.json")
				oldMetaBackup := filepath.Join(db.path, name, "meta.json.backup")
				println("WARN: renaming", oldMeta, "to", oldMetaBackup)
				if osErr := os.Rename(oldMeta, oldMetaBackup); osErr != nil {
					println("Fatal: failed to rename", oldMeta, "to", oldMetaBackup, ":", osErr)
					panic(osErr)
				}

				// likely defunct lockfile is present too, remove it
				if _, serr := os.Stat(filepath.Join(db.path, name, "lock")); serr == nil {
					println("WARN: removing defunct lockfile")
					_ = os.Remove(filepath.Join(db.path, name, "lock"))
				}

				retry = true
			})
			if retry {
				goto openUp
			}
			errs = append(errs, e)
			continue
		}
		aclosed := &atomic.Bool{}
		aclosed.Store(false)
		db.store[name] = &Store{Bitcask: c, closed: aclosed}
		stores = append(stores, name)
	}

	for _, e := range errs {
		if err == nil {
			err = e
			continue
		}
		err = fmt.Errorf("%w: %v", err, e)
	}

	return stores, err
}

// Path returns the base path where we store our bitcask "stores".
func (db *DB) Path() string {
	return db.path
}

var defaultBitcaskOptions []bitcask.Option

// SetDefaultBitcaskOptions options will set the options used for all subsequent bitcask stores that are initialized.
func SetDefaultBitcaskOptions(bitcaskopts ...bitcask.Option) {
	defaultBitcaskOptions = append(defaultBitcaskOptions, bitcaskopts...)
}

// WithMaxDatafileSize is a shim for bitcask's WithMaxDataFileSize function.
func WithMaxDatafileSize(size int) bitcask.Option {
	return bitcask.WithMaxDatafileSize(size)
}

// WithMaxKeySize is a shim for bitcask's WithMaxKeySize function.
func WithMaxKeySize(size uint32) bitcask.Option {
	return bitcask.WithMaxKeySize(size)
}

// WithMaxValueSize is a shim for bitcask's WithMaxValueSize function.
func WithMaxValueSize(size uint64) bitcask.Option {
	return bitcask.WithMaxValueSize(size)
}

// Init opens a bitcask store at the given path to be referenced by storeName.
func (db *DB) Init(storeName string, opts ...any) error {
	if err := db.init(); err != nil {
		return err
	}
	var bitcaskopts []bitcask.Option
	for _, opt := range opts {
		_, isOptOK := opt.(bitcask.Option)
		_, isOptsOK := opt.([]bitcask.Option)
		if !isOptOK && !isOptsOK {
			return fmt.Errorf("invalid bitcask option type (%T): %v", opt, opt)
		}
		if isOptOK {
			bitcaskopts = append(bitcaskopts, opt.(bitcask.Option))
		}
		if isOptsOK {
			bitcaskopts = append(bitcaskopts, opt.([]bitcask.Option)...)
		}
	}

	if len(defaultBitcaskOptions) > 0 {
		bitcaskopts = append(defaultBitcaskOptions, bitcaskopts...)
	}

	db.mu.Lock()
	err := db.initStore(storeName, bitcaskopts...)
	db.mu.Unlock()

	return err
}

// initStore is a helper function to initialize a bitcask store, caller must hold keeper's lock.
func (db *DB) initStore(storeName string, opts ...bitcask.Option) error {
	if _, ok := db.store[storeName]; ok {
		return ErrStoreExists
	}

	c, e := bitcask.Open(filepath.Join(db.Path(), storeName), opts...)
	if e != nil {
		return e
	}

	aclosed := &atomic.Bool{}
	aclosed.Store(false)
	db.store[storeName] = &Store{Bitcask: c, closed: aclosed}
	return nil
}

// Destroy will remove the bitcask store and all data associated with it.
func (db *DB) Destroy(storeName string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	st, ok := db.store[storeName]
	if !ok {
		return ErrBogusStore
	}
	err := st.Close()
	if err != nil {
		return err
	}
	delete(db.store, storeName)
	return os.RemoveAll(filepath.Join(db.path, storeName))
}

// With calls the given underlying bitcask instance.
func (db *DB) With(storeName string) database.Store {
	if err := db.init(); err != nil {
		panic(err)
	}
	db.mu.RLock()
	d, ok := db.store[storeName]
	if ok && !d.closed.Load() {
		db.mu.RUnlock()
		return d
	}
	if ok && d.closed.Load() {
		db.mu.RUnlock()
		db.mu.Lock()
		delete(db.store, storeName)
		db.mu.Unlock()
		return nil
	}
	db.mu.RUnlock()
	return nil
}

// WithNew calls the given underlying bitcask instance, if it doesn't exist, it creates it.
func (db *DB) WithNew(storeName string, opts ...any) database.Filer {
	if err := db.init(); err != nil {
		panic(err)
	}

	newOpts := make([]bitcask.Option, 0)
	for _, opt := range opts {
		if _, ok := opt.(bitcask.Option); !ok {
			fmt.Println("invalid bitcask option type: ", opt)
			continue
		}
		newOpts = append(newOpts, opt.(bitcask.Option))
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	d, ok := db.store[storeName]
	if ok {
		if d.Bitcask == nil || d.closed == nil || d.closed.Load() {
			delete(db.store, storeName)
			if err := db.initStore(storeName, newOpts...); err != nil {
				fmt.Println("error re-initializing bitcask store: ", err.Error())
			}
			return db.store[storeName]
		}
		return d
	}

	err := db.initStore(storeName, newOpts...)
	if err != nil {
		fmt.Println("error creating bitcask store: ", err)
	}
	return db.store[storeName]
}

// Close is a simple shim for bitcask's Close function.
func (db *DB) Close(storeName string) error {
	db.mu.RLock()
	st, ok := db.store[storeName]
	if !ok {
		db.mu.RUnlock()
		return ErrBogusStore
	}
	db.mu.RUnlock()
	err := st.Close()
	if err != nil {
		return err
	}
	db.mu.Lock()
	delete(db.store, storeName)
	db.mu.Unlock()
	return nil
}

// Sync is a simple shim for bitcask's Sync function.
func (db *DB) Sync(storeName string) error {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if _, ok := db.store[storeName]; !ok {
		return ErrBogusStore
	}
	return db.store[storeName].Sync()
}

// withAllAction
type withAllAction uint8

const (
	// dclose
	dclose withAllAction = iota
	// dsync
	dsync
)

// withAll performs an action on all bitcask stores that we have open.
// In the case of an error, withAll will continue and return a compound form of any errors that occurred.
// For now this is just for Close and Sync, thusly it does a hard lock on the Keeper.
func (db *DB) withAll(action withAllAction) error {
	if db == nil || db.store == nil {
		panic("bitcask: nil db or db.store")
	}
	if len(db.store) < 1 {
		return ErrNoStores
	}
	var errs = make([]error, len(db.store))
	for name, store := range db.store {
		var err error
		if store.Bitcask == nil {
			errs = append(errs, namedErr(name, ErrBogusStore))
			continue
		}
		if store.closed.Load() {
			continue
		}
		switch action {
		case dclose:
			db.mu.Lock()
			closeErr := store.Close()
			if errors.Is(closeErr, fs.ErrClosed) {
				continue
			}
			err = namedErr(name, closeErr)
			delete(db.store, name)
			db.mu.Unlock()
		case dsync:
			err = namedErr(name, store.Sync())
		default:
			return ErrUnknownAction
		}
		if err == nil {
			continue
		}
		errs = append(errs, err)
	}
	return compoundErrors(errs)
}

// SyncAndCloseAll implements the method from Keeper to sync and close all bitcask stores.
func (db *DB) SyncAndCloseAll() error {
	if db == nil || db.store == nil || len(db.store) < 1 {
		return ErrNoStores
	}
	var errs = make([]error, len(db.store))
	errSync := namedErr("sync", db.SyncAll())
	if errSync != nil {
		errs = append(errs, errSync)
	}
	errClose := namedErr("close", db.CloseAll())
	if errClose != nil {
		errs = append(errs, errClose)
	}
	return compoundErrors(errs)
}

// CloseAll closes all bitcask datastores.
func (db *DB) CloseAll() error {
	return db.withAll(dclose)
}

// SyncAll syncs all bitcask datastores.
func (db *DB) SyncAll() error {
	return db.withAll(dsync)
}

// Type returns the type of keeper, in this case "bitcask".
// This is in order to implement [database.Keeper].
func (db *DB) Type() string {
	return "bitcask"
}
