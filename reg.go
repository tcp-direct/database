package database

import (
	"sync"
)

var (
	keepers = make(map[string]KeeperCreator)
	regMu   = &sync.RWMutex{}
)

func Register(name string, keeper KeeperCreator) {
	regMu.Lock()
	keepers[name] = keeper
	regMu.Unlock()
}

func Get(name string) KeeperCreator {
	regMu.RLock()
	k := keepers[name]
	regMu.RUnlock()
	return k
}

func All() map[string]KeeperCreator {
	regMu.RLock()
	ks := keepers
	regMu.RUnlock()
	return ks
}
