package metadata

import (
	"errors"

	"github.com/tcp-direct/database/models"
)

var ErrNotCanonicalMetadata = errors.New("metadata is of a different type, cannot cast")

func IsCanonicalMetadata(m models.Metadata) bool {
	_, ok := m.(*Metadata)
	return ok
}

func CastToMetadata(m models.Metadata) (*Metadata, error) {
	if !IsCanonicalMetadata(m) {
		return nil, ErrNotCanonicalMetadata
	}
	return m.(*Metadata), nil
}
