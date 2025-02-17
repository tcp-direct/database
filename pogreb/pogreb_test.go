package pogreb

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	c "git.tcp.direct/kayos/common/entropy"
	"github.com/davecgh/go-spew/spew"

	"github.com/tcp-direct/database"
)

func newTestDB(t *testing.T) (string, database.Keeper) {
	t.Helper()
	tpath := t.TempDir()
	t.Cleanup(func() {
		t.Logf("[CLEANUP] removing temp dir %s", tpath)
		err := os.RemoveAll(tpath)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			panic(err.Error())
		}
	})
	tdb := OpenDB(tpath)
	if tdb == nil {
		t.Fatalf("failed to open testdb at %s, got nil", tpath)
	}
	return tdb.Path(), tdb
}

func seedRandKV(db database.Keeper, store string) error {
	return db.With(store).Put([]byte(c.RandStr(55)), []byte(c.RandStr(55)))
}

func seedRandStores(db database.Keeper, t *testing.T) []string {
	t.Helper()
	names := make([]string, 0, 5)
	for n := 0; n != 5; n++ {
		randstore := c.RandStr(5)
		err := db.Init(randstore)
		if err != nil {
			t.Errorf("failed to initialize store for test SyncAndCloseAll: %s", err.Error())
		}
		err = seedRandKV(db, randstore)
		if err != nil {
			t.Errorf("failed to initialize random values in store %s for test SyncAndCloseAll: %e", randstore, err)
		}
		names = append(names, randstore)
	}
	t.Logf("seeded random stores with random values for test %s", t.Name())
	return names
}

