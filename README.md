# database

[![Coverage](https://codecov.io/gh/yunginnanet/database/branch/master/graph/badge.svg)](https://codecov.io/gh/yunginnanet/database)
[![Build Status](https://github.com/yunginnanet/database/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/yunginnanet/database/actions/workflows/go.yml)

#### type Casket

```go
type Casket struct {
	*bitcask.Bitcask
	Searcher
}
```

Casket is an implmentation of a Filer and a Searcher using Bitcask.

#### func (Casket) AllKeys

```go
func (c Casket) AllKeys() (keys []string)
```

#### func (Casket) PrefixScan

```go
func (c Casket) PrefixScan(prefix string) map[string]interface{}
```

#### func (Casket) Search

```go
func (c Casket) Search(query string) map[string]interface{}
```

#### func (Casket) ValueExists

```go
func (c Casket) ValueExists(value []byte) (key []byte, ok bool)
```

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

#### func (*DB) Close

```go
func (db *DB) Close(bucketName string) error
```
Close is a simple shim for bitcask's Close function.

#### func (*DB) CloseAll

```go
func (db *DB) CloseAll() error
```
CloseAll closes all bitcask datastores.

#### func (*DB) Init

```go
func (db *DB) Init(bucketName string) error
```
Init opens a bitcask store at the given path to be referenced by bucketName.

#### func (*DB) Path

```go
func (db *DB) Path() string
```
Path returns the base path where we store our bitcask "buckets".

#### func (*DB) Sync

```go
func (db *DB) Sync(bucketName string) error
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
func (db *DB) With(bucketName string) Casket
```
With calls the given underlying bitcask instance.

#### func (*DB) WithAll

```go
func (db *DB) WithAll(action withAllAction) error
```
WithAll performs an action on all bitcask stores that we have open. In the case
of an error, WithAll will continue and return a compound form of any errors that
occurred. For now this is just for Close and Sync, thusly it does a hard lock on
the Keeper.

#### type Filer

```go
type Filer interface {

	// Has should return true if the given key has an associated value.
	Has(key []byte) bool
	// Get should retrieve the byte slice corresponding to the given key, and any associated errors upon failure.
	Get(key []byte) ([]byte, error)
	// Put should insert the value data in a way that is associated and can be retrieved by the given key data.
	Put(key []byte, value []byte) error
	// Delete should delete the key and the value associated with the given key, and return an error upon failure.
	Delete(key []byte) error
}
```

Filer is is a way to implement any generic key/value store. These functions
should be plug and play with most of the popular key/value store golang
libraries.

#### type Keeper

```go
type Keeper interface {
	// Path should return the base path where all buckets should be stored under. (likely as subdirectories)
	Path() string
	// Init should initialize our Filer at the given path, to be referenced and called by bucketName.
	Init(bucketName string) error
	// With provides access to the given bucketName by providing a pointer to the related Filer.
	With(bucketName string) Filer
	// Close should safely end any Filer operations of the given bucketName and close any relevant handlers.
	Close(bucketName string) error
	// Sync should take any volatile data and solidify it somehow if relevant. (ram to disk in most cases)
	Sync(bucketName string) error

	CloseAll() error
	SyncAll() error
}
```

Keeper will be in charge of the more meta operations involving Filers. This
includes operations like initialization, syncing to disk if applicable, and
backing up.

NOTE: Many key/value golang libraries will already implement this interface
already. This exists for more potential granular control in the case that they
don't. Otherwise you'd have to build a wrapper around an existing key/value
store to satisfy an overencompassing interface.

#### type Searcher

```go
type Searcher interface {
	// AllKeys must retrieve all keys in the datastore with the given bucketName.
	AllKeys() []string
	// PrefixScan must return all keys that begin with the given prefix.
	PrefixScan(prefix string) map[string]interface{}
	// Search must be able to search through the contents of our database and return a map of results.
	Search(query string) map[string]interface{}
	// ValueExists searches for an exact match of the given value and returns the key that contains it.
	ValueExists(value []byte) (key []byte, ok bool)
}
```

Searcher must be able to search through our datastore(s) with strings.
