# bitcask



```go
var (
	ErrUnknownAction = errors.New("unknown action")
	ErrBogusStore    = errors.New("bogus store backend")
	ErrStoreExists   = errors.New("store name already exists")
	ErrNoStores      = errors.New("no stores initialized")
)
```

#### func  SetDefaultBitcaskOptions

```go
func SetDefaultBitcaskOptions(bitcaskopts ...bitcask.Option)
```
SetDefaultBitcaskOptions options will set the options used for all subsequent
bitcask stores that are initialized.

#### func  WithMaxDatafileSize

```go
func WithMaxDatafileSize(size int) bitcask.Option
```
WithMaxDatafileSize is a shim for bitcask's WithMaxDataFileSize function.

#### func  WithMaxKeySize

```go
func WithMaxKeySize(size uint32) bitcask.Option
```
WithMaxKeySize is a shim for bitcask's WithMaxKeySize function.

#### func  WithMaxValueSize

```go
func WithMaxValueSize(size uint64) bitcask.Option
```
WithMaxValueSize is a shim for bitcask's WithMaxValueSize function.

#### type DB

```go
type DB struct {
}
```

DB is a mapper of a Filer and Searcher implementation using Bitcask.

#### func  OpenDB

```go
func OpenDB(path string) *DB
```
OpenDB will either open an existing set of bitcask datastores at the given
directory, or it will create a new one.

#### func (*DB) AllStores

```go
func (db *DB) AllStores() map[string]database.Filer
```
AllStores returns a map of the names of all bitcask datastores and the
corresponding Filers.

#### func (*DB) BackupAll

```go
func (db *DB) BackupAll(archivePath string) (models.Backup, error)
```

#### func (*DB) Close

```go
func (db *DB) Close(storeName string) error
```
Close is a simple shim for bitcask's Close function.

#### func (*DB) CloseAll

```go
func (db *DB) CloseAll() error
```
CloseAll closes all bitcask datastores.

#### func (*DB) Destroy

```go
func (db *DB) Destroy(storeName string) error
```
Destroy will remove the bitcask store and all data associated with it.

#### func (*DB) Discover

```go
func (db *DB) Discover() ([]string, error)
```
Discover will discover and initialize all existing bitcask stores at the path
opened by [OpenDB].

#### func (*DB) Init

```go
func (db *DB) Init(storeName string, opts ...any) error
```
Init opens a bitcask store at the given path to be referenced by storeName.

#### func (*DB) Meta

```go
func (db *DB) Meta() models.Metadata
```
Meta returns the [models.Metadata] implementation of the bitcask keeper.

#### func (*DB) Path

```go
func (db *DB) Path() string
```
Path returns the base path where we store our bitcask "stores".

#### func (*DB) RestoreAll

```go
func (db *DB) RestoreAll(archivePath string) error
```

#### func (*DB) Sync

```go
func (db *DB) Sync(storeName string) error
```
Sync is a simple shim for bitcask's Sync function.

#### func (*DB) SyncAll

```go
func (db *DB) SyncAll() error
```
SyncAll syncs all bitcask datastores.

#### func (*DB) SyncAndCloseAll

```go
func (db *DB) SyncAndCloseAll() error
```
SyncAndCloseAll implements the method from Keeper to sync and close all bitcask
stores.

#### func (*DB) Type

```go
func (db *DB) Type() string
```
Type returns the type of keeper, in this case "bitcask". This is in order to
implement [database.Keeper].

#### func (*DB) With

```go
func (db *DB) With(storeName string) database.Filer
```
With calls the given underlying bitcask instance.

#### func (*DB) WithNew

```go
func (db *DB) WithNew(storeName string, opts ...any) database.Filer
```
WithNew calls the given underlying bitcask instance, if it doesn't exist, it
creates it.

#### type Store

```go
type Store struct {
	*bitcask.Bitcask
	database.Searcher
}
```

Store is an implmentation of a Filer and a Searcher using Bitcask.

#### func (*Store) Backend

```go
func (s *Store) Backend() any
```
Backend returns the underlying bitcask instance.

#### func (*Store) Close

```go
func (s *Store) Close() error
```
Close is a wrapper around the bitcask Close function.

#### func (*Store) Get

```go
func (s *Store) Get(key []byte) ([]byte, error)
```
Get is a wrapper around the bitcask Get function for error regularization.

#### func (*Store) Keys

```go
func (s *Store) Keys() (keys [][]byte)
```
Keys will return all keys in the database as a slice of byte slices.

#### func (*Store) PrefixScan

```go
func (s *Store) PrefixScan(prefix string) (<-chan kv.KeyValue, chan error)
```
PrefixScan will scan a Store for all keys that have a matching prefix of the
given string and return a map of keys and values. (map[Key]Value)

#### func (*Store) Search

```go
func (s *Store) Search(query string) (<-chan kv.KeyValue, chan error)
```
Search will search for a given string within all values inside of a Store. Note,
type casting will be necessary. (e.g: []byte or string)

#### func (*Store) ValueExists

```go
func (s *Store) ValueExists(value []byte) (key []byte, ok bool)
```
ValueExists will check for the existence of a Value anywhere within the
keyspace; returning the first Key found, true if found || nil and false if not
found.

---
