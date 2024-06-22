package metadata

import (
	"time"
)

type Placeholder struct {
	name string
}

func (d Placeholder) Type() string {
	return d.name
}

func (d Placeholder) Timestamp() time.Time {
	return time.Now()
}

func NewPlaceholder(name string) Placeholder {
	return Placeholder{name: name}
}
