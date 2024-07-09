package loader

import (
	"fmt"
	"os"
	"path/filepath"

	"git.tcp.direct/tcp.direct/database"
	"git.tcp.direct/tcp.direct/database/metadata"
	"git.tcp.direct/tcp.direct/database/registry"
)

func OpenKeeper(path string, opts ...any) (database.Keeper, error) {
	stat, statErr := os.Stat(path)
	if statErr != nil {
		return nil, statErr
	}
	var metaDat []byte
	var readErr error
	if stat.IsDir() {
		if stat, statErr = os.Stat(filepath.Join(path, "meta.json")); statErr != nil {
			return nil, fmt.Errorf("meta.json not found in target directory: %w", os.ErrNotExist)
		}
		metaDat, readErr = os.ReadFile(filepath.Join(path, "meta.json"))
	} else {
		metaDat, readErr = os.ReadFile(path)
	}

	if readErr != nil {
		return nil, fmt.Errorf("error reading meta.json: %w", readErr)
	}
	if len(metaDat) == 0 {
		return nil, fmt.Errorf("meta.json is empty")
	}

	meta, err := metadata.LoadMeta(metaDat)
	if err != nil {
		return nil, fmt.Errorf("error parsing meta.json: %w", err)
	}
	var keeperCreator database.KeeperCreator
	if keeperCreator = registry.GetKeeper(meta.KeeperType); keeperCreator == nil {
		return nil, fmt.Errorf("keeper type %s not found in registry", meta.KeeperType)
	}
	keeper, err := keeperCreator(path, meta.DefStoreOpts)
	if err != nil {
		return nil, fmt.Errorf("error substantiating keeper: %w", err)
	}
	if _, err = keeper.Discover(); err != nil {
		return nil, fmt.Errorf("error opening existing stores: %w", err)
	}
	return keeper, nil
}
