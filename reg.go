package database

import (
	"sync"
)

var (
	keepers = make(map[string]KeeperCreator)
	regMu   = &sync.RWMutex{}
)

func RegisterKeeper(name string, keeper KeeperCreator) {
	regMu.Lock()
	keepers[name] = keeper
	regMu.Unlock()
}

func GetKeeper(name string) KeeperCreator {
	regMu.RLock()
	k := keepers[name]
	regMu.RUnlock()
	return k
}

func AllKeepers() map[string]KeeperCreator {
	regMu.RLock()
	ks := keepers
	regMu.RUnlock()
	return ks
}
