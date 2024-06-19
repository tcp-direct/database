package bitcask

import (
	"errors"
	"fmt"
)

//goland:noinspection GoExportedElementShouldHaveComment
var (
	ErrUnknownAction = errors.New("unknown action")
	ErrBogusStore    = errors.New("bogus store backend")
	ErrStoreExists   = errors.New("store name already exists")
	ErrNoStores      = errors.New("no stores initialized")
)

func namedErr(name string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", name, err)
}

func compoundErrors(errs []error) (err error) {
	return errors.Join(errs...)
}