func TestDB_Init(t *testing.T) { //nolint:funlen,gocognit,cyclop
	var _, db = newTestDB(t)
	type args struct{ storeName string }
	type test struct {
		name    string
		args    args
		wantErr bool
		specErr error
	}
	tests := []test{
		{
			name:    "simple",
			args:    args{"simple"},
			wantErr: false,
		},
		{
			name:    "storeExists",
			args:    args{"simple"},
			wantErr: true,
			specErr: ErrStoreExists,
		},
		{
			name:    "newStore",
			args:    args{"notsimple"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.Init(tt.args.storeName)
			if (err != nil) != tt.wantErr {
				t.Errorf("[FAIL] Init() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (err != nil) != tt.wantErr && tt.specErr != nil && !errors.Is(err, tt.specErr) {
				t.Errorf("[FAIL] wanted error %e, got error %e", tt.specErr, err)
			}
		})
	}

	t.Run("withStoreTest", func(t *testing.T) {
		key := []byte{51, 50}
		value := []byte("string")
		err := db.With("simple").Put(key, value)
		t.Logf("Put Value %v at Key %v", string(value), key)
		if err != nil {
			t.Fatalf("[FAIL] %s", err.Error())
		}
		gvalue, gerr := db.With("simple").Get(key)
		if gerr != nil {
			t.Fatalf("[FAIL] %e", gerr)
		}
		if !bytes.Equal(gvalue, value) {
			t.Errorf("[FAIL] wanted %v, got %v", string(value), string(gvalue))
		}
		t.Logf("Got Value %v at Key %v", string(gvalue), key)
	})
	t.Run("withNewStoreDoesExist", func(t *testing.T) {
		nope := db.WithNew("bing")
		if err := nope.Put([]byte("key"), []byte("value")); err != nil {
			t.Fatalf("[FAIL] %s", err.Error())
		}
		err := nope.Put([]byte("bing"), []byte("bong"))
		if err != nil {
			t.Fatalf("[FAIL] %s", err.Error())
		}
		yup := db.WithNew("bing")
		res, err := yup.Get([]byte("bing"))
		if err != nil {
			t.Errorf("[FAIL] %s", err.Error())
		}
		if !bytes.Equal(res, []byte("bong")) {
			t.Errorf("[FAIL] wanted %v, got %v", string("bong"), string(res))
		}
	})
	t.Run("withNewStoreDoesntExist", func(t *testing.T) {
		if nope := db.WithNew("asdfqwerty"); nope.Backend() == nil || nope.Backend() == nilBackend {
			t.Fatalf("[FAIL] got nil result for nonexistent store when it should have made itself: %T, %v", nope, nope)
		} else {
			t.Logf("[SUCCESS] got new store with valid backend when calling WithNew for store that doesn't exist")
		}
	})
	t.Run("withStoreDoesntExist", func(t *testing.T) {
		nope := db.With(c.RandStr(10))
		if nope != nil && nope.Backend() != nilBackend {
			t.Fatalf("[FAIL] got non nil result for nonexistent store: %T, %v", nope.Backend(), nope.Backend())
		} else {
			t.Logf("[SUCCESS] got nil Value for store that doesn't exist")
		}
	})
	t.Run("syncAllShouldFail", func(t *testing.T) {
		db.(*DB).store["wtf"] = &Store{}
		t.Cleanup(func() {
			t.Logf("deleting bogus store map entry")
			delete(db.(*DB).store, "wtf")
		})
		err := db.SyncAll()
		if err == nil {
			t.Fatalf("[FAIL] we should have gotten an error from bogus store map entry")
		}
		t.Logf("[SUCCESS] got compound error: %s", err.Error())
	})

	// TODO: make sure sync is ACTUALLY sycing instead of only checking for nil err... ( ._. )

	t.Run("syncAll", func(t *testing.T) {
		err := db.SyncAll()
		if err != nil {
			t.Fatalf("[FAIL] got compound error: %s", err.Error())
		}
	})
	t.Run("closeAll", func(t *testing.T) {
		t.Cleanup(func() {
			err := os.RemoveAll("./testdata")
			if err != nil {
				t.Fatalf("[CLEANUP FAIL] %s", err.Error())
			}
			t.Logf("[CLEANUP] cleaned up ./testdata")
		})
		err := db.CloseAll()
		if err != nil {
			t.Fatalf("[FAIL] got compound error: %s", err.Error())
		}
		db = nil
	})
	t.Run("SyncAndCloseAll", func(t *testing.T) {
		var tdbp string
		tdbp, db = newTestDB(t)
		names := seedRandStores(db, t)
		err := db.SyncAndCloseAll()
		if err != nil {
			t.Fatalf("[FAIL] failed to SyncAndCloseAll: %s", err.Error())
		}
		db = OpenDB(tdbp)
		found, err := db.(*DB).Discover()
		if err != nil {
			t.Fatalf("[FAIL] failed to discover stores: %s", err.Error())
		}
		if len(found) == 0 {
			t.Fatalf("[FAIL] found no stores")
		}
		for _, n := range names {
			matched := false
			for _, f := range found {
				if f == n {
					matched = true
					break
				}
			}
			if !matched {
				t.Errorf("[FAIL] failed to find store %s", n)
			}
		}
	})
}

func Test_Sync(t *testing.T) {
	// TODO: make sure sync is ACTUALLY sycing instead of only checking for nil err...
	var _, db = newTestDB(t)
	seedRandStores(db, t)
	t.Run("Sync", func(t *testing.T) {
		for d := range db.(*DB).store {
			err := db.With(d).Sync()
			if err != nil {
				t.Errorf("[FAIL] failed to sync %s: %e", d, err)
			} else {
				t.Logf("[+] Sync() successful for %s", d)
			}
		}
	})
}

func Test_Close(t *testing.T) {
	var _, db = newTestDB(t)
	defer func() {
		db = nil
	}()
	seedRandStores(db, t)
	var oldstores []string
	t.Run("Close", func(t *testing.T) {
		for d := range db.AllStores() {
			oldstores = append(oldstores, d)
			err := db.Close(d)
			if err != nil {
				t.Fatalf("[FAIL] failed to close %s: %v", d, err)
			}
			t.Logf("[+] Close() successful for %s", d)
		}
		t.Run("AssureClosed", func(t *testing.T) {
			for _, d := range oldstores {
				if st := db.With(d); st != nil {
					spew.Dump(st)
					t.Errorf("[FAIL] store %s should have been deleted", d)
				}
			}
			t.Logf("[SUCCESS] Confirmed that all stores have been closed")
		})
	})

	t.Run("CantCloseBogusStore", func(t *testing.T) {
		err := db.Close(c.RandStr(55))
		if !errors.Is(err, ErrBogusStore) {
			t.Errorf("[FAIL] got err %s, wanted err %s", err.Error(), ErrBogusStore)
		}
	})
}

func Test_withAll(t *testing.T) {
	var _, db = newTestDB(t)
	asdf1 := c.RandStr(10)
	asdf2 := c.RandStr(10)

	defer func() {
		if err := db.CloseAll(); err != nil && !errors.Is(err, fs.ErrClosed) {
			t.Errorf("[FAIL] failed to close all stores: %v", err)
		}
	}()
	t.Run("withAllNoStores", func(t *testing.T) {
		err := db.(*DB).withAll(121)
		if !errors.Is(err, ErrNoStores) {
			t.Errorf("[FAIL] got err %s, wanted err %s", err.Error(), ErrNoStores)
		}
	})
	t.Run("withAllNilMap", func(t *testing.T) {
		_, nilDb := newTestDB(t)
		nilDb.(*DB).store = nil
		err := nilDb.(*DB).withAll(dclose)
		if err == nil {
			t.Errorf("[FAIL] got nil err from trying to work on nil map, wanted err")
		}
	})
	t.Run("withAllBogusAction", func(t *testing.T) {
		err := db.Init(asdf1)
		if err != nil {
			t.Errorf("[FAIL] unexpected error: %s", err.Error())
		}
		wAllErr := db.(*DB).withAll(121)
		if !errors.Is(wAllErr, ErrUnknownAction) {
			t.Errorf("[FAIL] wanted error %s, got error %s", ErrUnknownAction.Error(), err)
		}
	})
	t.Run("ListAll", func(t *testing.T) {
		allStores := db.AllStores()
		if len(allStores) == 0 {
			t.Errorf("[FAIL] no stores found")
		}
		for n, s := range allStores {
			if n == "" {
				t.Errorf("[FAIL] store name is empty")
			}
			if s == nil {
				t.Errorf("[FAIL] store is nil")
			}
			t.Logf("[+] found store named %s: %v", n, s)
		}
		if len(allStores) != len(db.(*DB).store) {
			t.Errorf("[FAIL] found %d stores, expected %d", len(allStores), len(db.(*DB).store))
		}
	})
	t.Run("ListAllAndInteract", func(t *testing.T) {
		err := db.Init(asdf2)
		if err != nil {
			t.Errorf("[FAIL] unexpected error: %s", err.Error())
		}
		err = db.With(asdf1).Put([]byte("asdf"), []byte("asdf"))
		if err != nil {
			t.Errorf("[FAIL] unexpected error: %s", err.Error())
		}
		err = db.With(asdf2).Put([]byte("asdf"), []byte("asdf"))
		if err != nil {
			t.Errorf("[FAIL] unexpected error: %s", err.Error())
		}
		allStores := db.AllStores()
		if len(allStores) == 0 {
			t.Errorf("[FAIL] no stores found")
		}
		for n, s := range allStores {
			if n == "" {
				t.Errorf("[FAIL] store name is empty")
			}
			if s == nil {
				t.Errorf("[FAIL] store is nil")
			}
			if len(db.(*DB).store) != 2 {
				t.Errorf("[SANITY FAIL] found %d stores, expected %d", len(allStores), len(db.(*DB).store))
			}
			t.Logf("[+] found store named %s: %v", n, s)
			if len(allStores) != len(db.(*DB).store) {
				t.Errorf("[FAIL] found %d stores, expected %d", len(allStores), len(db.(*DB).store))
			}
			var res []byte
			res, err = db.With(n).Get([]byte("asdf"))
			if err != nil {
				t.Errorf("[FAIL] unexpected error: %v", err)
			}
			if !bytes.Equal(res, []byte("asdf")) {
				t.Errorf("[FAIL] expected %s, got %s", n, res)
			} else {
				t.Logf("[+] found %s in store %s", res, n)
			}
		}
	})
	t.Run("WithAllIncludingBadStore", func(t *testing.T) {
		db.(*DB).store["yeeterson"] = &Store{}
		err := db.(*DB).withAll(dclose)
		if err == nil {
			t.Errorf("[FAIL] got nil err, wanted any error")
		}
		delete(db.(*DB).store, "yeeterson")
	})

	// initialize store for the defer closure call
	if err := db.Init(asdf1); err != nil {
		t.Fatalf("[FAIL] %s", err.Error())
	}

}

func Test_WithOptions(t *testing.T) { //nolint:funlen,gocognit,cyclop
	tpath := t.TempDir()
	t.Cleanup(func() {
		if err := os.RemoveAll(tpath); err != nil {
			panic(err)
		}
	})
	tdb := OpenDB(tpath)
	if tdb == nil {
		t.Fatalf("failed to open testdb at %s, got nil", tpath)
	}
	// FIXME: inconsistent with other implementations (pogreb)
	defer func() {
		t.Helper()
		err := tdb.CloseAll()
		if err == nil {
			t.Fatalf("[FAIL] was able to close uninitialized store, expected error")
		}
	}()
	t.Run("InitWithBogusOption", func(t *testing.T) {
		_, db := newTestDB(t)
		err := db.Init("bogus", "yeet")
		if err == nil {
			t.Errorf("[FAIL] Init should have failed with bogus option")
		}
	})
}
func Test_PhonyInit(t *testing.T) {
	newtmp := t.TempDir()
	t.Cleanup(func() {
		if err := os.RemoveAll(newtmp); err != nil {
			panic(err)
		}
	})
	err := os.MkdirAll(newtmp+"/"+t.Name(), 0755)
	if err != nil {
		t.Fatalf("[FAIL] failed to create test directory: %s", err.Error())
	}
	err = os.Symlink("/dev/null", filepath.Join(newtmp, t.Name(), "lock"))
	if err != nil {
		t.Fatal(err.Error())
	}
	tdb := OpenDB(newtmp)
	defer func() {
		_ = tdb.CloseAll()
	}()
	err = tdb.Init(t.Name())
	if err == nil {
		t.Error("[FAIL] expected error while trying to open a store where lock exists, got nil")
	}
}
