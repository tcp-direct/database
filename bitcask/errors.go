package bitcask

import (
	"errors"
	"strings"
)

var (
	errUnknownAction = errors.New("unknown action")
	errBogusStore    = errors.New("bogus store backend")
	errStoreExists   = errors.New("store name already exists")
	errNoStores      = errors.New("no stores initialized")
)

func namedErr(name string, err error) error {
	if err == nil {
		return nil
	}
	return errors.New(name + ": " + err.Error())
}

func compoundErrors(errs []error) error {
	var errstrs []string
	var isnil = true
	for _, err := range errs {
		if err != nil {
			isnil = false
			errstrs = append(errstrs, err.Error())
		}
	}
	if isnil {
		return nil
	}
	return errors.New(strings.Join(errstrs, ","))
}
