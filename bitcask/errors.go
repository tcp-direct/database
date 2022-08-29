package bitcask

import (
	"errors"

	"github.com/hashicorp/go-multierror"
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
	return multierror.Prefix(err, name)
}

func compoundErrors(errs []error) (err error) {
	for _, e := range errs {
		if e == nil {
			continue
		}
		err = multierror.Append(err, e)
	}
	return
}
