package bitcask

import (
	"errors"
	"strings"
	"sync"

	"git.tcp.direct/Mirrors/bitcask-mirror"

	"git.tcp.direct/tcp.direct/database"
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
	store map[string]Store
	path  string
	mu    *sync.RWMutex
}

// AllStores returns a map of the names of all bitcask datastores and the corresponding Filers.
func (db *DB) AllStores() map[string]database.Filer {
	db.mu.RLock()
	defer db.mu.RUnlock()
	var stores = make(map[string]database.Filer)
	for n, s := range db.store {
		stores[n] = s
	}
	return stores
}

// OpenDB will either open an existing set of bitcask datastores at the given directory, or it will create a new one.
func OpenDB(path string) *DB {
	return &DB{
		store: make(map[string]Store),
		path:  path,
		mu:    &sync.RWMutex{},
	}
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
	db.mu.RLock()
	defer db.mu.RUnlock()
	d, ok := db.store[storeName]
	if ok {
		return d
	}
	return nil
}

// WithNew calls the given underlying bitcask instance, if it doesn't exist, it creates it.
func (db *DB) WithNew(storeName string) database.Filer {
	db.mu.RLock()
	defer db.mu.RUnlock()
	d, ok := db.store[storeName]
	if ok {
		return d
	}
	db.mu.RUnlock()
	err := db.Init(storeName)
	db.mu.RLock()
	if err == nil {
		return db.store[storeName]
	}
	return Store{Bitcask: nil}
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
			err = namedErr(name, store.Close())
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
