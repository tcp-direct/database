package models

import "time"

type Metadata interface {
	Type() string
	Timestamp() time.Time
}
