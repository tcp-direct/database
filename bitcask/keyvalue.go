package bitcask

import (
	"bytes"
	"git.tcp.direct/tcp.direct/database"
)

// KeyValue represents a key and a value from a key/value store.
type KeyValue struct {
	Key   Key
	Value Value
}

// Key represents a key in a key/value store.
type Key struct {
	database.Key
	b []byte
}

// Bytes returns the raw byte slice form of the Key.
func (k Key) Bytes() []byte {
	return k.b
}

// String returns the string slice form of the Key.
func (k Key) String() string {
	return string(k.b)
}

// Equal determines if two keys are equal.
func (k Key) Equal(k2 Key) bool {
	return bytes.Equal(k.Bytes(), k2.Bytes())
}

// Value represents a value in a key/value store.
type Value struct {
	database.Value
	b []byte
}

// Bytes returns the raw byte slice form of the Value.
func (v Value) Bytes() []byte {
	return v.b
}

// String returns the string slice form of the Value.
func (v Value) String() string {
	return string(v.b)
}

// Equal determines if two values are equal.
func (v Value) Equal(v2 Value) bool {
	return bytes.Equal(v.Bytes(), v2.Bytes())
}
