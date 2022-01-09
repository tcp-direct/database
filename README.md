# database

#### type Casket

```go
type Casket struct {
	*bitcask.Bitcask
	Searcher
}
```

Casket is an implmentation of a Filer and a Searcher using Bitcask.

#### func (Casket) AllKeys

```go
func (c Casket) AllKeys() (keys []string)
```

#### func (Casket) PrefixScan

```go
func (c Casket) PrefixScan(prefix string) map[string]interface{}
```

#### func (Casket) Search

```go
func (c Casket) Search(query string) map[string]interface{}
```

#### func (Casket) ValueExists

```go
func (c Casket) ValueExists(value []byte) (key []byte, ok bool)
```

#### type DB

```go
type DB struct {
}
```

DB is a mapper of a Filer and Searcher implementation using Bitcask.

#### func  OpenDB

```go
func OpenDB(path string) *DB
```
OpenDB will either open an existing set of bitcask datastores at the given
directory, or it will create a new one.

#### func (*DB) Close

```go
func (db *DB) Close(bucketName string) error
```
Close is a simple shim for bitcask's Close function.

#### func (*DB) CloseAll

```go
func (db *DB) CloseAll() error
```
CloseAll closes all bitcask datastores.

#### func (*DB) Init

```go
func (db *DB) Init(bucketName string) error
```
Init opens a bitcask store at the given path to be referenced by bucketName.

#### func (*DB) Path

```go
func (db *DB) Path() string
```
Path returns the base path where we store our bitcask "buckets".

#### func (*DB) Sync

```go
func (db *DB) Sync(bucketName string) error
```
Sync is a simple shim for bitcask's Sync function.

#### func (*DB) SyncAll

```go
func (db *DB) SyncAll() error
```
SyncAll syncs all bitcask datastores.

#### func (*DB) SyncAndCloseAll

```go
func (db *DB) SyncAndCloseAll() error
```
SyncAndCloseAll implements the method from Keeper.

#### func (*DB) With

```go
func (db *DB) With(bucketName string) Casket
```
With calls the given underlying bitcask instance.

#### func (*DB) WithAll

```go
func (db *DB) WithAll(action withAllAction) error
```
WithAll performs an action on all bitcask stores that we have open. In the case
of an error, WithAll will continue and return a compound form of any errors that
occurred. For now this is just for Close and Sync, thusly it does a hard lock on
the Keeper.

#### type Filer

```go
type Filer interface {

	// Has should return true if the given key has an associated value.
	Has(key []byte) bool
	// Get should retrieve the byte slice corresponding to the given key, and any associated errors upon failure.
	Get(key []byte) ([]byte, error)
	// Put should insert the value data in a way that is associated and can be retrieved by the given key data.
	Put(key []byte, value []byte) error
	// Delete should delete the key and the value associated with the given key, and return an error upon failure.
	Delete(key []byte) error
}
```

Filer is is a way to implement any generic key/value store. These functions
should be plug and play with most of the popular key/value store golang
libraries.

#### type Keeper

```go
type Keeper interface {
	// Path should return the base path where all buckets should be stored under. (likely as subdirectories)
	Path() string
	// Init should initialize our Filer at the given path, to be referenced and called by bucketName.
	Init(bucketName string) error
	// With provides access to the given bucketName by providing a pointer to the related Filer.
	With(bucketName string) Filer
	// Close should safely end any Filer operations of the given bucketName and close any relevant handlers.
	Close(bucketName string) error
	// Sync should take any volatile data and solidify it somehow if relevant. (ram to disk in most cases)
	Sync(bucketName string) error

	CloseAll() error
	SyncAll() error
}
```

Keeper will be in charge of the more meta operations involving Filers. This
includes operations like initialization, syncing to disk if applicable, and
backing up.

NOTE: Many key/value golang libraries will already implement this interface
already. This exists for more potential granular control in the case that they
don't. Otherwise you'd have to build a wrapper around an existing key/value
store to satisfy an overencompassing interface.

