package bitcask

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"

	c "git.tcp.direct/kayos/common/entropy"
	"github.com/davecgh/go-spew/spew"

	"git.tcp.direct/tcp.direct/database"
)

func newTestDB(t *testing.T) database.Keeper {
	t.Helper()
	tpath := t.TempDir()
	tdb := OpenDB(tpath)
	if tdb == nil {
		t.Fatalf("failed to open testdb at %s, got nil", tpath)
	}
	return tdb
}

func seedRandKV(db database.Keeper, store string) error {
	return db.With(store).Put([]byte(c.RandStr(55)), []byte(c.RandStr(55)))
}

func seedRandStores(db database.Keeper, t *testing.T) {
	t.Helper()
	for n := 0; n != 5; n++ {
		randstore := c.RandStr(5)
		err := db.Init(randstore)
		if err != nil {
			t.Errorf("failed to initialize store for test SyncAndCloseAll: %e", err)
		}
		err = seedRandKV(db, randstore)
		if err != nil {
			t.Errorf("failed to initialize random values in store %s for test SyncAndCloseAll: %e", randstore, err)
		}
	}
	t.Logf("seeded random stores with random values for test %s", t.Name())
}

func TestDB_Init(t *testing.T) { //nolint:funlen,gocognit,cyclop
	var db = newTestDB(t)
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
			t.Fatalf("[FAIL] %e", err)
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
			t.Fatalf("[FAIL] %e", err)
		}
		err := nope.Put([]byte("bing"), []byte("bong"))
		if err != nil {
			t.Fatalf("[FAIL] %e", err)
		}
		yup := db.WithNew("bing")
		res, err := yup.Get([]byte("bing"))
		if err != nil {
			t.Errorf("[FAIL] %e", err)
		}
		if !bytes.Equal(res, []byte("bong")) {
			t.Errorf("[FAIL] wanted %v, got %v", string([]byte("bong")), string(res))
		}
	})
	t.Run("withNewStoreDoesntExist", func(t *testing.T) {
		if nope := db.WithNew("asdfqwerty"); nope.Backend() == nil {
			t.Fatalf("[FAIL] got nil result for nonexistent store when it should have made itself: %T, %v", nope, nope)
		} else {
			t.Logf("[SUCCESS] got nil Value for store that doesn't exist")
		}
	})
	t.Run("withStoreDoesntExist", func(t *testing.T) {
		nope := db.With(c.RandStr(10))
		if nope != nil {
			t.Fatalf("[FAIL] got non nil result for nonexistent store: %T, %v", nope.Backend(), nope.Backend())
		} else {
			t.Logf("[SUCCESS] got nil Value for store that doesn't exist")
		}
	})
	t.Run("syncAllShouldFail", func(t *testing.T) {
		db.(*DB).store["wtf"] = Store{}
		t.Cleanup(func() {
			t.Logf("deleting bogus store map entry")
			delete(db.(*DB).store, "wtf")
		})
		err := db.SyncAll()
		if err == nil {
			t.Fatalf("[FAIL] we should have gotten an error from bogus store map entry")
		}
		t.Logf("[SUCCESS] got compound error: %e", err)
	})

	// TODO: make sure sync is ACTUALLY sycing instead of only checking for nil err... ( ._. )

	t.Run("syncAll", func(t *testing.T) {
		err := db.SyncAll()
		if err != nil {
			t.Fatalf("[FAIL] got compound error: %e", err)
		}
	})
	t.Run("closeAll", func(t *testing.T) {
		t.Cleanup(func() {
			err := os.RemoveAll("./testdata")
			if err != nil {
				t.Fatalf("[CLEANUP FAIL] %e", err)
			}
			t.Logf("[CLEANUP] cleaned up ./testdata")
		})
		err := db.CloseAll()
		if err != nil {
			t.Fatalf("[FAIL] got compound error: %e", err)
		}
		db = nil
	})
	t.Run("SyncAndCloseAll", func(t *testing.T) {
		db = newTestDB(t)
		seedRandStores(db, t)
		err := db.SyncAndCloseAll()
		if err != nil {
			t.Errorf("[FAIL] failed to SyncAndCloseAll: %e", err)
		}
	})
}

