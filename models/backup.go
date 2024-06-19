package models

import "encoding/json"

type Backup interface {
	Metadata() Metadata
	Format() string
	Path() string
	json.Marshaler
}