#### type Searcher

```go
type Searcher interface {
	// AllKeys must retrieve all keys in the datastore with the given bucketName.
	AllKeys() []string
	// PrefixScan must return all keys that begin with the given prefix.
	PrefixScan(prefix string) map[string]interface{}
	// Search must be able to search through the contents of our database and return a map of results.
	Search(query string) map[string]interface{}
	// ValueExists searches for an exact match of the given value and returns the key that contains it.
	ValueExists(value []byte) (key []byte, ok bool)
}
```

Searcher must be able to search through our datastore(s) with strings.


---

## Test results

```
=== RUN   Test_Search
    bitcask_search_test.go:52: opening database at ./testsearch
    bitcask_search_test.go:82: created random data: {"Bar":"io66b","Yeet":25,"What":{"62aqk":7,"zaj5n":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"nk63e","Yeet":11,"What":{"pg3xt":2,"ykg2q":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"41or3","Yeet":40,"What":{"btgkc":2,"fy5jt":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"jglc5","Yeet":14,"What":{"twe6q":7,"zsrme":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"jfekj","Yeet":28,"What":{"5vonr":7,"rzg1f":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"p1rlx","Yeet":37,"What":{"a1vdr":7,"x6fgc":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"gdlau","Yeet":31,"What":{"c5f2y":7,"mayoc":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"c31fi","Yeet":23,"What":{"dqtcc":7,"y2gca":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"e35oc","Yeet":19,"What":{"ip6i4":2,"wf6x2":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"ianpq","Yeet":46,"What":{"fioll":2,"yhh2e":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"ikjb3","Yeet":10,"What":{"q3ax4":2,"weyly":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"zrrdd","Yeet":31,"What":{"aposv":2,"tlnjz":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"sk5qr","Yeet":6,"What":{"5n3nt":7,"knmd5":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"m634o","Yeet":12,"What":{"jkg1l":2,"lwfvx":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"w6nhe","Yeet":43,"What":{"kmkls":2,"qklch":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"lomkh","Yeet":35,"What":{"6qbzi":7,"jehqz":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"4nvwk","Yeet":47,"What":{"gn3ol":2,"ngbtb":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"f3j5o","Yeet":48,"What":{"1lkyr":2,"6wqae":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"6qstc","Yeet":13,"What":{"2xsjh":7,"ceiwv":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"nhqua","Yeet":10,"What":{"smfx2":7,"vxj3z":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"v6rzc","Yeet":2,"What":{"qowh1":2,"xd364":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"b3krt","Yeet":9,"What":{"155tj":2,"3m56p":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"365ov","Yeet":42,"What":{"ghpco":2,"jjuwn":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"agntr","Yeet":39,"What":{"pqsfq":2,"wsg1i":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"tdznl","Yeet":13,"What":{"nb2b3":2,"yzzpf":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"iqa4j","Yeet":9,"What":{"n1olp":2,"psilb":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"etd6s","Yeet":31,"What":{"2iqrz":2,"jfvgd":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"o3lwi","Yeet":39,"What":{"6pw6n":7,"yusp4":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"3ffsr","Yeet":9,"What":{"jizvu":2,"os6j2":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"go2ja","Yeet":48,"What":{"mdzi6":2,"rikkx":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"rrmum","Yeet":32,"What":{"emkgs":7,"qd1c5":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"kkdq3","Yeet":0,"What":{"e4xqm":2,"fw2of":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"lh2t6","Yeet":19,"What":{"6zi2d":2,"hkn2x":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"wswpt","Yeet":19,"What":{"1q32w":7,"gspzi":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"lqaf2","Yeet":44,"What":{"fzgpm":7,"ir2ul":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"mhqef","Yeet":46,"What":{"gyugg":7,"sslov":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"hwfhe","Yeet":43,"What":{"xsohr":7,"zjpbm":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"ayp1z","Yeet":48,"What":{"p3hz2":2,"r5wwv":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"xk5rp","Yeet":7,"What":{"a2dy6":2,"e6eij":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"xbkc5","Yeet":16,"What":{"s32ph":7,"ujcwz":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"tafyeet","Yeet":20,"What":{"gp3xt":2,"rgg2v":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"3sraj","Yeet":9,"What":{"5wvbd":2,"qpdn5":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"y4hls","Yeet":47,"What":{"6hmbi":2,"wzzm3":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"apbea","Yeet":47,"What":{"aoryg":2,"uab3e":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"mxui1","Yeet":44,"What":{"52sbe":7,"qcru3":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"a12qa","Yeet":23,"What":{"mko6a":2,"wxq63":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"lubhv","Yeet":12,"What":{"1k3z5":7,"qbolc":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"t64zp","Yeet":31,"What":{"omfew":7,"ryw4x":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"3qww4","Yeet":32,"What":{"eftdz":7,"w325d":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"zceap","Yeet":34,"What":{"bbxza":7,"kwwpc":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"guk6a","Yeet":26,"What":{"mztj2":7,"wjdr1":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"keihl","Yeet":18,"What":{"nzrym":2,"ty5a1":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"tf4dv","Yeet":16,"What":{"kf6mt":7,"wquku":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"ivu66","Yeet":3,"What":{"1ra2z":2,"vkkdz":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"zxi4e","Yeet":11,"What":{"2cnvb":7,"sptra":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"yeet","Yeet":21,"What":{"j2xpl":2,"p3ozp":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"t6dvv","Yeet":3,"What":{"s3znd":2,"sidp6":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"tpjp5","Yeet":9,"What":{"61nli":7,"snsee":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"dvf4j","Yeet":4,"What":{"2hcid":7,"ln3l5":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"1cgso","Yeet":15,"What":{"5vz5c":7,"p46t1":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"c2ys3","Yeet":34,"What":{"2q5sh":2,"bnkzj":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"4f6zh","Yeet":11,"What":{"pjf6v":7,"qi4jd":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"2zpv6","Yeet":25,"What":{"ue2hd":7,"yp2kn":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"niicy","Yeet":1,"What":{"aonjd":2,"vxadi":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"4q5gb","Yeet":14,"What":{"5qdj5":7,"w1lgi":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"a3ere","Yeet":1,"What":{"b6fme":7,"z13gr":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"sjfbm","Yeet":12,"What":{"3yw3w":2,"q4cbv":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"gdusg","Yeet":17,"What":{"6gye2":7,"cg6bm":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"aqhfp","Yeet":25,"What":{"23gov":2,"iu5eu":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"skvoq","Yeet":21,"What":{"pihyz":2,"qmzry":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"pyeet","Yeet":25,"What":{"ayn3q":2,"qmucb":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"hllbu","Yeet":26,"What":{"bzz5w":7,"q6map":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"kl34n","Yeet":32,"What":{"1mnlx":2,"zpgwd":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"2hfaj","Yeet":12,"What":{"4c2k6":7,"rjujd":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"slmxq","Yeet":6,"What":{"ocf1b":7,"xdfws":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"1xrup","Yeet":8,"What":{"5zujd":7,"sxrx1":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"coujr","Yeet":31,"What":{"ir4gh":7,"og5f5":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"du66h","Yeet":48,"What":{"h1dlz":2,"s5uqx":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"2dura","Yeet":3,"What":{"gswvk":2,"kqcgq":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"he42u","Yeet":19,"What":{"ucwet":7,"zu42v":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"k6tta","Yeet":37,"What":{"cnxwr":7,"olifa":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"cbgpg","Yeet":43,"What":{"fhkjj":2,"zm4jh":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"jctlo","Yeet":19,"What":{"njurv":2,"pgiwo":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"soa4k","Yeet":18,"What":{"1besq":7,"p6yij":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"aq54m","Yeet":45,"What":{"jjpzy":7,"ocby4":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"xxsvyeet","Yeet":21,"What":{"3ezix":7,"g5h6b":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"yszrz","Yeet":36,"What":{"cqfhf":2,"lzgrn":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"hkxvz","Yeet":15,"What":{"awynm":7,"rdg5l":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"tlb1v","Yeet":27,"What":{"2ljrv":7,"dxm5g":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"f3z6f","Yeet":41,"What":{"bvof2":7,"k6nkw":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"3ctor","Yeet":4,"What":{"omf2e":2,"vebx3":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"nw2oq","Yeet":8,"What":{"jridl":7,"kbylb":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"fuwzi","Yeet":3,"What":{"aanf3":2,"s5uqq":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"gtrdv","Yeet":26,"What":{"o4kgl":2,"q5sip":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"aoyba","Yeet":11,"What":{"n13m4":7,"xhtwl":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"w2jwm","Yeet":18,"What":{"iltin":2,"x66c3":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"54vhp","Yeet":8,"What":{"lnffy":7,"m5fk2":2}}
    bitcask_search_test.go:82: created random data: {"Bar":"oxbrq","Yeet":0,"What":{"u6n6m":2,"yh6yd":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"gu4s4","Yeet":7,"What":{"dnyv5":2,"rpt33":7}}
    bitcask_search_test.go:82: created random data: {"Bar":"u2fac","Yeet":1,"What":{"kftbe":2,"yeetbs5x":7}}
    bitcask_search_test.go:90: executing search for yeet
    bitcask_search_test.go:105: Found key: 40, Value: {"Bar":"tafyeet","Yeet":20,"What":{"gp3xt":2,"rgg2v":7}}
    bitcask_search_test.go:105: Found key: 55, Value: {"Bar":"yeet","Yeet":21,"What":{"j2xpl":2,"p3ozp":7}}
    bitcask_search_test.go:105: Found key: 70, Value: {"Bar":"pyeet","Yeet":25,"What":{"ayn3q":2,"qmucb":7}}
    bitcask_search_test.go:105: Found key: 85, Value: {"Bar":"xxsvyeet","Yeet":21,"What":{"3ezix":7,"g5h6b":2}}
    bitcask_search_test.go:105: Found key: 99, Value: {"Bar":"u2fac","Yeet":1,"What":{"kftbe":2,"yeetbs5x":7}}
    bitcask_search_test.go:62: cleaning up file at; ./testsearch
--- PASS: Test_Search (0.02s)
=== RUN   TestDB_NewDB
--- PASS: TestDB_NewDB (0.00s)
=== RUN   TestDB_Init
=== RUN   TestDB_Init/simple
=== RUN   TestDB_Init/bucketExists
=== RUN   TestDB_Init/newBucket
=== RUN   TestDB_Init/withBucketTest
    bitcask_test.go:60: Put value string at key [51 50]
    bitcask_test.go:71: Got value string at key [51 50]
=== RUN   TestDB_Init/withBucketDoesntExist
    bitcask_test.go:77: [SUCCESS] got nil value for bucket that doesn't exist
=== RUN   TestDB_Init/syncAllShouldFail
    bitcask_test.go:89: [SUCCESS] got compound error: &{%!e(string=wtf: bogus store backend)}
    bitcask_test.go:82: deleting bogus store map entry
=== RUN   TestDB_Init/syncAll
=== RUN   TestDB_Init/closeAll
    bitcask_test.go:103: cleaned up ./testdata
--- PASS: TestDB_Init (0.03s)
    --- PASS: TestDB_Init/simple (0.00s)
    --- PASS: TestDB_Init/bucketExists (0.00s)
    --- PASS: TestDB_Init/newBucket (0.00s)
    --- PASS: TestDB_Init/withBucketTest (0.00s)
    --- PASS: TestDB_Init/withBucketDoesntExist (0.00s)
    --- PASS: TestDB_Init/syncAllShouldFail (0.01s)
    --- PASS: TestDB_Init/syncAll (0.00s)
    --- PASS: TestDB_Init/closeAll (0.02s)
PASS
ok  	git.tcp.direct/kayos/database	0.051s

```
