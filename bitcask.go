package database

import (
	"errors"
	"strings"
	"sync"

	"git.tcp.direct/Mirrors/bitcask-mirror"
)

// DB is an implementation of Filer using bitcask.
type DB struct {
	store map[string]*bitcask.Bitcask
	path  string
	mu    *sync.RWMutex
}

func NewDB(path string) *DB {
	return &DB{
		store: make(map[string]*bitcask.Bitcask),
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

	db.store[bucketName] = c

	return nil
}

// With calls the given underlying bitcask instance.
func (db *DB) With(bucketName string) Filer {
	db.mu.RLock()
	defer db.mu.RUnlock()
	d, ok := db.store[bucketName]
	if !ok {
		return nil
	}
	return d
}

// Close is a simple shim for bitcask's Close function.
func (db *DB) Close(bucketName string) error {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.store[bucketName].Close()
}

// Sync is a simple shim for bitcask's Sync function.
func (db *DB) Sync(bucketName string) error {
	db.mu.RLock()
	defer db.mu.RUnlock()
	return db.store[bucketName].Sync()
}

type withAllAction uint8

const (
	dclose withAllAction = iota
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
		if store == nil {
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

func (db *DB) CloseAll() error {
	return db.WithAll(dclose)
}

func (db *DB) SyncAll() error {
	return db.WithAll(dsync)
}
