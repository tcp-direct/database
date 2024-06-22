# migrate

Package migrate implements the migration of data from one type of Keeper to
another.


```go
var (
	ErrNoStores = errors.New("no stores found in source keeper")
	ErrDupKeys  = errors.New(
		"duplicate keys found in destination stores, enable skipping or clobbering of existing data to continue migration",
	)
)
```

#### type ErrDuplicateKeys

```go
type ErrDuplicateKeys struct {
	// map[store][]keys
	Duplicates map[string][][]byte
}
```


#### func  NewDuplicateKeysErr

```go
func NewDuplicateKeysErr(duplicates map[string][][]byte) *ErrDuplicateKeys
```

#### func (ErrDuplicateKeys) Error

```go
func (e ErrDuplicateKeys) Error() string
```

#### func (ErrDuplicateKeys) Unwrap

```go
func (e ErrDuplicateKeys) Unwrap() error
```

#### type Migrator

```go
type Migrator struct {
	From database.Keeper
	To   database.Keeper
}
```


#### func  NewMigrator

```go
func NewMigrator(from, to database.Keeper) (*Migrator, error)
```

#### func (*Migrator) CheckDupes

```go
func (m *Migrator) CheckDupes() error
```

#### func (*Migrator) Migrate

```go
func (m *Migrator) Migrate() error
```

#### func (*Migrator) WithClobber

```go
func (m *Migrator) WithClobber() *Migrator
```
WithClobber sets the clobber flag on the Migrator, allowing it to overwrite
existing data in the destination Keeper.

#### func (*Migrator) WithSkipExisting

```go
func (m *Migrator) WithSkipExisting() *Migrator
```
WithSkipExisting sets the skipExisting flag on the Migrator, allowing it to skip
existing data in the destination Keeper.

---
