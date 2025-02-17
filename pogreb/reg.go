package pogreb

import (
	"github.com/tcp-direct/database"
	"github.com/tcp-direct/database/registry"
)

func init() {
	registry.RegisterKeeper("pogreb", func(path string, opts ...any) (database.Keeper, error) {
		if len(opts) > 1 {
			return nil, ErrInvalidOptions
		}
		if len(opts) == 1 {
			casted, castErr := castOptions(opts...)
			if castErr != nil {
				return nil, castErr
			}
			defOptMu.Lock()
			defaultPogrebOptions = casted
			defOptMu.Unlock()
		}
		db := OpenDB(path)
		err := db.init()
		return db, err
	})
}
