package database

import (
	"errors"
	"sync"

	"git.tcp.direct/tcp.direct/database/models"
)

type MockFiler struct {
	name   string
	values map[string][]byte
	closed bool
	mu     sync.RWMutex
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

type MockKeeper struct {
	name   string
	path   string
	stores map[string]Filer
	mu     sync.RWMutex
}

func NewMockKeeper(name string) *MockKeeper {
	return &MockKeeper{
		name:   name,
		stores: make(map[string]Filer),
	}
}

func (m *MockKeeper) Path() string {
	return m.path
}

func (m *MockKeeper) Init(name string, options ...any) error {
	m.mu.Lock()
	m.stores[name] = &MockFiler{name: name, values: make(map[string][]byte)}
	m.mu.Unlock()
	return nil
}

func (m *MockKeeper) With(name string) Filer {
	m.mu.RLock()
	s, ok := m.stores[name]
	m.mu.RUnlock()
	if !ok {
		return nil
	}
	return s
}

func (m *MockKeeper) WithNew(name string, options ...any) Filer {
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
	m.mu.RLock()
	names := make([]string, 0, len(m.stores))
	for name := range m.stores {
		names = append(names, name)
	}
	m.mu.RUnlock()
	return names, nil
}

func (m *MockKeeper) AllStores() map[string]Filer {
	m.mu.RLock()
	stores := make(map[string]Filer, len(m.stores))
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
	panic("not implemented")
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
