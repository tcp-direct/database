package pogreb

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/akrylysov/pogreb"

	"git.tcp.direct/tcp.direct/database"
	"git.tcp.direct/tcp.direct/database/metadata"
	"git.tcp.direct/tcp.direct/database/models"
)

type Option func(*WrappedOptions)

var OptionAllowRecovery = func(opts *WrappedOptions) {
	opts.AllowRecovery = true
}

func AllowRecovery() Option {
	return OptionAllowRecovery
}

func SetPogrebOptions(options pogreb.Options) Option {
	return func(opts *WrappedOptions) {
		opts.Options = &options
	}
}

type WrappedOptions struct {
	*pogreb.Options
	// AllowRecovery allows the database to be recovered if a lockfile is detected upon running Init.
	AllowRecovery bool
}

func (pstore *Store) Len() int {
	return int(pstore.DB.Count())
}

func (pstore *Store) Keys() [][]byte {
	iter := pstore.DB.Items()
	ks := make([][]byte, pstore.DB.Count())
	for k, _, _ := iter.Next(); k != nil; k, _, _ = iter.Next() {
		ks = append(ks, k)
	}
	return ks
}

func (pstore *Store) Has(key []byte) bool {
	ok, err := pstore.DB.Has(key)
	if err != nil {
		_, _ = os.Stderr.WriteString("error checking pogreb store for key: " + err.Error())
	}
	return ok
}

// Store is an implmentation of a Filer and a Searcher using Bitcask.
type Store struct {
	*pogreb.DB
	database.Searcher
	opts   *WrappedOptions
	closed bool
}

// Backend returns the underlying pogreb instance.
func (pstore *Store) Backend() any {
	return pstore.DB
}

// DB is a mapper of a Filer and Searcher implementation using pogreb.
type DB struct {
	store map[string]*Store
	path  string
	mu    *sync.RWMutex
	meta  *metadata.Metadata

	initialized bool
}

func (db *DB) Meta() models.Metadata {
	return db.meta
}

// AllStores returns a map of the names of all pogreb datastores and the corresponding Filers.
func (db *DB) AllStores() map[string]database.Filer {
	db.mu.RLock()
	defer db.mu.RUnlock()
	var stores = make(map[string]database.Filer)
	for n, s := range db.store {
		stores[n] = s
	}
	return stores
}

// FIXME: not returning the error is probably pretty irresponsible.

