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

func TestDB_Init(t *testing.T) {
	var db = newTestDB(t)

	type args struct {
		bucketName string
	}
	type test struct {
		name    string
		fields  *DB
		args    args
		wantErr bool
	}

	tests := []test{
		{
			name:    "simple",
			fields:  db,
			args:    args{"simple"},
			wantErr: false,
		},
		{
			name:    "bucketExists",
			fields:  db,
			args:    args{"simple"},
			wantErr: true,
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
			if err := db.Init(tt.args.bucketName); (err != nil) != tt.wantErr {
				t.Errorf("[FAIL] Init() error = %v, wantErr %v", err, tt.wantErr)
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
			t.Errorf("[FAIL] got non nil result for nonexistent bucket: %T, %v", nope, nope)
		}
		t.Logf("[SUCCESS] got nil Value for bucket that doesn't exist")
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

	t.Run("sync", func(t *testing.T) {
		for d := range db.store {
			err := db.With(d).Sync()
			if err != nil {
				t.Errorf("failed to sync %s: %e", d, err)
			}
		}
	})

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
			t.Logf("cleaned up ./testdata")
		})
		err := db.CloseAll()
		if err != nil {
			t.Fatalf("[FAIL] got compound error: %e", err)
		}
		db = nil
	})

	t.Run("SyncAndCloseAll", func(t *testing.T) {
		db = newTestDB(t)
		for n := 0; n != 5; n++ {
			err := db.Init(c.RandStr(5))
			if err != nil {
				t.Errorf("failed to initialize store for test SyncAndCloseAll: %e", err)
			}
		}
		err := db.SyncAndCloseAll()
		if err != nil {
			t.Errorf("failed to SyncAndCloseAll: %e", err)
		}
	})
}
