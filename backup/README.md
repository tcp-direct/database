# backup



#### func  RestoreTarGzBackup

```go
func RestoreTarGzBackup(inPath string, outPath string) error
```

#### func  VerifyBackup

```go
func VerifyBackup(metadata BackupMetadata) error
```

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


#### func  NewTarGzBackup

```go
func NewTarGzBackup(inPath string, outPath string, stores []string, extraData ...[]byte) (BackupMetadata, error)
```

#### func (BackupMetadata) Format

```go
func (bm BackupMetadata) Format() string
```

#### func (BackupMetadata) MarshalJSON

```go
func (bm BackupMetadata) MarshalJSON() ([]byte, error)
```

#### func (BackupMetadata) Path

```go
func (bm BackupMetadata) Path() string
```

#### func (BackupMetadata) Timestamp

```go
func (bm BackupMetadata) Timestamp() time.Time
```

#### func (BackupMetadata) Type

```go
func (bm BackupMetadata) Type() string
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


#### func (*TarGzBackup) Format

```go
func (tgz *TarGzBackup) Format() string
```

#### func (*TarGzBackup) Metadata

```go
func (tgz *TarGzBackup) Metadata() BackupMetadata
```

---
