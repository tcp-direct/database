package models

type Backup interface {
	Metadata() Metadata
	Format() string
	Path() string
}
