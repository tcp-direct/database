# database

[![Coverage](https://codecov.io/gh/tcp-direct/database/branch/main/graph/badge.svg)](https://codecov.io/gh/tcp-direct/database/tree/main)
[![Build Status](https://github.com/tcp-direct/database/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/tcp-direct/database/actions/workflows/go.yml)

`import "github.com/tcp-direct/database"`

> [!WARNING]  
> This package is pre-v1 and the API is NOT stable!

## Documentation

```go
var ErrKeyNotFound = errors.New("key not found")
```

#### type Filer

```go
type Filer interface {

	// Backend returns the underlying key/value store.
	Backend() any

	// Has should return true if the given key has an associated value.
	Has(key []byte) bool
	// Get should retrieve the byte slice corresponding to the given key, and any associated errors upon failure.
	Get(key []byte) ([]byte, error)
	// Put should insert the value data in a way that is associated and can be retrieved by the given key data.
	Put(key []byte, value []byte) error
	// Delete should delete the key and the value associated with the given key, and return an error upon failure.
	Delete(key []byte) error
	// Close should safely end any Filer operations of the given dataStore and close any relevant handlers.
	Close() error
	// Sync should take any volatile data and solidify it somehow if relevant. (ram to disk in most cases)
	Sync() error

	Keys() [][]byte
	Len() int
}
```

Filer is is a way to implement any generic key/value store. These functions
should be plug and play with most of the popular key/value store golang
libraries.

NOTE: Many key/value golang libraries will already implement this interface
already. This exists for more potential granular control in the case that they
don't. Otherwise you'd have to build a wrapper around an existing key/value
store to satisfy an overencompassing interface.

#### type Keeper

```go
type Keeper interface {
	// Path should return the base path where all stores should be stored under. (likely as subdirectories)
	Path() string

	// Init should initialize our Filer at the given path, to be referenced and called by dataStore.
	Init(name string, options ...any) error
	// With provides access to the given dataStore by providing a pointer to the related Filer.
	With(name string) Filer
	// WithNew should initialize a new Filer at the given path and return a pointer to it.
	WithNew(name string, options ...any) Filer

	// Destroy should remove the Filer by the given name.
	// It is up to the implementation to decide if the data should be removed or not.
	Destroy(name string) error

	Discover() ([]string, error)

	AllStores() map[string]Filer

	// BackupAll should create a backup of all [Filer] instances in the [Keeper].
	BackupAll(archivePath string) (models.Backup, error)

	// RestoreAll should restore all [Filer] instances from the given archive.
	RestoreAll(archivePath string) error

	Meta() models.Metadata

	Close(name string) error

	CloseAll() error
	SyncAll() error
	SyncAndCloseAll() error
}
```

Keeper will be in charge of the more meta operations involving Filers. This
includes operations like initialization, syncing to disk if applicable, and
backing up.

    - When opening a folder of Filers, it should be able to discover and initialize all of them.
    - Additionally, it should be able to confirm the type of the underlying key/value store.

#### type KeeperCreator

```go
type KeeperCreator func(path string) (Keeper, error)
```


#### type MockFiler

```go
type MockFiler struct {
}
```


#### func (*MockFiler) Backend

```go
func (m *MockFiler) Backend() any
```

#### func (*MockFiler) Close

```go
func (m *MockFiler) Close() error
```

#### func (*MockFiler) Delete

```go
func (m *MockFiler) Delete(key []byte) error
```

#### func (*MockFiler) Get

```go
func (m *MockFiler) Get(key []byte) ([]byte, error)
```

#### func (*MockFiler) Has

```go
func (m *MockFiler) Has(key []byte) bool
```

#### func (*MockFiler) Keys

```go
func (m *MockFiler) Keys() [][]byte
```

#### func (*MockFiler) Len

```go
func (m *MockFiler) Len() int
```

#### func (*MockFiler) Put

```go
func (m *MockFiler) Put(key []byte, value []byte) error
```

#### func (*MockFiler) Sync

```go
func (m *MockFiler) Sync() error
```

#### type MockKeeper

```go
type MockKeeper struct {
}
```


#### func  NewMockKeeper

```go
func NewMockKeeper(name string) *MockKeeper
```

#### func (*MockKeeper) AllStores

```go
func (m *MockKeeper) AllStores() map[string]Filer
```

#### func (*MockKeeper) BackupAll

```go
func (m *MockKeeper) BackupAll(archivePath string) (models.Backup, error)
```

#### func (*MockKeeper) Close

```go
func (m *MockKeeper) Close(name string) error
```

#### func (*MockKeeper) CloseAll

```go
func (m *MockKeeper) CloseAll() error
```

#### func (*MockKeeper) Destroy

```go
func (m *MockKeeper) Destroy(name string) error
```

#### func (*MockKeeper) Discover

```go
func (m *MockKeeper) Discover() ([]string, error)
```

#### func (*MockKeeper) Init

```go
func (m *MockKeeper) Init(name string, options ...any) error
```

#### func (*MockKeeper) Meta

```go
func (m *MockKeeper) Meta() models.Metadata
```

#### func (*MockKeeper) Path

```go
func (m *MockKeeper) Path() string
```

#### func (*MockKeeper) RestoreAll

```go
func (m *MockKeeper) RestoreAll(archivePath string) error
```

#### func (*MockKeeper) SyncAll

```go
func (m *MockKeeper) SyncAll() error
```

#### func (*MockKeeper) SyncAndCloseAll

```go
func (m *MockKeeper) SyncAndCloseAll() error
```

#### func (*MockKeeper) With

```go
func (m *MockKeeper) With(name string) Filer
```

#### func (*MockKeeper) WithNew

```go
func (m *MockKeeper) WithNew(name string, options ...any) Filer
```

#### type Searcher

```go
type Searcher interface {
	// PrefixScan must retrieve all keys in the datastore and stream them to the given channel.
	PrefixScan(prefix string) (<-chan kv.KeyValue, chan error)
	// Search must be able to search through the value contents of our database and stream the results to the given channel.
	Search(query string) (<-chan kv.KeyValue, chan error)
	// ValueExists searches for an exact match of the given value and returns the key that contains it.
	ValueExists(value []byte) (key []byte, ok bool)
}
```

Searcher must be able to search through our datastore(s) with strings.

#### type Store

```go
type Store interface {
	Filer
	Searcher
}
```

Store is an implementation of a Filer and a Searcher.

#### func  ToStore

```go
func ToStore(filer Filer) (Store, error)
```

#### func  IsStore

```go
func IsStore(filer Filer) bool
```

