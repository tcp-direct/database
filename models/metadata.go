package models

import "time"

// Metadata is an interface that defines the minimum requirements for [database.Keeper] metadata implementations.
type Metadata interface {
	Type() string
	// Timestamp should return the last time the metadata's parent was opened.
	Timestamp() time.Time
}
