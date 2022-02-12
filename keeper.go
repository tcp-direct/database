package database

// Keeper will be in charge of the more meta operations involving Filers.
// This includes operations like initialization, syncing to disk if applicable, and backing up.
type Keeper interface {
	// Path should return the base path where all stores should be stored under. (likely as subdirectories)
	Path() string
	// Init should initialize our Filer at the given path, to be referenced and called by dataStore.
	Init(dataStore []byte) error
	// With provides access to the given dataStore by providing a pointer to the related Filer.
	With(dataStore []byte) Filer
	// Close should safely end any Filer operations of the given dataStore and close any relevant handlers.
	Close(dataStore []byte) error
	// Sync should take any volatile data and solidify it somehow if relevant. (ram to disk in most cases)
	Sync(dataStore []byte) error

	// TODO: Backups

	CloseAll() error
	SyncAll() error
}
