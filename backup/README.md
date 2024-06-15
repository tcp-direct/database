# backup



#### type BackupMetadata

```go
type BackupMetadata struct {
	Date       time.Time `json:"timestamp"`
	FileFormat string    `json:"format"`
	FilePath   string    `json:"path"`
	Stores     []string  `json:"stores,omitempty"`
	Checksum   Checksum  `json:"checksum,omitempty"`
	Size       int64     `json:"size,omitempty"`
}
```


#### func (BackupMetadata) Format

```go
func (bm BackupMetadata) Format() string
```

#### func (BackupMetadata) Path

```go
func (bm BackupMetadata) Path() string
```

#### func (BackupMetadata) Timestamp

```go
func (bm BackupMetadata) Timestamp() time.Time
```

#### type Checksum

```go
type Checksum struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}
```


#### type Format

```go
type Format string
```


```go
const (
	FormatTarGz Format = "tar.gz"
	FormatTar   Format = "tar"
	FormatZip   Format = "zip"
)
```

#### type TarGzBackup

```go
type TarGzBackup struct {
}
```


#### func  NewTarGzBackup

```go
func NewTarGzBackup(inPath string, outPath string, stores []string, extraData ...[]byte) (*TarGzBackup, error)
```

#### func (*TarGzBackup) Format

```go
func (tgz *TarGzBackup) Format() string
```

#### func (*TarGzBackup) Metadata

```go
func (tgz *TarGzBackup) Metadata() BackupMetadata
```

---
