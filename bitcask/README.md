# bitcask

`import "git.tcp.direct/tcp.direct/database/bitcask"`

## Documentation

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
func (db *DB) AllStores() []database.Filer
```
AllStores returns a list of all bitcask datastores.

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

#### func (*DB) Init

```go
func (db *DB) Init(storeName string, opts ...any) error
```
Init opens a bitcask store at the given path to be referenced by storeName.

#### func (*DB) Path

```go
func (db *DB) Path() string
```
Path returns the base path where we store our bitcask "stores".

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

#### func (*DB) With

```go
func (db *DB) With(storeName string) database.Store
```
With calls the given underlying bitcask instance.

#### func (*DB) WithNew

```go
func (db *DB) WithNew(storeName string) database.Filer
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

#### func (Store) Backend

```go
func (s Store) Backend() any
```
Backend returns the underlying bitcask instance.

#### func (Store) Keys

```go
func (s Store) Keys() (keys [][]byte)
```
Keys will return all keys in the database as a slice of byte slices.

#### func (Store) PrefixScan

```go
func (s Store) PrefixScan(prefix string) (<-chan *kv.KeyValue, chan error)
```
PrefixScan will scan a Store for all keys that have a matching prefix of the
given string and return a map of keys and values. (map[Key]Value)

#### func (Store) Search

```go
func (s Store) Search(query string) (<-chan *kv.KeyValue, chan error)
```
Search will search for a given string within all values inside of a Store. Note,
type casting will be necessary. (e.g: []byte or string)

#### func (Store) ValueExists

```go
func (s Store) ValueExists(value []byte) (key []byte, ok bool)
```
ValueExists will check for the existence of a Value anywhere within the
keyspace; returning the first Key found, true if found || nil and false if not
found.
