package database

import (
	"strings"

	"git.tcp.direct/kayos/common"
)

func (c Casket) Search(query string) map[string]interface{} {
	res := make(map[string]interface{})
	for _, key := range c.AllKeys() {
		raw, _ := c.Get([]byte(key))
		if raw == nil {
			continue
		}
		if strings.Contains(string(raw), query) {
			res[key] = raw
		}
	}
	return res
}

func (c Casket) ValueExists(value []byte) (key []byte, ok bool) {
	var raw []byte
	for _, k := range c.AllKeys() {
		raw, _ = c.Get([]byte(k))
		if raw == nil {
			continue
		}
		if common.CompareChecksums(value, raw) {
			ok = true
			return
		}
	}
	return
}

func (c Casket) PrefixScan(prefix string) map[string]interface{} {
	res := make(map[string]interface{})
	c.Scan([]byte(prefix), func(key []byte) error {
		raw, err := c.Get(key)
		if  err != nil {
			return err
		}
		res[string(key)] = raw
		return nil
	})
	return res
}

func (c Casket) AllKeys() (keys []string) {
	keychan := c.Keys()
	for key := range keychan {
		keys = append(keys, string(key))
	}
	return
}
