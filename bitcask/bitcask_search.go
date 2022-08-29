package bitcask

import (
	"strings"

	"git.tcp.direct/tcp.direct/database/kv"
)

// Search will search for a given string within all values inside of a Store.
// Note, type casting will be necessary. (e.g: []byte or string)
func (s Store) Search(query string) (<-chan *kv.KeyValue, chan error) {
	var errChan = make(chan error)
	var resChan = make(chan *kv.KeyValue, 5)
	go func() {
		defer func() {
			close(resChan)
			close(errChan)
		}()
		for _, key := range s.Keys() {
			raw, err := s.Get(key)
			if err != nil {
				errChan <- err
				continue
			}
			if raw != nil && strings.Contains(string(raw), query) {
				keyVal := kv.NewKeyValue(kv.NewKey(key), kv.NewValue(raw))
				resChan <- keyVal
			}
		}
	}()
	return resChan, errChan
}

// ValueExists will check for the existence of a Value anywhere within the keyspace;
// returning the first Key found, true if found || nil and false if not found.
func (s Store) ValueExists(value []byte) (key []byte, ok bool) {
	var raw []byte
	var needle = kv.NewValue(value)
	for _, key = range s.Keys() {
		raw, _ = s.Get(key)
		v := kv.NewValue(raw)
		if v.Equal(needle) {
			ok = true
			return
		}
	}
	return
}

// PrefixScan will scan a Store for all keys that have a matching prefix of the given string
// and return a map of keys and values. (map[Key]Value)
func (s Store) PrefixScan(prefix string) (<-chan *kv.KeyValue, chan error) {
	errChan := make(chan error)
	resChan := make(chan *kv.KeyValue, 5)
	go func() {
		var err error
		defer func(e error) {
			if e != nil {
				errChan <- e
			}
			close(resChan)
			close(errChan)
		}(err)
		err = s.Scan([]byte(prefix), func(key []byte) error {
			raw, _ := s.Get(key)
			if key != nil && raw != nil {
				k := kv.NewKey(key)
				resChan <- kv.NewKeyValue(k, kv.NewValue(raw))
			}
			return nil
		})
	}()
	return resChan, errChan
}

// Keys will return all keys in the database as a slice of byte slices.
func (s Store) Keys() (keys [][]byte) {
	allkeys := s.Bitcask.Keys()
	for key := range allkeys {
		keys = append(keys, key)
	}
	return
}
