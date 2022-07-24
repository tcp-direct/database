package bitcask

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	c "git.tcp.direct/kayos/common/entropy"
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
	db.store["yeet"] = Store{Bitcask: nil}
	t.Run("BasicSearch", func(t *testing.T) {
		t.Logf("executing search for %s", needle)

		results, err := db.With(storename).Search(needle)
		if err != nil {
			t.Errorf("failed to search: %e", err)
		}
		var keys = []int{one, two, three, four, five}
		var needed = len(keys)
		for _, kv := range results {
			keyint, err := strconv.Atoi(kv.Key.String())
			if err != nil {
				t.Fatalf("failed to convert Key to int: %e", err)
			}
			for _, k := range keys {
				if keyint == k {
					needed--
				}
			}
			keys = append(keys, keyint)
			t.Logf("Found Key: %s, Value: %s", kv.Key.String(), kv.Value.String())
		}
		if needed != 0 {
			t.Errorf("Needed %d results, got %d", len(keys), len(keys)-needed)
		}
	})

	t.Run("NoResultsSearch", func(t *testing.T) {
		bogus := c.RandStr(55)
		t.Logf("executing search for %s", bogus)

		results, err := db.With(storename).Search(bogus)
		if err != nil {
			t.Errorf("failed to search: %e", err)
		}
		if len(results) > 0 {
			t.Errorf("[FAIL] got %d results, wanted 0", len(results))
		}
	})
}

func Test_ValueExists(t *testing.T) {
	var storename = "test_value_exists"
	var db = setupTest(storename, t)

	t.Run("ValueExists", func(t *testing.T) {
		needles := addJunk(db, storename, c.RNG(100), c.RNG(100), c.RNG(100), c.RNG(100), c.RNG(100), t, true)

		for _, needle := range needles {
			if k, exists := db.With(storename).ValueExists(needle); !exists {
				t.Errorf("[FAIL] store should have contained a value %s somewhere, it did not.", string(needle))
			} else {
				t.Logf("[SUCCESS] successfully located value: %s, at key: %s", string(k), string(needle))
			}
		}
	})

	t.Run("ValueShouldNotExist", func(t *testing.T) {
		for n := 0; n != 5; n++ {
			garbage := c.RandStr(55)
			if _, exists := db.With(storename).ValueExists([]byte(garbage)); exists {
				t.Errorf("[FAIL] store should have not contained value %v, but it did", []byte(garbage))
			} else {
				t.Logf("[SUCCESS] store succeeded in not having random value %s", garbage)
			}
		}
	})

	t.Run("ValueExistNilBitcask", func(t *testing.T) {
		db.store["asdb"] = Store{Bitcask: nil}
		garbage := "yeet"
		if _, exists := db.With(storename).ValueExists([]byte(garbage)); exists {
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
	var needles = []KeyValue{
		{Key: Key{b: []byte("user:Frickhole")}, Value: Value{b: []byte(c.RandStr(55))}},
		{Key: Key{b: []byte("user:Johnson")}, Value: Value{b: []byte(c.RandStr(55))}},
		{Key: Key{b: []byte("user:Jackson")}, Value: Value{b: []byte(c.RandStr(55))}},
		{Key: Key{b: []byte("user:Frackhole")}, Value: Value{b: []byte(c.RandStr(55))}},
		{Key: Key{b: []byte("user:Baboshka")}, Value: Value{b: []byte(c.RandStr(55))}},
	}
	for _, kv := range needles {
		err := db.With(storename).Put(kv.Key.Bytes(), kv.Value.Bytes())
		if err != nil {
			t.Fatalf("failed to add data to %s: %e", storename, err)
		} else {
			t.Logf("added needle with key(value): %s(%s)", kv.Key.String(), kv.Value.String())
		}
	}
	res, err := db.With(storename).PrefixScan("user:")
	if err != nil {
		t.Errorf("failed to PrefixScan: %e", err)
	}
	if len(res) != len(needles) {
		t.Errorf("[FAIL] Length of results (%d) is not the amount of needles we generated (%d)", len(res), len(needles))
	}
	var keysmatched = 0
	for _, kv := range res {
		for _, ogkv := range needles {
			if kv.Key.String() != ogkv.Key.String() {
				continue
			}
			t.Logf("[%s] Found needle key", ogkv.Key.String())
			keysmatched++
			if kv.Value.String() != ogkv.Value.String() {
				t.Errorf("[FAIL] values of key %s should have matched. wanted: %s, got: %s", kv.Key.String(), ogkv.Value.String(), kv.Value.String())
			}
			t.Logf("[%s] Found needle value: %s", ogkv.Key.String(), ogkv.Value.String())
		}
	}
	if keysmatched != len(needles) {
		t.Errorf("Needed to match %d keys, only matched %d", len(needles), len(needles))
	}
}
