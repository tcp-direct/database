package pogreb

import "git.tcp.direct/tcp.direct/database"

func init() {
	database.RegisterKeeper("pogreb", func(path string) (database.Keeper, error) {
		db := OpenDB(path)
		err := db.init()
		return db, err
	})
}
