package pogreb

import (
	"errors"
	"fmt"
)

//goland:noinspection GoExportedElementShouldHaveComment
var (
	ErrUnknownAction = errors.New("unknown action")
	ErrBogusStore    = errors.New("bogus store backend")
	ErrBadOptions    = errors.New("invalid pogreb options")
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
