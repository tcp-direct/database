package database

import "errors"

type KeeperCreator func(path string, opt ...any) (Keeper, error)

var ErrNotStore = errors.New("provided Filer does not implement Store")

func IsStore(filer Filer) bool {
	_, ok := filer.(Store)
	return ok
}

func ToStore(filer Filer) (Store, error) {
	if IsStore(filer) {
		return filer.(Store), nil
	}
	return nil, ErrNotStore
}
