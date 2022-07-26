package bitcask

import (
	"testing"

	"git.tcp.direct/tcp.direct/database"
)

func needKeeper(keeper database.Keeper) {}
func needFiler(filer database.Filer)    {}

func Test_Keeper(t *testing.T) {
	needKeeper(OpenDB(""))
}

func Test_Filer(t *testing.T) {
	needFiler(OpenDB("").With(""))
}
