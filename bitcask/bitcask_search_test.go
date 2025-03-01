package bitcask

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	c "git.tcp.direct/kayos/common/entropy"
	"github.com/davecgh/go-spew/spew"

	"github.com/tcp-direct/database"
	"github.com/tcp-direct/database/kv"
)

var needle = "yeet"

type Foo struct {
	Bar  string
	Yeet int
	What map[string]int
}

func setupTest(storename string, t *testing.T) *DB {
	var db *DB
	testpath := t.TempDir()

	t.Logf("opening database at %s", testpath)
	db = OpenDB(testpath)

	err := db.Init(storename)
	if err != nil {
		t.Fatalf("[FAIL] couuldn't initialize store: %e", err)
	}

	t.Cleanup(func() {
		t.Logf("cleaning up file at; %s", testpath)
		if err := os.RemoveAll(testpath); err != nil {
			t.Error(err)
		}
	})
	return db
}

func genJunk(t *testing.T, correct bool) []byte {
	item := c.RandStr(5)
	bar := c.RandStr(5)
	if correct {
		if c.RNG(100) > 50 {
			item = needle + c.RandStr(c.RNG(5))
		} else {
			bar = c.RandStr(c.RNG(5)) + needle
		}
	}

	f := Foo{
		Bar:  bar,
		Yeet: c.RNG(50),
		What: map[string]int{c.RandStr(5): 2, item: 7},
	}

	raw, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err.Error())
	}
	return raw
}

func addJunk(db *DB, storename string, one, two, three, four, five int, t *testing.T, echo bool) [][]byte {
	var needles [][]byte
	for n := 0; n != 100; n++ {
		var rawjson []byte
		switch n {
		case one, two, three, four, five:
			rawjson = genJunk(t, true)
			needles = append(needles, rawjson)
		default:
			rawjson = genJunk(t, false)
		}
		err := db.With(storename).Put([]byte(fmt.Sprintf("%d", n)), rawjson)
		if err != nil {
			t.Fail()
			t.Logf("%e", err)
		}
	}
	if echo {
		t.Logf(
			"created 100 entries of random data with needles located at %d, %d, %d, %d, %d",
			one, two, three, four, five,
		)
	} else {
		t.Log("created 100 entries of junk")
	}

	return needles
}

func Test_Search(t *testing.T) {
	var storename = "test_search"
	var db = setupTest(storename, t)

	one := c.RNG(100)
	two := c.RNG(100)
	three := c.RNG(100)
	four := c.RNG(100)
	five := c.RNG(100)

	addJunk(db, storename, one, two, three, four, five, t, true)

	// For coverage
	db.store["yeet"] = &Store{Bitcask: nil}
	t.Run("BasicSearch", func(t *testing.T) {
		t.Logf("executing search for %s", needle)
		resChan, errChan := db.With(storename).(database.Store).Search(needle)
		var keys = []int{one, two, three, four, five}
		var needed = len(keys)

		for keyValue := range resChan {
			keyint, err := strconv.Atoi(keyValue.Key.String())
			for _, k := range keys {
				if keyint == k {
					needed--
				}
			}
			keys = append(keys, keyint)
			t.Logf("Found Key: %s, Value: %s", keyValue.Key.String(), keyValue.Value.String())

			if err != nil {
				t.Fatalf("failed to convert Key to int: %e", err)
			}
			select {
			case err := <-errChan:
				if err != nil {
					t.Fatalf("failed to search: %e", err)
				}
			default:
				continue
			}
		}
		if needed != 0 {
			t.Errorf("Needed %d results, got %d", len(keys), len(keys)-needed)
		}
	})

	t.Run("NoResultsSearch", func(t *testing.T) {
		bogus := c.RandStr(55)
		t.Logf("executing search for %s", bogus)
		var results []kv.KeyValue
		resChan, errChan := db.With(storename).(database.Store).Search(bogus)
		select {
		case err := <-errChan:
			t.Errorf("failed to search: %s", err.Error())
		case r := <-resChan:
			if r.Key.String() != "" && r.Value.String() != "" {
				spew.Dump(r)
				results = append(results, r)
			}
			if len(results) > 0 {
				t.Errorf("[FAIL] got %d results, wanted 0", len(results))
			}
		}
	})
}

