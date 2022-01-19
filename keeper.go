package database

// Keeper will be in charge of the more meta operations involving Filers.
// This includes operations like initialization, syncing to disk if applicable, and backing up.
type Keeper interface {
	// Path should return the base path where all stores should be stored under. (likely as subdirectories)
	Path() string
	// Init should initialize our Filer at the given path, to be referenced and called by storeName.
	Init(storeName string) error
	// With provides access to the given storeName by providing a pointer to the related Filer.
	With(storeName string) Filer
	// Close should safely end any Filer operations of the given storeName and close any relevant handlers.
	Close(storeName string) error
	// Sync should take any volatile data and solidify it somehow if relevant. (ram to disk in most cases)
	Sync(storeName string) error

	// TODO: Backups

	CloseAll() error
	SyncAll() error
}
