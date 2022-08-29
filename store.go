package database

// Store is an implementation of a Filer and a Searcher.
type Store interface {
	Filer
	Searcher
}
