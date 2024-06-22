# models



#### type Backup

```go
type Backup interface {
	Format() string
	Path() string
	Timestamp() time.Time
	json.Marshaler
}
```


#### type Metadata

```go
type Metadata interface {
	Type() string
	// Timestamp should return the last time the metadata's parent was opened.
	Timestamp() time.Time
}
```

Metadata is an interface that defines the minimum requirements for
[database.Keeper] metadata implementations.

---
