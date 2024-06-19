package metadata

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"git.tcp.direct/tcp.direct/database/models"
)

// Metadata is a struct that holds the metadata for a [Keeper]'s DB.
// This is critical for migrating data between [Keeper]s.
// The only absolute requirement is that the [Type] field is set.
type Metadata struct {
	KeeperType   string                   `json:"type"`
	Created      time.Time                `json:"created,omitempty"`
	LastOpened   time.Time                `json:"last_opened,omitempty"`
	KnownStores  []string                 `json:"stores,omitempty"`
	Backups      map[string]models.Backup `json:"backups,omitempty"`
	Extra        map[string]interface{}   `json:"extra,omitempty"`
	DefStoreOpts any                      `json:"default_store_opts,omitempty"`
	w            io.WriteCloser
	path         string
}

func (m *Metadata) Type() string {
	return m.KeeperType
}

func (m *Metadata) Ping() {
	m.LastOpened = time.Now()
}

func (m *Metadata) AddStore(name string) {
	m.KnownStores = append(m.KnownStores, name)
}

func (m *Metadata) RemoveStore(name string) {
	var newStores []string
	for _, store := range m.KnownStores {
		if store != name {
			newStores = append(newStores, store)
		}
	}
	m.KnownStores = newStores
}

func (m *Metadata) Timestamp() time.Time {
	return m.LastOpened
}

func NewMeta(keeperType string) *Metadata {
	return &Metadata{
		KeeperType:  keeperType,
		Created:     time.Now(),
		LastOpened:  time.Now(),
		KnownStores: make([]string, 0),
		Backups:     make(map[string]models.Backup),
	}
}

func (m *Metadata) WithExtra(extra map[string]interface{}) *Metadata {
	m.Extra = extra
	return m
}

func (m *Metadata) WithDefaultStoreOpts(opts any) *Metadata {
	m.DefStoreOpts = opts
	return m
}

func NewMetaFile(keeperType, path string) (*Metadata, error) {
	meta := &Metadata{
		KeeperType:  keeperType,
		Created:     time.Now(),
		LastOpened:  time.Now(),
		KnownStores: make([]string, 0),
		Backups:     make(map[string]models.Backup),
		path:        path,
	}
	stat, err := os.Stat(path)
	if err == nil && stat.IsDir() {
		path = filepath.Join(path, "meta.json")
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	jsonData, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	_, err = file.Write(jsonData)
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	if err = file.Sync(); err != nil {
		_ = file.Close()
		return nil, err
	}
	meta.w = file
	return meta, nil
}

func OpenMetaFile(path string) (*Metadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	meta := &Metadata{}
	err = json.Unmarshal(data, meta)
	if err != nil {
		return nil, err
	}
	if meta.KeeperType == "" {
		return nil, errors.New("metadata file does not have a type")
	}
	meta.path = path
	return meta, nil
}

func (m *Metadata) WithStores(stores ...string) *Metadata {
	m.KnownStores = stores
	return m
}

func (m *Metadata) WithCreated(created time.Time) *Metadata {
	m.Created = created
	return m
}

func (m *Metadata) WithLastOpened(lastOpened time.Time) *Metadata {
	m.LastOpened = lastOpened
	return m
}

func (m *Metadata) WithBackups(backups ...models.Backup) *Metadata {
	for _, bu := range backups {
		m.Backups[bu.Metadata().Timestamp().String()] = bu
	}
	return m
}

func (m *Metadata) WithWriter(w io.WriteCloser) *Metadata {
	if m.w != nil {
		if m.w.(io.Closer) != nil {
			_ = m.w.(io.Closer).Close()
		}
	}
	m.w = w
	return m
}

// Sync writes the metadata to the designated [io.Writer]. If there is no writer, it will create "meta.json" at m.path.
func (m *Metadata) Sync() error {
	dat, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if m.w == nil {
		if m.w, err = os.Create(m.path); err != nil {
			return err
		}
	}
	if _, err = m.w.Write(dat); err != nil {
		return err
	}
	return err
}

// Close calls [Sync] and then closes the metadata writer, if it is an io.Closer.
func (m *Metadata) Close() error {
	if err := m.Sync(); err != nil {
		return err
	}
	if closer, closerOK := m.w.(io.Closer); closerOK {
		return closer.Close()
	}
	return nil
}