// OpenDB will either open an existing set of pogreb datastores at the given directory, or it will create a new one.
func OpenDB(path string) *DB {
	db := &DB{
		store: make(map[string]*Store),
		path:  path,
		mu:    &sync.RWMutex{},
		meta:  nil,

		initialized: false,
	}
	return db
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
			return fmt.Errorf("meta.json is not a pogreb meta file")
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

// Path returns the base path where we store our pogreb "stores".
func (db *DB) Path() string {
	return db.path
}

var defaultPogrebOptions = &WrappedOptions{
	Options:       nil,
	AllowRecovery: false,
}

// SetDefaultPogrebOptions options will set the options used for all subsequent pogreb stores that are initialized.
func SetDefaultPogrebOptions(pogrebopts ...any) {
	inner, pgoptOk := pogrebopts[0].(pogreb.Options)
	wrapped, pgoptWrappedOk := pogrebopts[0].(*WrappedOptions)
	switch {
	case !pgoptOk && !pgoptWrappedOk:
		panic("invalid pogreb options")
	case pgoptOk:
		defaultPogrebOptions = &WrappedOptions{
			Options:       &inner,
			AllowRecovery: false,
		}
	case pgoptWrappedOk:
		defaultPogrebOptions = wrapped
	}
}

func normalizeOptions(opts ...any) *WrappedOptions {
	var pogrebopts *WrappedOptions
	pgInner, pgOK := opts[0].(pogreb.Options)
	pgWrapped, pgWrappedOK := opts[0].(WrappedOptions)
	switch {
	case !pgOK && !pgWrappedOK:
		return nil
	case pgOK:
		pogrebopts = &WrappedOptions{
			Options:       &pgInner,
			AllowRecovery: false,
		}
	case pgWrappedOK:
		pogrebopts = &pgWrapped
	}
	return pogrebopts
}

func (db *DB) initStore(storeName string, pogrebOpts *WrappedOptions) error {
	if _, ok := db.store[storeName]; ok {
		return ErrStoreExists
	}
	path := db.Path()
	if _, err := os.Stat(filepath.Join(path, storeName, "lock")); !os.IsNotExist(err) && !pogrebOpts.AllowRecovery {
		return fmt.Errorf("%w: and seems to be running... "+
			"Please close it first, or use InitWithRecovery", ErrStoreExists)
	}
	c, e := pogreb.Open(filepath.Join(path, storeName), pogrebOpts.Options)
	if e != nil {
		return e
	}
	db.store[storeName] = &Store{DB: c}
	return nil
}

// Init opens a pogreb store at the given path to be referenced by storeName.
func (db *DB) Init(storeName string, opts ...any) error {
	if err := db.init(); err != nil {
		return err
	}
	pogrebopts := defaultPogrebOptions
	if len(opts) > 0 {
		pogrebopts = normalizeOptions(opts...)
		if pogrebopts == nil {
			return ErrBadOptions
		}
	}
	db.mu.Lock()
	err := db.initStore(storeName, pogrebopts)
	db.mu.Unlock()

	return err
}

// With calls the given underlying pogreb instance.
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

// WithNew calls the given underlying pogreb instance, if it doesn't exist, it creates it.
func (db *DB) WithNew(storeName string, opts ...any) database.Filer {
	if err := db.init(); err != nil {
		panic(err)
	}
	pogrebopts := defaultPogrebOptions
	if len(opts) > 0 {
		if pogrebopts = normalizeOptions(opts...); pogrebopts == nil {
			return nil
		}
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	d, ok := db.store[storeName]
	if ok && d == nil {
		delete(db.store, storeName)
		ok = false
	}

	if ok && d != nil {
		return d
	}
	err := db.initStore(storeName, pogrebopts)
	if err == nil {
		return db.store[storeName]
	}
	_, _ = os.Stderr.WriteString("error creating pogreb store: " + err.Error())
	return &Store{DB: nil}
}

// Close is a simple shim for pogreb's Close function.
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

// Sync is a simple shim for pogreb's Sync function.
func (db *DB) Sync(storeName string) error {
	db.mu.RLock()
	defer db.mu.RUnlock()
	if _, ok := db.store[storeName]; !ok {
		return ErrBogusStore
	}
	return db.store[storeName].Backend().(*pogreb.DB).Sync()
}

// withAllAction
type withAllAction uint8

const (
	// dclose
	dclose withAllAction = iota
	// dsync
	dsync
)

// withAll performs an action on all pogreb stores that we have open.
// In the case of an error, withAll will continue and return a compound form of any errors that occurred.
// For now this is just for Close and Sync, thusly it does a hard lock on the Keeper.
func (db *DB) withAll(action withAllAction) error {
	if db == nil || db.store == nil || len(db.store) < 1 {
		return ErrNoStores
	}
	var errs = make([]error, len(db.store))
	for name, store := range db.store {
		var err error
		if store == nil || store.Backend().(*pogreb.DB) == nil {
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

// SyncAndCloseAll implements the method from Keeper to sync and close all pogreb stores.
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

// CloseAll closes all pogreb datastores.
func (db *DB) CloseAll() error {
	return db.withAll(dclose)
}

// SyncAll syncs all pogreb datastores.
func (db *DB) SyncAll() error {
	return db.withAll(dsync)
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
		if err = db.initStore(name, defaultPogrebOptions); err != nil {
			errs = append(errs, err)
			continue
		}
		stores = append(stores, name)
		db.store[name] = &Store{DB: db.store[name].DB}
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

func (db *DB) Type() string {
	return "pogreb"
}