func Test_Sync(t *testing.T) {
	// TODO: make sure sync is ACTUALLY sycing instead of only checking for nil err...
	var db = newTestDB(t)
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
	var db = newTestDB(t)
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
			t.Errorf("[FAIL] got err %e, wanted err %e", err, ErrBogusStore)
		}
	})
}

func Test_withAll(t *testing.T) {
	var db = newTestDB(t)
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
			t.Errorf("[FAIL] got err %e, wanted err %e", err, ErrNoStores)
		}
	})
	t.Run("withAllNilMap", func(t *testing.T) {
		nilDb := newTestDB(t)
		nilDb.(*DB).store = nil
		err := nilDb.(*DB).withAll(dclose)
		if err == nil {
			t.Errorf("[FAIL] got nil err from trying to work on nil map, wanted err")
		}
	})
	t.Run("withAllBogusAction", func(t *testing.T) {
		err := db.Init(asdf1)
		if err != nil {
			t.Errorf("[FAIL] unexpected error: %e", err)
		}
		wAllErr := db.(*DB).withAll(121)
		if !errors.Is(wAllErr, ErrUnknownAction) {
			t.Errorf("[FAIL] wanted error %e, got error %e", ErrUnknownAction, err)
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
			t.Errorf("[FAIL] unexpected error: %e", err)
		}
		err = db.With(asdf1).Put([]byte("asdf"), []byte("asdf"))
		if err != nil {
			t.Errorf("[FAIL] unexpected error: %e", err)
		}
		err = db.With(asdf2).Put([]byte("asdf"), []byte("asdf"))
		if err != nil {
			t.Errorf("[FAIL] unexpected error: %e", err)
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
		db.(*DB).store["yeeterson"] = Store{}
		err := db.(*DB).withAll(dclose)
		if err == nil {
			t.Errorf("[FAIL] got nil err, wanted any error")
		}
		delete(db.(*DB).store, "yeeterson")
	})
}

