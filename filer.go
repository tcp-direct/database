package database

// Filer is is a way to implement any generic key/value store.
// These functions should be plug and play with most of the popular key/value store golang libraries.
//
// NOTE: Many key/value golang libraries will already implement this interface already.
// This exists for more potential granular control in the case that they don't.
// Otherwise you'd have to build a wrapper around an existing key/value store to satisfy an overencompassing interface.
type Filer interface {
	// NOTE: One can easily cast anything to a byte slice. (e.g: []byte("fuckholejones") )
	// json.Marshal also returns a byte slice by default ;)

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
