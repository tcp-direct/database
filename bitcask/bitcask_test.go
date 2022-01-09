package bitcask

import (
	"os"
	"testing"

	c "git.tcp.direct/kayos/common"
)

func newTestDB(t *testing.T) *DB {
	tpath := t.TempDir()
	tdb := OpenDB(tpath)
	if tdb == nil {
		t.Fatalf("failed to open testdb at %s, got nil", tpath)
	}
	return tdb
}

func seedRandKV(db *DB, store string) error {
	return db.With(store).Put([]byte(c.RandStr(55)), []byte(c.RandStr(55)))
}

func seedRandStores(db *DB, t *testing.T) {
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

func TestDB_Init(t *testing.T) {
	var db = newTestDB(t)

	type args struct {
		storeName string
	}
	type test struct {
		name    string
		fields  *DB
		args    args
		wantErr bool
		specErr error
	}

	tests := []test{
		{
			name:    "simple",
			fields:  db,
			args:    args{"simple"},
			wantErr: false,
		},
		{
			name:    "storeExists",
			fields:  db,
			args:    args{"simple"},
			wantErr: true,
			specErr: errStoreExists,
		},
		{
			name:    "newBucket",
			fields:  db,
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
			if (err != nil) != tt.wantErr && tt.specErr != nil && err != tt.specErr {
				t.Errorf("[FAIL] wanted error %e, got error %e", tt.specErr, err)
			}
		})
	}

	t.Run("withBucketTest", func(t *testing.T) {
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
		if string(gvalue) != string(value) {
			t.Errorf("[FAIL] wanted %v, got %v", string(value), string(gvalue))
		}
		t.Logf("Got Value %v at Key %v", string(gvalue), key)
	})
	t.Run("withBucketDoesntExist", func(t *testing.T) {
		if nope := db.With("asdfqwerty"); nope.Bitcask != nil {
			t.Errorf("[FAIL] got non nil result for nonexistent store: %T, %v", nope, nope)
		}
		t.Logf("[SUCCESS] got nil Value for store that doesn't exist")
	})
	t.Run("syncAllShouldFail", func(t *testing.T) {
		db.store["wtf"] = Store{}
		t.Cleanup(func() {
			t.Logf("deleting bogus store map entry")
			delete(db.store, "wtf")
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
	// TODO: make sure sync is ACTUALLY sycing instead of only checking for nil err... ( ._. )

	var db = newTestDB(t)
	seedRandStores(db, t)
	t.Run("Sync", func(t *testing.T) {
		for d := range db.store {
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
	seedRandStores(db, t)
	var oldstores []string
	t.Run("Close", func(t *testing.T) {
		for d := range db.store {
			oldstores = append(oldstores, d)
			err := db.Close(d)
			if err != nil {
				t.Errorf("[FAIL] failed to close %s: %e", d, err)
			} else {
				t.Logf("[+] Close() successful for %s", d)
			}
		}
	})
	t.Run("AssureClosed", func(t *testing.T) {
		for _, d := range oldstores {
			if st := db.With(d); st.Bitcask != nil {
				t.Errorf("[FAIL] store %s should have been deleted", d)
			}
		}
		t.Logf("[SUCCESS] Confirmed that all stores have been closed")
	})

	t.Run("CantCloseBogusStore", func(t *testing.T) {
		err := db.Close(c.RandStr(55))
		if err != errBogusStore {
			t.Errorf("[FAIL] got err %e, wanted err %e", err, errBogusStore)
		}
	})
}

func Test_withAll(t *testing.T) {
	var db = newTestDB(t)
	t.Run("withAllNoStores", func(t *testing.T) {
		err := db.withAll(121)
		if err != errNoStores {
			t.Errorf("[FAIL] got err %e, wanted err %e", err, errBogusStore)
		}
	})
	t.Run("withAllBogusAction", func(t *testing.T) {
		err := db.Init("asdf")
		if err != nil {
			t.Errorf("[FAIL] unexpected error: %e", err)
		}
		wAllErr := db.withAll(121)
		if wAllErr != errUnknownAction {
			t.Errorf("[FAIL] wanted error %e, got error %e", errUnknownAction, err)
		}
	})
}
