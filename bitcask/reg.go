package bitcask

import (
	"errors"

	"git.mills.io/prologic/bitcask"

	"git.tcp.direct/tcp.direct/database"
	"git.tcp.direct/tcp.direct/database/registry"
)

var ErrBadOpt = errors.New("invalid bitcask options")

func init() {
	registry.RegisterKeeper("bitcask", func(path string, opt ...any) (database.Keeper, error) {
		bitcaskOptions := make([]bitcask.Option, 0, len(opt))
		for _, o := range opt {
			var casted bitcask.Option
			switch v := o.(type) {
			case bitcask.Option:
				casted = v
			case *bitcask.Option:
				casted = *v
			case []bitcask.Option:
				bitcaskOptions = append(bitcaskOptions, v...)
				continue
			default:
				return nil, ErrBadOpt
			}
			bitcaskOptions = append(bitcaskOptions, casted)
		}
		if len(bitcaskOptions) > 0 {
			defaultBitcaskOptions = append(defaultBitcaskOptions, bitcaskOptions...)
		}
		db := OpenDB(path)
		err := db.init()
		return db, err
	})
}
