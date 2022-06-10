# bitcask

`import "git.tcp.direct/tcp.direct/database/bitcask"`

## Documentation

#### type DB

```go
type DB struct {}
```

DB is a mapper of a Filer and Searcher implementation using Bitcask.

#### func  OpenDB

```go
func OpenDB(path string) *DB
```
OpenDB will either open an existing set of bitcask datastores at the given
directory, or it will create a new one.

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
func (db *DB) Init(storeName string, bitcaskopts ...bitcask.Option) error
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
SyncAndCloseAll implements the method from Keeper.

#### func (*DB) With

```go
func (db *DB) With(storeName string) Store
```
With calls the given underlying bitcask instance.

#### type Key

```go
type Key struct {
	database.Key
}
```

Key represents a key in a key/value store.

#### func (Key) Bytes

```go
func (k Key) Bytes() []byte
```
Bytes returns the raw byte slice form of the Key.

#### func (Key) Equal

```go
func (k Key) Equal(k2 Key) bool
```
Equal determines if two keys are equal.

#### func (Key) String

```go
func (k Key) String() string
```
String returns the string slice form of the Key.

#### type KeyValue

```go
type KeyValue struct {
	Key   Key
	Value Value
}
```

KeyValue represents a key and a value from a key/value store.

#### type Store

```go
type Store struct {
	*bitcask.Bitcask
	database.Searcher
}
```

Store is an implmentation of a Filer and a Searcher using Bitcask.

#### func (Store) AllKeys

```go
func (c Store) AllKeys() (keys [][]byte)
```
AllKeys will return all keys in the database as a slice of byte slices.

#### func (Store) PrefixScan

```go
func (c Store) PrefixScan(prefix string) ([]KeyValue, error)
```
PrefixScan will scan a Store for all keys that have a matching prefix of the
given string and return a map of keys and values. (map[Key]Value)

#### func (Store) Search

```go
func (c Store) Search(query string) ([]KeyValue, error)
```
Search will search for a given string within all values inside of a Store. Note,
type casting will be necessary. (e.g: []byte or string)

#### func (Store) ValueExists

```go
func (c Store) ValueExists(value []byte) (key []byte, ok bool)
```
ValueExists will check for the existence of a Value anywhere within the
keyspace, returning the Key and true if found, or nil and false if not found.

#### type Value

```go
type Value struct {
	database.Value
}
```

Value represents a value in a key/value store.

#### func (Value) Bytes

```go
func (v Value) Bytes() []byte
```
Bytes returns the raw byte slice form of the Value.

#### func (Value) Equal

```go
func (v Value) Equal(v2 Value) bool
```
Equal determines if two values are equal.

#### func (Value) String

```go
func (v Value) String() string
```
String returns the string slice form of the Value.
