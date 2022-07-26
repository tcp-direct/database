package database

type Store interface {
	Filer
	Searcher
}
