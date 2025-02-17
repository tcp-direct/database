# pogreb



```go
var (
	ErrUnknownAction = errors.New("unknown action")
	ErrBogusStore    = errors.New("bogus store backend")
	ErrBadOptions    = errors.New("invalid pogreb options")
	ErrStoreExists   = errors.New("store name already exists")
	ErrNoStores      = errors.New("no stores initialized")
)
```

```go
var OptionAllowRecovery = func(opts *WrappedOptions) {
	opts.AllowRecovery = true
}
```

#### func  SetDefaultPogrebOptions

```go
func SetDefaultPogrebOptions(pogrebopts ...any)
```
SetDefaultPogrebOptions options will set the options used for all subsequent
pogreb stores that are initialized.

#### type CombinedMetrics

```go
type CombinedMetrics struct {
	Puts           int64 `json:"puts"`
	Dels           int64 `json:"dels"`
	Gets           int64 `json:"gets"`
	HashCollisions int64 `json:"hash_collisions"`
}
```


#### func  CombineMetrics

```go
func CombineMetrics(metrics ...*pogreb.Metrics) *CombinedMetrics
```

#### func (*CombinedMetrics) Equal

```go
func (cm *CombinedMetrics) Equal(other *CombinedMetrics) bool
```

#### type DB

```go
type DB struct {
}
```

DB is a mapper of a Filer and Searcher implementation using pogreb.

#### func  OpenDB

```go
func OpenDB(path string) *DB
```
OpenDB will either open an existing set of pogreb datastores at the given
directory, or it will create a new one.

#### func (*DB) AllStores

```go
func (db *DB) AllStores() map[string]database.Filer
```
AllStores returns a map of the names of all pogreb datastores and the
corresponding Filers.

#### func (*DB) BackupAll

```go
func (db *DB) BackupAll(archivePath string) (models.Backup, error)
```

#### func (*DB) Close

```go
func (db *DB) Close(storeName string) error
```
Close is a simple shim for pogreb's Close function.

#### func (*DB) CloseAll

```go
func (db *DB) CloseAll() error
```
CloseAll closes all pogreb datastores.

#### func (*DB) Destroy

```go
func (db *DB) Destroy(name string) error
```
Destroy will remove a pogreb store and all data associated with it.

#### func (*DB) Discover

```go
func (db *DB) Discover() ([]string, error)
```
Discover will discover and initialize all existing pogreb stores at the path
opened by [OpenDB].

#### func (*DB) Init

```go
func (db *DB) Init(storeName string, opts ...any) error
```
Init opens a pogreb store at the given path to be referenced by storeName.

#### func (*DB) Meta

```go
func (db *DB) Meta() models.Metadata
```
Meta returns the metadata for the pogreb database.

#### func (*DB) Path

```go
func (db *DB) Path() string
```
Path returns the base path where we store our pogreb "stores".

#### func (*DB) RestoreAll

```go
func (db *DB) RestoreAll(archivePath string) error
```

#### func (*DB) Sync

```go
func (db *DB) Sync(storeName string) error
```
Sync is a simple shim for pogreb's Sync function.

#### func (*DB) SyncAll

```go
func (db *DB) SyncAll() error
```
SyncAll syncs all pogreb datastores.

#### func (*DB) SyncAndCloseAll

```go
func (db *DB) SyncAndCloseAll() error
```
SyncAndCloseAll implements the method from Keeper to sync and close all pogreb
stores.

#### func (*DB) Type

```go
func (db *DB) Type() string
```

#### func (*DB) UpdateMetrics

```go
func (db *DB) UpdateMetrics()
```

#### func (*DB) With

```go
func (db *DB) With(storeName string) database.Filer
```
With calls the given underlying pogreb instance.

#### func (*DB) WithNew

```go
func (db *DB) WithNew(storeName string, opts ...any) database.Filer
```
WithNew calls the given underlying pogreb instance, if it doesn't exist, it
creates it.

#### type Option

```go
type Option func(*WrappedOptions)
```


#### func  AllowRecovery

```go
func AllowRecovery() Option
```

#### func  SetPogrebOptions

```go
func SetPogrebOptions(options pogreb.Options) Option
```

#### type Store

```go
type Store struct {
	*pogreb.DB
	database.Searcher
}
```

Store is an implmentation of a Filer and a Searcher using pogreb.

#### func (*Store) Backend

```go
func (pstore *Store) Backend() any
```
Backend returns the underlying pogreb instance.

#### func (*Store) Close

```go
func (pstore *Store) Close() error
```
Close is a simple shim for pogreb's Close function.

#### func (*Store) Get

```go
func (pstore *Store) Get(key []byte) ([]byte, error)
```
Get is a wrapper for pogreb's Get function to regularize errors when keys do not
exist.

#### func (*Store) Has

```go
func (pstore *Store) Has(key []byte) bool
```

#### func (*Store) Keys

```go
func (pstore *Store) Keys() [][]byte
```

#### func (*Store) Len

```go
func (pstore *Store) Len() int
```

#### func (*Store) PrefixScan

```go
func (pstore *Store) PrefixScan(prefixs string) (<-chan kv.KeyValue, chan error)
```
PrefixScan will scan a Store for all keys that have a matching prefix of the
given string and return a map of keys and values. (map[Key]Value) error channel
will block, so be sure to read from it.

#### func (*Store) Search

```go
func (pstore *Store) Search(query string) (<-chan kv.KeyValue, chan error)
```
Search will search for a given string within all values inside of a Store. Note,
type casting will be necessary. (e.g: []byte or string)

#### func (*Store) ValueExists

```go
func (pstore *Store) ValueExists(value []byte) (key []byte, ok bool)
```
ValueExists will check for the existence of a Value anywhere within the
keyspace; returning the first Key found, true if found || nil and false if not
found.

#### type WrappedOptions

```go
type WrappedOptions struct {
	*pogreb.Options
	// AllowRecovery allows the database to be recovered if a lockfile is detected upon running Init.
	AllowRecovery bool
}
```


#### func (*WrappedOptions) MarshalJSON

```go
func (w *WrappedOptions) MarshalJSON() ([]byte, error)
```

---
