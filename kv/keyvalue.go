package kv

import (
	"bytes"
)

// KeyValue represents a key and a value from a key/value store.
type KeyValue struct {
	Key   *Key
	Value *Value
}

// Key represents a key in a key/value store.
type Key struct {
	b []byte
}

// NewKey creates a new Key from a byte slice.
func NewKey(data []byte) *Key {
	k := Key{b: data}
	return &k
}

// NewValue creates a new Value from a byte slice.
func NewValue(data []byte) *Value {
	v := Value{b: data}
	return &v
}

// NewKeyValue creates a new KeyValue from a key and value.
func NewKeyValue(k *Key, v *Value) *KeyValue {
	return &KeyValue{Key: k, Value: v}
}

func (kv *KeyValue) String() string {
	return kv.Key.String() + ":" + kv.Value.String()
}

// Equal determines if two key/value pairs are equal.
func (kv *KeyValue) Equal(kv2 *KeyValue) bool {
	return kv.Key.Equal(kv2.Key) && kv.Value.Equal(kv2.Value)
}

// Bytes returns the raw byte slice form of the Key.
func (k *Key) Bytes() []byte {
	return k.b
}

// String returns the string slice form of the Key.
func (k *Key) String() string {
	return string(k.b)
}

// Equal determines if two keys are equal.
func (k *Key) Equal(k2 *Key) bool {
	return bytes.Equal(k.Bytes(), k2.Bytes())
}

// Value represents a value in a key/value store.
type Value struct {
	b []byte
}

// Bytes returns the raw byte slice form of the Value.
func (v *Value) Bytes() []byte {
	return v.b
}

// String returns the string slice form of the Value.
func (v *Value) String() string {
	return string(v.b)
}

// Equal determines if two values are equal.
func (v *Value) Equal(v2 *Value) bool {
	return bytes.Equal(v.Bytes(), v2.Bytes())
}
