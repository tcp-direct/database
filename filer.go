package database

// Filer is is a way to implement any generic key/value store.
// These functions should be plug and play with most of the popular key/value store golang libraries.
type Filer interface {
	// NOTE: One can easily cast anything to a byte slice. (e.g: []byte("fuckholejones") )
	// json.Marshal also returns a byte slice by default ;)

	// Has should return true if the given key has an associated value.
	Has(key []byte) bool
	// Get should retrieve the byte slice corresponding to the given key, and any associated errors upon failure.
	Get(key []byte) ([]byte, error)
	// Put should insert the value data in a way that is associated and can be retrieved by the given key data.
	Put(key []byte, value []byte) error
	// Delete should delete the key and the value associated with the given key, and return an error upon failure.
	Delete(key []byte) error
}
