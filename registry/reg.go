package registry

import (
	"sync"

	"git.tcp.direct/tcp.direct/database"
)

var (
	keeperIndex = make(map[string]database.KeeperCreator)
	regMu       = &sync.RWMutex{}
)

// RegisterKeeper registers a new [KeeperCreator];
// a function that creates a new [Keeper] implementation,
// under the given name in the global registry.
func RegisterKeeper(name string, keeper database.KeeperCreator) {
	regMu.Lock()
	keeperIndex[name] = keeper
	regMu.Unlock()
}

// GetKeeper retrieves a [KeeperCreator] from the global registry by name.
func GetKeeper(name string) database.KeeperCreator {
	regMu.RLock()
	k := keeperIndex[name]
	regMu.RUnlock()
	return k
}

// AllKeepers returns a slice of all registered [Keeper] implementation names.
func AllKeepers() []string {
	keeperNames := make([]string, len(keeperIndex))
	index := 0
	for k := range keeperIndex {
		keeperNames[index] = k
		index++
	}
	return keeperNames
}