func Test_WithOptions(t *testing.T) { //nolint:funlen,gocognit,cyclop
	tpath := t.TempDir()
	tdb := OpenDB(tpath)
	if tdb == nil {
		t.Fatalf("failed to open testdb at %s, got nil", tpath)
	}
	defer func() {
		err := tdb.CloseAll()
		if err != nil {
			t.Fatalf("[FAIL] failed to close testdb: %e", err)
		}
	}()
	t.Run("WithMaxKeySize", func(t *testing.T) {
		err := tdb.Init(t.Name(), WithMaxKeySize(10))
		if err != nil {
			t.Fatalf("[FAIL] failed to init testdb for %s: %e", t.Name(), err)
		}
		err = tdb.With(t.Name()).Put([]byte(c.RandStr(10)), []byte("asdf"))
		if err != nil {
			t.Errorf("[FAIL] failed to put key: %e", err)
		}
		err = tdb.With(t.Name()).Put([]byte(c.RandStr(11)), []byte("asdf"))
		if err == nil {
			t.Errorf("[FAIL] expected error while using a key larger than the max key value option, got nil")
		}
	})
	t.Run("WithMaxValueSize", func(t *testing.T) {
		err := tdb.Init(t.Name(), WithMaxValueSize(10))
		if err != nil {
			t.Fatalf("[FAIL] failed to init testdb for %s: %e", t.Name(), err)
		}
		err = tdb.With(t.Name()).Put([]byte("asdf"), []byte(c.RandStr(10)))
		if err != nil {
			t.Errorf("[FAIL] failed to put key: %e", err)
		}
		err = tdb.With(t.Name()).Put([]byte("asdf"), []byte(c.RandStr(11)))
		if err == nil {
			t.Errorf("[FAIL] expected error while using a value larger than the max key value option, got nil")
		}
	})
	t.Run("WithMaxDataFileSize", func(t *testing.T) {
		err := tdb.Init(t.Name(), WithMaxDatafileSize(10))
		if err != nil {
			t.Fatalf("[FAIL] failed to init testdb for %s: %e", t.Name(), err)
		}
		checkDir := func() int {
			targetDir := tpath + "/" + t.Name()
			var files []os.DirEntry
			files, err = os.ReadDir(targetDir)
			if err != nil {
				t.Fatalf("[FAIL] failed to read directory %s: %e", targetDir, err)
			}
			datafilecount := 0
			for _, file := range files {
				if strings.Contains(file.Name(), ".data") {
					datafilecount++
				}
			}
			return datafilecount
		}
		err = tdb.With(t.Name()).Put([]byte("asdf"), []byte(c.RandStr(8)))
		if err != nil {
			t.Fatalf("[FAIL] failed to put key: %e", err)
		}
		if checkDir() != 1 {
			t.Errorf("[FAIL] expected 1 datafile, got %d", checkDir())
		}
		err = tdb.With(t.Name()).Put([]byte("asdf"), []byte(c.RandStr(10)))
		if err != nil {
			t.Fatalf("[FAIL] failed to put key: %e", err)
		}
		if checkDir() != 2 {
			t.Errorf("[FAIL] expected 2 datafile, got %d", checkDir())
		}
	})
	t.Run("SetDefaultBitcaskOptions", func(t *testing.T) {
		SetDefaultBitcaskOptions(
			WithMaxKeySize(20),
			WithMaxValueSize(20),
			WithMaxDatafileSize(20),
		)
		err := tdb.Init(t.Name())
		if err != nil {
			t.Fatalf("[FAIL] failed to init testdb for %s: %e", t.Name(), err)
		}
		checkDir := func() int {
			targetDir := tpath + "/" + t.Name()
			var files []os.DirEntry
			files, err = os.ReadDir(targetDir)
			if err != nil {
				t.Fatalf("[FAIL] failed to read directory %s: %e", targetDir, err)
			}
			datafilecount := 0
			for _, file := range files {
				if strings.Contains(file.Name(), ".data") {
					datafilecount++
				}
			}
			return datafilecount
		}
		err = tdb.With(t.Name()).Put([]byte(c.RandStr(20)), []byte("asdf"))
		if err != nil {
			t.Errorf("[FAIL] failed to put key: %e", err)
		}
		err = tdb.With(t.Name()).Put([]byte(c.RandStr(21)), []byte("asdf"))
		if err == nil {
			t.Errorf("[FAIL] expected error while using a key larger than the max key value option, got nil")
		}
		//
		err = tdb.With(t.Name()).Put([]byte("asdf"), []byte(c.RandStr(9)))
		if err != nil {
			t.Errorf("[FAIL] failed to put key: %e", err)
		}
		err = tdb.With(t.Name()).Put([]byte("asdf"), []byte(c.RandStr(21)))
		if err == nil {
			t.Errorf("[FAIL] expected error while using a value larger than the max key value option, got nil")
		}
		//
		if checkDir() != 2 {
			t.Fatalf("[FAIL] expected 2 datafiles, got %d", checkDir())
		}
		//
		err = tdb.With(t.Name()).Put([]byte("asdf"), []byte(c.RandStr(11)))
		if err != nil {
			t.Fatalf("[FAIL] failed to put key: %e", err)
		}
		if checkDir() != 3 {
			t.Fatalf("[FAIL] expected 3 datafiles, got %d", checkDir())
		}
		err = tdb.With(t.Name()).Put([]byte("asdf"), []byte(c.RandStr(10)))
		if err != nil {
			t.Fatalf("[FAIL] failed to put key: %e", err)
		}
		if checkDir() != 4 {
			t.Errorf("[FAIL] expected 4 datafile, got %d", checkDir())
		}
	})
	t.Run("InitWithBogusOption", func(t *testing.T) {
		db := newTestDB(t)
		err := db.Init("bogus", "yeet")
		if err == nil {
			t.Errorf("[FAIL] Init should have failed with bogus option")
		}
	})
}
func Test_PhonyInit(t *testing.T) {
	newtmp := t.TempDir()
	err := os.MkdirAll(newtmp+"/"+t.Name(), 0755)
	if err != nil {
		t.Fatalf("[FAIL] failed to create test directory: %e", err)
	}
	err = os.Symlink("/dev/null", newtmp+"/"+t.Name()+"/config.json")
	if err != nil {
		t.Fatal(err.Error())
	}
	tdb := OpenDB(newtmp)
	defer func() {
		_ = tdb.CloseAll()
	}()
	err = tdb.Init(t.Name())
	if err == nil {
		t.Error("[FAIL] expected error while trying to open a store where a config file exists, got nil")
	}
}