func Test_ValueExists(t *testing.T) {
	var storename = "test_value_exists"
	var db = setupTest(storename, t)

	t.Run("ValueExists", func(t *testing.T) {
		needles := addJunk(db, storename, c.RNG(100), c.RNG(100), c.RNG(100), c.RNG(100), c.RNG(100), t, true)

		for _, ndl := range needles {
			k, exists := db.With(storename).(database.Store).ValueExists(ndl)
			if !exists {
				t.Fatalf("[FAIL] store should have contained a value %s somewhere, it did not.", string(ndl))
			}
			if k == nil {
				t.Fatalf("[FAIL] store should have contained a value %s somewhere, "+
					"it said it did but key was nil", string(ndl))
			}
			v, _ := db.With(storename).Get(k)
			if string(v) != string(ndl) {
				t.Fatalf("[FAIL] retrieved value does not match search target %s != %s", string(v), string(ndl))
			}
			t.Logf("[SUCCESS] successfully located value: %s, at key: %s", string(k), string(v))
		}
	})

	t.Run("ValueShouldNotExist", func(t *testing.T) {
		for n := 0; n != 5; n++ {
			garbage := c.RandStr(55)
			if _, exists := db.With(storename).(database.Store).ValueExists([]byte(garbage)); exists {
				t.Errorf("[FAIL] store should have not contained value %v, but it did", []byte(garbage))
			} else {
				t.Logf("[SUCCESS] store succeeded in not having random value %s", garbage)
			}
		}
	})

	t.Run("ValueExistNilBitcask", func(t *testing.T) {
		db.store["asdb"] = &Store{Bitcask: nil}
		garbage := "yeet"
		if _, exists := db.With(storename).(database.Store).ValueExists([]byte(garbage)); exists {
			t.Errorf("[FAIL] store should have not contained value %v, should have been nil", []byte(garbage))
		} else {
			t.Log("[SUCCESS] store succeeded in being nil")
		}
	})
}

func Test_PrefixScan(t *testing.T) {
	var storename = "test_prefix_scan"
	var db = setupTest(storename, t)
	addJunk(db, storename, c.RNG(5), c.RNG(5), c.RNG(5), c.RNG(5), c.RNG(5), t, false)
	var needles = []kv.KeyValue{
		kv.NewKeyValue(kv.NewKey([]byte("user:Frickhole")), kv.NewValue([]byte(c.RandStr(55)))),
		kv.NewKeyValue(kv.NewKey([]byte("user:Johnson")), kv.NewValue([]byte(c.RandStr(55)))),
		kv.NewKeyValue(kv.NewKey([]byte("user:Jackson")), kv.NewValue([]byte(c.RandStr(55)))),
		kv.NewKeyValue(kv.NewKey([]byte("user:Frackhole")), kv.NewValue([]byte(c.RandStr(55)))),
		kv.NewKeyValue(kv.NewKey([]byte("user:Baboshka")), kv.NewValue([]byte(c.RandStr(55)))),
	}
	for _, combo := range needles {
		err := db.With(storename).Put(combo.Key.Bytes(), combo.Value.Bytes())
		if err != nil {
			t.Fatalf("failed to add data to %s: %e", storename, err)
		} else {
			t.Logf("added needle with key(value): %s(%s)", combo.Key.String(), combo.Value.String())
		}
	}
	resChan, errChan := db.With(storename).(database.Store).PrefixScan("user:")
	var results []kv.KeyValue
	for keyValue := range resChan {
		results = append(results, keyValue)
		select {
		case err := <-errChan:
			if err != nil {
				t.Fatalf("failed to PrefixScan: %e", err)
			}
			break
		default:
			continue
		}
	}
	if len(results) != len(needles) {
		t.Errorf("[FAIL] Length of results (%d) is not the amount of needles we generated (%d)", len(results), len(needles))
	}
	var keysmatched = 0
	for _, result := range results {
		for _, ogkv := range needles {
			if result.Key.String() != ogkv.Key.String() {
				continue
			}
			t.Logf("Found needle key: %s", result.Key.String())
			keysmatched++
			if result.Value.String() != ogkv.Value.String() {
				t.Errorf("[FAIL] values of key %s should have matched. wanted: %s, got: %s",
					result.Key.String(), ogkv.Value.String(), result.Value.String())
			}
			t.Logf("Found needle value: %s", ogkv.Value.String())
		}
	}
	if keysmatched != len(needles) {
		t.Errorf("Needed to match %d keys, only matched %d", len(needles), len(needles))
	}
}
