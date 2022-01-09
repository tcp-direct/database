package bitcask

import (
	"errors"
	"strings"
	"sync"

	"git.tcp.direct/Mirrors/bitcask-mirror"

	"git.tcp.direct/kayos/database"
)

// Store is an implmentation of a Filer and a Searcher using Bitcask.
type Store struct {
	*bitcask.Bitcask
	database.Searcher
}

// DB is a mapper of a Filer and Searcher implementation using Bitcask.
type DB struct {
	store map[string]Store
	path  string
	mu    *sync.RWMutex
}

// OpenDB will either open an existing set of bitcask datastores at the given directory, or it will create a new one.
func OpenDB(path string) *DB {
	return &DB{
		store: make(map[string]Store),
		path:  path,
		mu:    &sync.RWMutex{},
	}
}

// Path returns the base path where we store our bitcask "buckets".
func (db *DB) Path() string {
	return db.path
}

// Init opens a bitcask store at the given path to be referenced by bucketName.
func (db *DB) Init(bucketName string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.store["bucketName"]; ok {
		return errors.New("bucket already exists")
	}
	path := db.Path()
	if !strings.HasSuffix("/", db.Path()) {
		path = db.Path() + "/"
	}
	c, e := bitcask.Open(path + bucketName)
	if e != nil {
		return e
	}

	db.store[bucketName] = Store{Bitcask: c}

	return nil
}

// With calls the given underlying bitcask instance.
func (db *DB) With(bucketName string) Store {
	db.mu.RLock()
	defer db.mu.RUnlock()
	d, ok := db.store[bucketName]
	if !ok {
		return Store{Bitcask: nil}
	}
	return d
}

// Close is a simple shim for bitcask's Close function.
func (db *DB) Close(bucketName string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	err := db.store[bucketName].Close()
	if err != nil {
		return err
	}
	delete(db.store, bucketName)
	return nil
}

// Sync is a simple shim for bitcask's Sync function.
func (db *DB) Sync(bucketName string) error {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.store[bucketName].Sync()
}

// withAllAction
type withAllAction uint8

const (
	// dclose
	dclose withAllAction = iota
	// dsync
	dsync
)

// WithAll performs an action on all bitcask stores that we have open.
// In the case of an error, WithAll will continue and return a compound form of any errors that occurred.
// For now this is just for Close and Sync, thusly it does a hard lock on the Keeper.
func (db *DB) WithAll(action withAllAction) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	var errs []error
	for name, store := range db.store {
		var err error
		if store.Bitcask == nil {
			errs = append(errs, namedErr(name, errBogusStore))
			continue
		}
		switch action {
		case dclose:
			err = namedErr(name, store.Close())
		case dsync:
			err = namedErr(name, store.Sync())
		default:
			return errUnknownAction
		}
		if err == nil {
			continue
		}
		errs = append(errs, err)
	}
	return compoundErrors(errs)
}

// SyncAndCloseAll implements the method from Keeper.
func (db *DB) SyncAndCloseAll() error {
	var errs []error
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
	return db.WithAll(dclose)
}

// SyncAll syncs all bitcask datastores.
func (db *DB) SyncAll() error {
	return db.WithAll(dsync)
}
