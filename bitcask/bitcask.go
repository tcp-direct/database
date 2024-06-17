package bitcask

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"git.tcp.direct/Mirrors/bitcask-mirror"

	"git.tcp.direct/tcp.direct/database"
	"git.tcp.direct/tcp.direct/database/metadata"
	"git.tcp.direct/tcp.direct/database/models"
)

// Store is an implmentation of a Filer and a Searcher using Bitcask.
type Store struct {
	*bitcask.Bitcask
	database.Searcher
	closed bool
}

// Backend returns the underlying bitcask instance.
func (s Store) Backend() any {
	return s.Bitcask
}

// DB is a mapper of a Filer and Searcher implementation using Bitcask.
type DB struct {
	store       map[string]Store
	path        string
	mu          *sync.RWMutex
	meta        *metadata.Metadata
	initialized bool
}

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
	db.mu.RLock()
	if db.initialized {
		db.mu.RUnlock()
		return nil
	}
	db.mu.RUnlock()
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
		db.initialized = true
		return nil
	}

	if errors.Is(err, os.ErrNotExist) {
		db.meta, err = metadata.NewMetaFile(db.Type(), filepath.Join(db.path, "meta.json"))
		if err != nil {
			return fmt.Errorf("error creating meta file: %w", err)
		}
		db.initialized = true
		return nil
	}

	return err
}

// OpenDB will either open an existing set of bitcask datastores at the given directory, or it will create a new one.
func OpenDB(path string) *DB {
	return &DB{
		store:       make(map[string]Store),
		path:        path,
		mu:          &sync.RWMutex{},
		meta:        nil,
		initialized: false,
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
		db.store = make(map[string]Store)
	}
	os.Stat(filepath.Join(db.path, "meta.json"))

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
				newMeta := filepath.Join(db.path, name, "meta.json.backup")
				println("WARN: renaming", oldMeta, "to", newMeta)
				// likely defunct lockfile is present too, remove it
				if osErr := os.Rename(oldMeta, newMeta); osErr != nil {
					println("WARN: failed to rename", oldMeta, "to", newMeta, ":", osErr)
					return
				}
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
		db.store[name] = Store{Bitcask: c}
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
		if _, ok := opt.(bitcask.Option); !ok {
			return errors.New("invalid bitcask option type")
		}
		bitcaskopts = append(bitcaskopts, opt.(bitcask.Option))
	}
	db.mu.Lock()
	defer db.mu.Unlock()

	if len(defaultBitcaskOptions) > 0 {
		bitcaskopts = append(bitcaskopts, defaultBitcaskOptions...)
	}

	if _, ok := db.store[storeName]; ok {
		return ErrStoreExists
	}
	path := db.Path()
	if !strings.HasSuffix(db.Path(), "/") {
		path = db.Path() + "/"
	}
	c, e := bitcask.Open(path+storeName, bitcaskopts...)
	if e != nil {
		return e
	}
	db.store[storeName] = Store{Bitcask: c}
	return nil
}

// With calls the given underlying bitcask instance.
func (db *DB) With(storeName string) database.Store {
	if err := db.init(); err != nil {
		panic(err)
	}
	db.mu.RLock()
	defer db.mu.RUnlock()
	d, ok := db.store[storeName]
	if ok {
		return d
	}
	return nil
}

// WithNew calls the given underlying bitcask instance, if it doesn't exist, it creates it.
func (db *DB) WithNew(storeName string, opts ...any) database.Filer {
	if err := db.init(); err != nil {
		panic(err)
	}
	db.mu.RLock()
	defer db.mu.RUnlock()
	for _, opt := range opts {
		if _, ok := opt.(bitcask.Option); !ok {
			fmt.Println("invalid bitcask option type: ", opt)
			continue
		}
		defaultBitcaskOptions = append(defaultBitcaskOptions, opt.(bitcask.Option))
	}
	d, ok := db.store[storeName]
	if ok {
		return d
	}
	db.mu.RUnlock()
	err := db.Init(storeName)
	db.mu.RLock()
	if err != nil {
		fmt.Println("error creating bitcask store: ", err)

	}
	return db.store[storeName]
}

// Close is a simple shim for bitcask's Close function.
func (db *DB) Close(storeName string) error {
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
	if db == nil || db.store == nil || len(db.store) < 1 {
		return ErrNoStores
	}
	var errs = make([]error, len(db.store))
	for name, store := range db.store {
		var err error
		if store.Bitcask == nil {
			errs = append(errs, namedErr(name, ErrBogusStore))
			continue
		}
		switch action {
		case dclose:
			closeErr := store.Close()
			if errors.Is(closeErr, fs.ErrClosed) {
				continue
			}
			err = namedErr(name, closeErr)
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

func (db *DB) Type() string {
	return "bitcask"
}
