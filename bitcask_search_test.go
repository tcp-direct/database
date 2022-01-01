package database

import (
	"encoding/json"
	"fmt"
	"os"
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
	if correct {
		item = needle + c.RandStr(5)
	}

	f := Foo{
		Bar:  c.RandStr(5),
		Yeet: 5,
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

	for n := 0; n != 100; n++ {
		var rawjson []byte
		switch n {
		case 5, 25, 35, 85:
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
	for key, value := range results {
		t.Logf("Found key: %s, Value: %s", key, string(value.([]byte)))
	}
}
