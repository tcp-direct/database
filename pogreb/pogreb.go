package pogreb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/akrylysov/pogreb"

	"git.tcp.direct/tcp.direct/database"
	"git.tcp.direct/tcp.direct/database/kv"
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

func (w *WrappedOptions) MarshalJSON() ([]byte, error) {
	optData, err := json.Marshal(w.Options)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Options       json.RawMessage `json:"options"`
		AllowRecovery bool            `json:"allow_recovery"`
	}{
		Options:       optData,
		AllowRecovery: w.AllowRecovery,
	})
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

type Metrics struct {
	Puts           int64 `json:"puts"`
	Dels           int64 `json:"dels"`
	Gets           int64 `json:"gets"`
	HashCollisions int64 `json:"hash_collisions"`
}

// Store is an implmentation of a Filer and a Searcher using Bitcask.
type Store struct {
	*pogreb.DB
	opts    *WrappedOptions
	closed  *atomic.Bool
	metrics *pogreb.Metrics
}

// Backend returns the underlying pogreb instance.
func (pstore *Store) Backend() any {
	return pstore.DB
}

// Get is a wrapper for pogreb's Get function to regularize errors when keys do not exist.
func (pstore *Store) Get(key []byte) ([]byte, error) {
	if pstore.closed.Load() {
		return nil, fs.ErrClosed
	}
	ret, err := pstore.DB.Get(key)
	if err = kv.RegularizeKVError(key, ret, err); err != nil {
		return nil, err
	}
	return ret, err
}

// Close is a simple shim for pogreb's Close function.
func (pstore *Store) Close() error {
	if pstore.closed.Load() {
		return fs.ErrClosed
	}
	pstore.closed.Store(true)
	pstore.metrics = pstore.DB.Metrics()
	return pstore.DB.Close()
}

// DB is a mapper of a Filer and Searcher implementation using pogreb.
type DB struct {
	store map[string]*Store
	path  string
	mu    *sync.RWMutex
	meta  *metadata.Metadata

	initialized *atomic.Bool
}

type CombinedMetrics struct {
	Puts           int64 `json:"puts"`
	Dels           int64 `json:"dels"`
	Gets           int64 `json:"gets"`
	HashCollisions int64 `json:"hash_collisions"`
}

func CombineMetrics(metrics ...*pogreb.Metrics) *CombinedMetrics {
	var c = &CombinedMetrics{}
	for _, m := range metrics {
		c.Puts += m.Puts.Value()
		c.Dels += m.Dels.Value()
		c.Gets += m.Gets.Value()
		c.HashCollisions += m.HashCollisions.Value()
	}
	return c
}

func (cm *CombinedMetrics) Equal(other *CombinedMetrics) bool {
	return cm.Puts == other.Puts && cm.Dels == other.Dels && cm.Gets == other.Gets && cm.HashCollisions == other.HashCollisions
}

// Meta returns the metadata for the pogreb database.
func (db *DB) Meta() models.Metadata {
	db.mu.RLock()
	if len(db.store) > 1 && db.meta != nil {
		mets := make([]*pogreb.Metrics, 0, len(db.store))
		for _, s := range db.store {
			s.metrics = s.DB.Metrics()
			mets = append(mets, s.metrics)
		}
		if db.meta.Extra == nil {
			db.meta.Extra = make(map[string]interface{})
		}
		newMet := CombineMetrics(mets...)
		if db.meta.Extra["metrics"] == nil || !newMet.Equal(db.meta.Extra["metrics"].(*CombinedMetrics)) {
			db.meta.Extra["metrics"] = newMet
			_ = db.meta.Sync()
		}
	}
	m := db.meta
	db.mu.RUnlock()
	if m == nil {
		return metadata.NewPlaceholder("pogreb")
	}
	return m
}

func (db *DB) updateMetrics() {
	for _, s := range db.store {
		if s != nil && s.DB != nil {
			s.metrics = s.DB.Metrics()
		}
	}
}

func (db *DB) UpdateMetrics() {
	db.mu.RLock()
	db.updateMetrics()
	db.mu.RUnlock()
}

func (db *DB) allStores() map[string]database.Filer {
	var stores = make(map[string]database.Filer)
	for n, s := range db.store {
		stores[n] = s
	}
	return stores
}

// AllStores returns a map of the names of all pogreb datastores and the corresponding Filers.
func (db *DB) AllStores() map[string]database.Filer {

	db.mu.RLock()
	ast := db.allStores()
	db.mu.RUnlock()
	return ast
}

// FIXME: not returning the error is probably pretty irresponsible.

