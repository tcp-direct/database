package database

import "git.tcp.direct/tcp.direct/database/models"

// Keeper will be in charge of the more meta operations involving Filers.
// This includes operations like initialization, syncing to disk if applicable, and backing up.
//   - When opening a folder of Filers, it should be able to discover and initialize all of them.
//   - Additionally, it should be able to confirm the type of the underlying key/value store.
type Keeper interface {
	// Path should return the base path where all stores should be stored under. (likely as subdirectories)
	Path() string

	// Init should initialize our Filer at the given path, to be referenced and called by dataStore.
	Init(name string, options ...any) error
	// With provides access to the given dataStore by providing a pointer to the related Filer.
	With(name string) Store
	// WithNew should initialize a new Filer at the given path and return a pointer to it.
	WithNew(name string, options ...any) Filer

	// Destroy should remove the Filer by the given name.
	// It is up to the implementation to decide if the data should be removed or not.
	Destroy(name string) error

	Discover() ([]string, error)

	AllStores() map[string]Filer

	// BackupAll should create a backup of all [Filer] instances in the [Keeper].
	BackupAll(archivePath string) (models.Backup, error)

	// RestoreAll should restore all [Filer] instances from the given archive.
	RestoreAll(archivePath string) error

	Meta() models.Metadata

	Close(name string) error

	CloseAll() error
	SyncAll() error
	SyncAndCloseAll() error
}

type KeeperCreator func(path string) (Keeper, error)
