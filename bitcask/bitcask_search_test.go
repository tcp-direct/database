package bitcask

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	c "git.tcp.direct/kayos/common"
)

var needle = "yeet"

type Foo struct {
	Bar  string
	Yeet int
	What map[string]int
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

func Test_Search(t *testing.T) {
	var (
		testpath = "./testsearch"
	)

	var db *DB

	t.Logf("opening database at %s", testpath)

	db = OpenDB(testpath)

	err := db.Init("searchtest")
	if err != nil {
		t.Fatalf("[FAIL] couuldn't initialize bucket: %e", err)
	}

	t.Cleanup(func() {
		t.Logf("cleaning up file at; %s", testpath)
		if err := os.RemoveAll(testpath); err != nil {
			t.Error(err)
		}
	})

	one := c.RNG(100)
	two := c.RNG(100)
	three := c.RNG(100)
	four := c.RNG(100)
	five := c.RNG(100)

	for n := 0; n != 100; n++ {
		var rawjson []byte
		switch n {
		case one, two, three, four, five:
			rawjson = genJunk(t, true)
		default:
			rawjson = genJunk(t, false)
		}
		t.Logf("created random data: %s", rawjson)
		err := db.With("searchtest").Put([]byte(fmt.Sprintf("%d", n)), rawjson)
		if err != nil {
			t.Fail()
			t.Logf("%e", err)
		}
	}

	t.Logf("executing search for %s", needle)
	results := db.With("searchtest").Search(needle)
	var keys = []int{one, two, three, four, five}
	var needed = len(keys)
	for key, value := range results {
		keyint, err := strconv.Atoi(key)
		if err != nil {
			t.Fatalf("failed to convert key to int: %e", err)
		}
		for _, k := range keys {
			if keyint == k {
				needed--
			}
		}
		keys = append(keys, keyint)
		t.Logf("Found key: %s, Value: %s", key, string(value.([]byte)))
	}
	if needed != 0 {
		t.Errorf("Needed %d results, got %d", len(keys), len(keys)-needed)
	}
}
