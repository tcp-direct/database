package database

// Key represents a key in a key/value Filer.
type Key interface {
	Bytes() []byte
	String() string
	Equal(Key) bool
}

// Value represents a value in a key/value Filer.
type Value interface {
	Bytes() []byte
	String() string
	Equal(Value) bool
}