// OpenDB will either open an existing set of pogreb datastores at the given directory, or it will create a new one.
func OpenDB(path string) *DB {
	ainit := &atomic.Bool{}
	ainit.Store(false)
	db := &DB{
		store: make(map[string]*Store),
		path:  path,
		mu:    &sync.RWMutex{},
		meta:  nil,

		initialized: ainit,
	}
	return db
}
func (db *DB) _init() error {
	if _, err := os.Stat(db.path); os.IsNotExist(err) {
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

func (db *DB) init() error {
	if db.initialized.Load() {
		return nil
	}
	db.mu.Lock()
	defer db.mu.Unlock()
	return db._init()
}

// Destroy will remove a pogreb store and all data associated with it.
func (db *DB) Destroy(name string) error {
	db.mu.Lock()
	defer db.mu.Unlock()
	if store, ok := db.store[name]; !ok {
		return fmt.Errorf("store %s does not exist", name)
	} else {
		_ = store.Close()
	}
	delete(db.store, name)
	err := os.RemoveAll(filepath.Join(db.path, name))
	if err != nil {
		err = fmt.Errorf("error removing pogreb store's data: %w", err)
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
	default: // case pgoptWrappedOk:
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
	aclosed := &atomic.Bool{}
	aclosed.Store(false)
	db.store[storeName] = &Store{DB: c, closed: aclosed, opts: pogrebOpts}
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
func (db *DB) With(storeName string) database.Filer {
	if err := db.init(); err != nil {
		panic(err)
	}
	db.mu.RLock()
	d, ok := db.store[storeName]
	if ok {
		if d.closed == nil || d.DB == nil || d.closed.Load() {
			db.mu.RUnlock()
			db.mu.Lock()
			delete(db.store, storeName)
			if err := db.initStore(storeName, defaultPogrebOptions); err != nil {
				_, _ = os.Stderr.WriteString("error creating pogreb store: " + err.Error())
				db.mu.Unlock()
				return nil
			}
			db.mu.Unlock()
			return db.store[storeName]
		}
		db.mu.RUnlock()
		return d
	}
	db.mu.RUnlock()
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
	return nil
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
			delete(db.store, name)
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

func (db *DB) syncAndCloseAll() error {
	if db == nil || db.store == nil || len(db.store) < 1 {
		return ErrNoStores
	}
	var errs = make([]error, len(db.store))
	errSync := namedErr("sync", db.SyncAll())
	if errSync != nil {
		errs = append(errs, errSync)
	}
	errClose := namedErr("close", db.closeAll())
	if errClose != nil {
		errs = append(errs, errClose)
	}
	errs = append(errs, errSync, errClose, db.meta.Sync())
	return compoundErrors(errs)
}

// SyncAndCloseAll implements the method from Keeper to sync and close all bitcask stores.
func (db *DB) SyncAndCloseAll() error {
	db.mu.Lock()
	err := db.syncAndCloseAll()
	db.mu.Unlock()
	return err
}

func (db *DB) closeAll() error {
	return db.withAll(dclose)
}

// CloseAll closes all bitcask datastores.
func (db *DB) CloseAll() error {
	db.mu.Lock()
	err := db.closeAll()
	db.mu.Unlock()
	return err
}

func (db *DB) allMetrics() map[string]*pogreb.Metrics {
	db.updateMetrics()
	allmet := make(map[string]*pogreb.Metrics, len(db.store))
	for name, store := range db.store {
		if store == nil || store.Backend().(*pogreb.DB) == nil {
			continue
		}
		allmet[name] = store.metrics
	}
	return allmet
}

func (db *DB) addAllStoresToMeta() {
	storeMap := db.allStores()
	storeNames := make([]string, len(storeMap))
	for name := range storeMap {
		storeNames = append(storeNames, name)
	}
	db.meta = db.meta.WithStores(storeNames...)
}

func (db *DB) syncMetaValues() {
	db.addAllStoresToMeta()
	db.meta = db.meta.WithExtra(map[string]interface{}{"metrics": db.allMetrics()})
}

// SyncAll syncs all pogreb datastores.
// TODO: investigate locking here, right now if we try to hold a lock during a backup we'll hang :^)
func (db *DB) SyncAll() error {
	db.syncMetaValues()
	var errs = make([]error, 0)
	errs = append(errs, db.withAll(dsync))
	errs = append(errs, db.meta.Sync())
	return compoundErrors(errs)
}

func (db *DB) discover() ([]string, error) {
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
		aclosed := &atomic.Bool{}
		aclosed.Store(false)
		db.store[name] = &Store{DB: db.store[name].DB, closed: aclosed, opts: defaultPogrebOptions}
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

// Discover will discover and initialize all existing bitcask stores at the path opened by [OpenDB].
func (db *DB) Discover() ([]string, error) {
	if err := db.init(); err != nil {
		return nil, err
	}
	db.mu.Lock()
	ret, err := db.discover()
	db.mu.Unlock()
	return ret, err
}

func (db *DB) Type() string {
	return "pogreb"
}
