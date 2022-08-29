# database

[![Coverage](https://codecov.io/gh/tcp-direct/database/branch/master/graph/badge.svg)](https://codecov.io/gh/tcp-direct/database)
[![Build Status](https://github.com/tcp-direct/database/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/tcp-direct/database/actions/workflows/go.yml)

`import "git.tcp.direct/tcp.direct/database"`

## Documentation

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
	With(name string) Store

	AllStores() map[string]Filer

	CloseAll() error
	SyncAll() error
}
```

Keeper will be in charge of the more meta operations involving Filers. This
includes operations like initialization, syncing to disk if applicable, and
backing up.

#### type Searcher

```go
type Searcher interface {
	// PrefixScan must retrieve all keys in the datastore and stream them to the given channel.
	PrefixScan(prefix string) (<-chan *kv.KeyValue, chan error)
	// Search must be able to search through the value contents of our database and stream the results to the given channel.
	Search(query string) (<-chan *kv.KeyValue, chan error)
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
