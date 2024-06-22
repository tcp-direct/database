# metadata



#### type Metadata

```go
type Metadata struct {
	KeeperType   string                   `json:"type"`
	Created      time.Time                `json:"created,omitempty"`
	LastOpened   time.Time                `json:"last_opened,omitempty"`
	KnownStores  []string                 `json:"stores,omitempty"`
	Backups      map[string]models.Backup `json:"backups,omitempty"`
	Extra        map[string]interface{}   `json:"extra,omitempty"`
	DefStoreOpts any                      `json:"default_store_opts,omitempty"`
}
```

Metadata is a struct that holds the metadata for a [Keeper]'s DB. This is
critical for migrating data between [Keeper]s. The only absolute requirement is
that the [Type] field is set.

#### func  NewMeta

```go
func NewMeta(keeperType string) *Metadata
```

#### func  NewMetaFile

```go
func NewMetaFile(keeperType, path string) (*Metadata, error)
```

#### func  OpenMetaFile

```go
func OpenMetaFile(path string) (*Metadata, error)
```

#### func (*Metadata) AddStore

```go
func (m *Metadata) AddStore(name string)
```

#### func (*Metadata) Close

```go
func (m *Metadata) Close() error
```
Close calls [Sync] and then closes the metadata writer, if it is an io.Closer.

#### func (*Metadata) Ping

```go
func (m *Metadata) Ping()
```

#### func (*Metadata) RemoveStore

```go
func (m *Metadata) RemoveStore(name string)
```

#### func (*Metadata) Sync

```go
func (m *Metadata) Sync() error
```
Sync writes the metadata to the designated [io.Writer]. If there is no writer,
it will create "meta.json" at m.path.

#### func (*Metadata) Timestamp

```go
func (m *Metadata) Timestamp() time.Time
```

#### func (*Metadata) Type

```go
func (m *Metadata) Type() string
```

#### func (*Metadata) WithBackups

```go
func (m *Metadata) WithBackups(backups ...models.Backup) *Metadata
```

#### func (*Metadata) WithCreated

```go
func (m *Metadata) WithCreated(created time.Time) *Metadata
```

#### func (*Metadata) WithDefaultStoreOpts

```go
func (m *Metadata) WithDefaultStoreOpts(opts any) *Metadata
```

#### func (*Metadata) WithExtra

```go
func (m *Metadata) WithExtra(extra map[string]interface{}) *Metadata
```

#### func (*Metadata) WithLastOpened

```go
func (m *Metadata) WithLastOpened(lastOpened time.Time) *Metadata
```

#### func (*Metadata) WithStores

```go
func (m *Metadata) WithStores(stores ...string) *Metadata
```

#### func (*Metadata) WithWriter

```go
func (m *Metadata) WithWriter(w io.WriteSeeker) *Metadata
```

---
