package bitcask

import (
	"strings"
)

// Search will search for a given string within all values inside of a Store.
// Note, type casting will be necessary. (e.g: []byte or string)
func (c Store) Search(query string) ([]KeyValue, error) {
	var errs []error
	var res []KeyValue
	for _, key := range c.AllKeys() {
		raw, _ := c.Get(key)
		k := Key{b: key}
		v := Value{b: raw}
		if strings.Contains(string(raw), query) {
			res = append(res, KeyValue{Key: k, Value: v})
		}
	}
	return res, compoundErrors(errs)
}

// ValueExists will check for the existence of a Value anywhere within the keyspace, returning the Key and true if found, or nil and false if not found.
func (c Store) ValueExists(value []byte) (key []byte, ok bool) {
	var raw []byte
	var needle = Value{b: value}
	for _, k := range c.AllKeys() {
		raw, _ = c.Get(k)
		v := Value{b: raw}
		if v.Equal(needle) {
			ok = true
			return
		}
	}
	return
}

// PrefixScan will scan a Store for all keys that have a matching prefix of the given string
// and return a map of keys and values. (map[Key]Value)
func (c Store) PrefixScan(prefix string) ([]KeyValue, error) {
	var res []KeyValue
	err := c.Scan([]byte(prefix), func(key []byte) error {
		raw, _ := c.Get(key)
		k := Key{b: key}
		kv := KeyValue{Key: k, Value: Value{b: raw}}
		res = append(res, kv)

		return nil
	})
	return res, err
}

// AllKeys will return all keys in the database as a slice of byte slices.
func (c Store) AllKeys() (keys [][]byte) {
	keychan := c.Keys()
	for key := range keychan {
		keys = append(keys, key)
	}
	return
}
