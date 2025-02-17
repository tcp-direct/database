package loader

import (
	"path/filepath"
	"testing"

	"github.com/tcp-direct/database/test"
)

func TestOpenKeeper(t *testing.T) {
	path := filepath.Join(t.TempDir(), "meta.json")

	nmk := database.NewMockKeeper("yeets1")
	if err := nmk.Init("yeets1"); err != nil {
		t.Fatalf("error initializing mock keeper: %v", err)
	}
	if err := nmk.With("yeets1").Put([]byte("yeet1"), []byte("yeet1")); err != nil {
		t.Fatalf("error putting value: %v", err)
	}
	if err := nmk.WithNew("yeets1").Put([]byte("yeet2"), []byte("yeet2")); err != nil {
		t.Fatalf("error putting value: %v", err)
	}
	val, err := nmk.With("yeets1").Get([]byte("yeet1"))
	if err != nil {
		t.Fatalf("error getting value: %v", err)
	}
	if string(val) != "yeet1" {
		t.Errorf("expected yeet1, got %s", val)
	}
	val, err = nmk.WithNew("yeets1").Get([]byte("yeet2"))
	if err != nil {
		t.Fatalf("error getting value: %v", err)
	}
	if string(val) != "yeet2" {
		t.Errorf("expected yeet2, got %s", val)
	}

	if err = nmk.WriteMeta(path); err != nil {
		t.Fatalf("error writing meta: %v", err)
	}

	keeper, err := OpenKeeper(path)
	if err != nil {
		t.Fatalf("error opening keeper: %v", err)
	}
	if keeper == nil {
		t.Fatal("expected keeper, got nil")
	}
	openedStores := keeper.AllStores()
	if openedStores == nil || len(openedStores) == 0 {
		t.Errorf("expected stores, got nil")
	}
	if _, ok := openedStores["yeets1"]; !ok {
		t.Errorf("expected store yeets1, got nil")
	}
	if val, err = keeper.With("yeets1").Get([]byte("yeet1")); err != nil {
		t.Fatalf("error getting value: %v", err)
	}
	if string(val) != "yeet1" {
		t.Errorf("expected yeet1, got %s", val)
	}
	if val, err = keeper.With("yeets1").Get([]byte("yeet2")); err != nil {
		t.Fatalf("error getting value: %v", err)
	}
	if string(val) != "yeet2" {
		t.Errorf("expected yeet2, got %s", val)
	}
}

func TestOpenKeeperWithOpts(t *testing.T) {
	nmK := database.NewMockKeeper("yeeties")
	_ = nmK.WithNew("yeets").Put([]byte("yeets"), []byte("yeets"))
	_ = nmK.WithNew("yeets1").Put([]byte("yeet1"), []byte("yeet1"))
	tmp := filepath.Join(t.TempDir(), "meta.json")
	if err := nmK.WriteMeta(tmp); err != nil {
		t.Fatalf("error writing meta: %v", err)
	}
	keeper, err := OpenKeeper(tmp, "yeeterson mcgee", "yeet it")
	if err != nil {
		t.Fatalf("error opening keeper: %v", err)
	}
	for _, store := range keeper.AllStores() {
		found := 0
		for _, opt := range store.(*database.MockFiler).Opts {
			if opt == "yeet it" {
				found++
			}
			if opt == "yeeterson mcgee" {
				found++
			}
			t.Logf("found option: %v", opt)
		}
		if found > 2 {
			t.Fatal("too many options")
		}
		if found != 2 {
			t.Fatal("not enough options")
		}
	}
}
