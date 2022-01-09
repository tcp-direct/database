package database

// Keeper will be in charge of the more meta operations involving Filers.
// This includes operations like initialization, syncing to disk if applicable, and backing up.
type Keeper interface {
	// Path should return the base path where all buckets should be stored under. (likely as subdirectories)
	Path() string
	// Init should initialize our Filer at the given path, to be referenced and called by bucketName.
	Init(bucketName string) error
	// With provides access to the given bucketName by providing a pointer to the related Filer.
	With(bucketName string) Filer
	// Close should safely end any Filer operations of the given bucketName and close any relevant handlers.
	Close(bucketName string) error
	// Sync should take any volatile data and solidify it somehow if relevant. (ram to disk in most cases)
	Sync(bucketName string) error

	// TODO: Backups

	CloseAll() error
	SyncAll() error
}
