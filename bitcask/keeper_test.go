package bitcask

import (
	"testing"

	"git.tcp.direct/tcp.direct/database"
)

func Test_Interfaces(t *testing.T) {
	v := OpenDB(t.TempDir())
	var keeper interface{} = v
	if _, ok := keeper.(database.Keeper); !ok {
		t.Error("Keeper interface not implemented")
	}
	vs := v.WithNew("test")
	var searcher interface{} = vs
	if _, ok := searcher.(database.Searcher); !ok {
		t.Error("Searcher interface not implemented")
	}
	var filer interface{} = vs
	if _, ok := filer.(database.Filer); !ok {
		t.Error("Filer interface not implemented")
	}
}
