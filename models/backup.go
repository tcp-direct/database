package models

import (
	"encoding/json"
	"time"
)

type Backup interface {
	Format() string
	Path() string
	Timestamp() time.Time
	json.Marshaler
}
