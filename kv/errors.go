package kv

import (
	"errors"
	"fmt"
)

// NonExistentKeyError is an error type for when a key does not exist or has no value.
// This allows us to maintain consistency in error handling across different key-value stores.
// See [RegularizeKVError] for more information.
type NonExistentKeyError struct {
	Key        []byte
	Underlying error
}

func (neke *NonExistentKeyError) Error() string {
	if neke.Underlying != nil {
		return fmt.Sprintf("key %s does not exist or has no value: %s", neke.Key, neke.Underlying)
	}
	return fmt.Sprintf("key %s does not exist or has no value", neke.Key)
}

// Unwrap returns the underlying error, if any. This implements the errors.Wrapper interface.
func (neke *NonExistentKeyError) Unwrap() error {
	return neke.Underlying
}

// RegularizeKVError returns a regularized error for a key-value store.
// This exists because some key-value stores return nil for a value and nil for an error when a key does not exist.
func RegularizeKVError(key []byte, value []byte, err error) error {
	neke := &NonExistentKeyError{}
	switch {
	case err == nil && value != nil:
		return nil
	case err == nil: // && value == nil
		neke.Key = key
		return neke
	case value == nil: // && err != nil
		neke.Key = key
		neke.Underlying = err
		return neke
	default: // err != nil && value != nil
		return err
	}
}

// IsNonExistentKey returns true if the error is a [NonExistentKeyError]. This is syntactic sugar.
func IsNonExistentKey(err error) bool {
	neke := &NonExistentKeyError{}
	if !errors.As(err, &neke) {
		return false
	}
	return true
}
