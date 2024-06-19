package bitcask

import (
	"git.tcp.direct/tcp.direct/database"
	"git.tcp.direct/tcp.direct/database/registry"
)

func init() {
	registry.RegisterKeeper("bitcask", func(path string) (database.Keeper, error) {
		db := OpenDB(path)
		err := db.init()
		return db, err
	})
}
