package bitcask

import (
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

// Path returns the base path where we store our bitcask "stores".
func (db *DB) Path() string {
	return db.path
}

// Init opens a bitcask store at the given path to be referenced by storeName.
func (db *DB) Init(storeName string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if _, ok := db.store[storeName]; ok {
		return errStoreExists
	}
	path := db.Path()
	if !strings.HasSuffix("/", db.Path()) {
		path = db.Path() + "/"
	}
	c, e := bitcask.Open(path + storeName)
	if e != nil {
		return e
	}

	db.store[storeName] = Store{Bitcask: c}

	return nil
}

// With calls the given underlying bitcask instance.
func (db *DB) With(storeName string) Store {
	db.mu.RLock()
	defer db.mu.RUnlock()
	d, ok := db.store[storeName]
	if !ok {
		return Store{Bitcask: nil}
	}
	return d
}

// Close is a simple shim for bitcask's Close function.
func (db *DB) Close(storeName string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	st, ok := db.store[storeName]
	if !ok {
		return errBogusStore
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
	var errs []error
	if len(db.store) < 1 {
		return errNoStores
	}
	for name, store := range db.store {
		var err error
		if store.Bitcask == nil {
			errs = append(errs, namedErr(name, errBogusStore))
			continue
		}
		switch action {
		case dclose:
			err = namedErr(name, db.Close(name))
		case dsync:
			err = namedErr(name, db.Sync(name))
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
	return db.withAll(dclose)
}

// SyncAll syncs all bitcask datastores.
func (db *DB) SyncAll() error {
	return db.withAll(dsync)
}
