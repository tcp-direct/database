package pogreb

import (
	"bytes"
	"errors"
	"strings"

	"github.com/akrylysov/pogreb"

	"git.tcp.direct/tcp.direct/database/kv"
)

// Search will search for a given string within all values inside of a Store.
// Note, type casting will be necessary. (e.g: []byte or string)
func (pstore *Store) Search(query string) (<-chan *kv.KeyValue, chan error) {
	var errChan = make(chan error, 5)
	var resChan = make(chan *kv.KeyValue, 5)
	go func() {
		defer func() {
			close(resChan)
			close(errChan)
		}()
		for _, key := range pstore.Keys() {
			if len(key) == 0 {
				continue
			}
			raw, err := pstore.Get(key)
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
func (pstore *Store) ValueExists(value []byte) (key []byte, ok bool) {
	var raw []byte
	var needle = kv.NewValue(value)
	for _, key = range pstore.Keys() {
		raw, _ = pstore.Get(key)
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
// error channel will block, so be sure to read from it.
func (pstore *Store) PrefixScan(prefixs string) (<-chan *kv.KeyValue, chan error) {
	prefix := []byte(prefixs)
	errChan := make(chan error)
	resChan := make(chan *kv.KeyValue, 5)
	go func() {
		var err error
		defer func(e error) {
			close(resChan)
			close(errChan)
		}(err)
		iter := pstore.DB.Items()
		for k, v, iterErr := iter.Next(); k != nil; k, v, iterErr = iter.Next() {
			if errors.Is(iterErr, pogreb.ErrIterationDone) {
				break
			}
			if iterErr != nil {
				errChan <- iterErr
				continue
			}
			if bytes.HasPrefix(k, prefix) {
				resChan <- kv.NewKeyValue(kv.NewKey(k), kv.NewValue(v))
			}
		}
	}()
	return resChan, errChan
}
