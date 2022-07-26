# kv

`import "git.tcp.direct/tcp.direct/database/kv"`

## Documentation

#### type Key

```go
type Key struct {}
```

Key represents a key in a key/value store.

#### func  NewKey

```go
func NewKey(data []byte) *Key
```
NewKey creates a new Key from a byte slice.

#### func (*Key) Bytes

```go
func (k *Key) Bytes() []byte
```
Bytes returns the raw byte slice form of the Key.

#### func (*Key) Equal

```go
func (k *Key) Equal(k2 *Key) bool
```
Equal determines if two keys are equal.

#### func (*Key) String

```go
func (k *Key) String() string
```
String returns the string slice form of the Key.

#### type KeyValue

```go
type KeyValue struct {
	Key   *Key
	Value *Value
}
```

KeyValue represents a key and a value from a key/value store.

#### func  NewKeyValue

```go
func NewKeyValue(k *Key, v *Value) *KeyValue
```
NewKeyValue creates a new KeyValue from a key and value.

#### func (*KeyValue) Equal

```go
func (kv *KeyValue) Equal(kv2 *KeyValue) bool
```
Equal determines if two key/value pairs are equal.

#### func (*KeyValue) String

```go
func (kv *KeyValue) String() string
```

#### type Value

```go
type Value struct {}
```

Value represents a value in a key/value store.

#### func  NewValue

```go
func NewValue(data []byte) *Value
```
NewValue creates a new Value from a byte slice.

#### func (*Value) Bytes

```go
func (v *Value) Bytes() []byte
```
Bytes returns the raw byte slice form of the Value.

#### func (*Value) Equal

```go
func (v *Value) Equal(v2 *Value) bool
```
Equal determines if two values are equal.

#### func (*Value) String

```go
func (v *Value) String() string
```
String returns the string slice form of the Value.
