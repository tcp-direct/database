package loader

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tcp-direct/database"
	"github.com/tcp-direct/database/metadata"
	"github.com/tcp-direct/database/registry"
)

var ErrEmptyMeta = errors.New("meta.json is empty")

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
		return nil, ErrEmptyMeta
	}

	meta, err := metadata.LoadMeta(metaDat)
	if err != nil {
		return nil, fmt.Errorf("error parsing meta.json: %w", err)
	}
	var keeperCreator database.KeeperCreator
	if keeperCreator = registry.GetKeeper(meta.KeeperType); keeperCreator == nil {
		return nil, fmt.Errorf("keeper type %s not found in registry", meta.KeeperType)
	}

	var (
		keeper database.Keeper
	)

	if len(opts) > 0 {
		keeper, err = keeperCreator(path, opts...)
	} else {
		keeper, err = keeperCreator(path)
	}

	if err != nil {
		return nil, fmt.Errorf("error substantiating keeper: %w", err)
	}
	if _, err = keeper.Discover(); err != nil {
		return nil, fmt.Errorf("error opening existing stores: %w", err)
	}
	return keeper, nil
}
