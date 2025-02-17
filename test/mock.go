package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/tcp-direct/database"
	"github.com/tcp-direct/database/metadata"
	"github.com/tcp-direct/database/registry"

	"github.com/tcp-direct/database/models"
)

var (
	mockKeepers = make(map[string]map[string]database.Filer)
	mockMu      sync.RWMutex
)

type MockFiler struct {
	name   string
	values map[string][]byte
	closed bool
	Opts   []MockOpt
	mu     sync.RWMutex
}

func (m *MockKeeper) WriteMeta(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()
	mockMu.Lock()
	mockKeepers[m.name] = m.AllStores()
	mockMu.Unlock()
	registry.RegisterKeeper(m.name, func(path string, opts ...any) (database.Keeper, error) {
		return NewMockKeeper(m.name, opts...), nil
	})

	return json.NewEncoder(f).Encode(m.Meta())
}

func (m *MockFiler) Backend() any {
	return m.values
}

func (m *MockFiler) Has(key []byte) bool {
	m.mu.RLock()
	_, ok := m.values[string(key)]
	m.mu.RUnlock()
	return ok
}

var ErrKeyNotFound = errors.New("key not found")

func (m *MockFiler) Get(key []byte) ([]byte, error) {
	if !m.Has(key) {
		return nil, ErrKeyNotFound
	}
	m.mu.RLock()
	val := m.values[string(key)]
	m.mu.RUnlock()
	return val, nil
}

func (m *MockFiler) Put(key []byte, value []byte) error {
	m.mu.Lock()
	m.values[string(key)] = value
	m.mu.Unlock()
	return nil
}

func (m *MockFiler) Delete(key []byte) error {
	m.mu.Lock()
	delete(m.values, string(key))
	m.mu.Unlock()
	return nil
}

func (m *MockFiler) Close() error {
	m.mu.Lock()
	m.closed = true
	m.mu.Unlock()
	return nil
}

func (m *MockFiler) Sync() error {
	return nil
}

func (m *MockFiler) Keys() [][]byte {
	m.mu.RLock()
	k := make([][]byte, 0, len(m.values))
	for key := range m.values {
		k = append(k, []byte(key))
	}
	m.mu.RUnlock()
	return k
}

func (m *MockFiler) Len() int {
	m.mu.RLock()
	l := len(m.values)
	m.mu.RUnlock()
	return l
}

type MockOpt string

type MockKeeper struct {
	name    string
	path    string
	defOpts []MockOpt
	stores  map[string]database.Filer
	mu      sync.RWMutex
}

func NewMockKeeper(name string, defopts ...any) *MockKeeper {
	opts := make([]MockOpt, 0, len(defopts))
	for _, opt := range defopts {
		if opt == nil {
			println("nil opt")
			continue
		}
		if strOpt, strOK := opt.(string); strOK {
			opt = MockOpt(strOpt)
		}
		if _, ok := opt.(MockOpt); !ok {
			panic(fmt.Errorf("%w: (%T): %v", ErrBadOptions, opt, opt))
		}
		opts = append(opts, opt.(MockOpt))
	}

	mk := &MockKeeper{
		name:   name,
		stores: make(map[string]database.Filer),
	}

	if len(opts) > 0 {
		mk.defOpts = opts
	}

	return mk
}

func (m *MockKeeper) Path() string {
	return m.path
}

var ErrBadOptions = errors.New("bad mock filer options")

func (m *MockKeeper) Init(name string, options ...any) error {
	m.mu.Lock()
	m.stores[name] = &MockFiler{name: name, values: make(map[string][]byte)}
	if len(options) > 0 {
		for _, opt := range options {
			strOpt, strOK := opt.(string)
			mockOpt, mockOK := opt.(MockOpt)
			if !strOK && !mockOK {
				return ErrBadOptions
			}
			if strOK {
				mockOpt = MockOpt(strOpt)
			}
			m.stores[name].(*MockFiler).Opts = append(m.stores[name].(*MockFiler).Opts, mockOpt)
		}
	} else if m.defOpts != nil {
		m.stores[name].(*MockFiler).Opts = append(m.stores[name].(*MockFiler).Opts, m.defOpts...)
	}
	m.mu.Unlock()
	return nil
}

func (m *MockKeeper) With(name string) database.Filer {
	m.mu.RLock()
	s, ok := m.stores[name]
	m.mu.RUnlock()
	if !ok {
		return nil
	}
	return s
}

func (m *MockKeeper) WithNew(name string, options ...any) database.Filer {
	m.mu.RLock()
	existing, ok := m.stores[name]
	m.mu.RUnlock()
	if ok {
		return existing
	}
	m.mu.Lock()
	m.stores[name] = &MockFiler{name: name, values: make(map[string][]byte)}
	m.mu.Unlock()
	return m.stores[name]
}

func (m *MockKeeper) Destroy(name string) error {
	m.mu.RLock()
	_, ok := m.stores[name]
	m.mu.RUnlock()
	if !ok {
		return errors.New("store not found")
	}
	m.mu.Lock()
	delete(m.stores, name)
	m.mu.Unlock()
	return nil
}

func (m *MockKeeper) Discover() ([]string, error) {
	getStores := func() []string {
		names := make([]string, 0, len(m.stores))
		for name := range m.stores {
			names = append(names, name)
		}
		return names
	}

	m.mu.RLock()
	if m.stores != nil && len(m.stores) > 0 {
		m.mu.RUnlock()
		return getStores(), nil
	}
	mockMu.RLock()
	stores, ok := mockKeepers[m.name]
	mockMu.RUnlock()
	if ok {
		m.mu.RUnlock()
		m.mu.Lock()
		for _, s := range stores {
			if m.defOpts != nil {
				for _, v := range m.defOpts {
					if !slices.Contains(s.(*MockFiler).Opts, v) {
						s.(*MockFiler).Opts = append(s.(*MockFiler).Opts, v)
					}
				}
			}
		}
		m.stores = stores
		m.mu.Unlock()
		m.mu.RLock()
	}

	m.mu.RUnlock()
	return getStores(), nil
}

func (m *MockKeeper) AllStores() map[string]database.Filer {
	m.mu.RLock()
	stores := make(map[string]database.Filer, len(m.stores))
	for name, store := range m.stores {
		stores[name] = store
	}
	m.mu.RUnlock()
	return stores
}

func (m *MockKeeper) BackupAll(archivePath string) (models.Backup, error) {
	panic("not implemented")
}

func (m *MockKeeper) RestoreAll(archivePath string) error {
	panic("not implemented")
}

func (m *MockKeeper) Meta() models.Metadata {
	st := m.AllStores()
	stores := make([]string, 0, len(st))
	for name := range st {
		stores = append(stores, name)
	}
	return metadata.NewMeta(metadata.KeeperType(m.name)).
		WithCreated(time.Now()).
		WithLastOpened(time.Now()).
		WithStores(stores...)
}

func (m *MockKeeper) Close(name string) error {
	m.mu.RLock()
	store, ok := m.stores[name]
	m.mu.RUnlock()
	if !ok {
		return errors.New("store not found")
	}
	return store.Close()
}

func (m *MockKeeper) CloseAll() error {
	m.mu.RLock()
	for _, store := range m.stores {
		if err := store.Close(); err != nil {
			return err
		}
	}
	m.mu.RUnlock()
	return nil
}

func (m *MockKeeper) SyncAll() error {
	return nil
}

func (m *MockKeeper) SyncAndCloseAll() error {
	return m.CloseAll()
}
