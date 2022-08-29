package database

// Keeper will be in charge of the more meta operations involving Filers.
// This includes operations like initialization, syncing to disk if applicable, and backing up.
type Keeper interface {
	// Path should return the base path where all stores should be stored under. (likely as subdirectories)
	Path() string
	// Init should initialize our Filer at the given path, to be referenced and called by dataStore.
	Init(name string, options ...any) error
	// With provides access to the given dataStore by providing a pointer to the related Filer.
	With(name string) Store

	AllStores() map[string]Filer
	// TODO: Backups

	CloseAll() error
	SyncAll() error
}
