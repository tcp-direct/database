# models



#### type Backup

```go
type Backup interface {
	Metadata() Metadata
	Format() string
	Path() string
}
```


#### type Metadata

```go
type Metadata interface {
	Type() string
	Timestamp() time.Time
}
```

---
