package pogreb

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/akrylysov/pogreb"
)

type Option func(*WrappedOptions)

var OptionAllowRecovery = func(opts *WrappedOptions) {
	opts.AllowRecovery = true
}

func AllowRecovery() Option {
	return OptionAllowRecovery
}

func SetPogrebOptions(options pogreb.Options) Option {
	return func(opts *WrappedOptions) {
		opts.Options = &options
	}
}

type WrappedOptions struct {
	*pogreb.Options `json:"options"`
	// AllowRecovery allows the database to be recovered if a lockfile is detected upon running Init.
	AllowRecovery bool `json:"allow_recovery,omitempty"`
}

func (w *WrappedOptions) MarshalJSON() ([]byte, error) {
	optData, err := json.Marshal(w.Options)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Options       json.RawMessage `json:"options"`
		AllowRecovery bool            `json:"allow_recovery"`
	}{
		Options:       optData,
		AllowRecovery: w.AllowRecovery,
	})
}

var defaultPogrebOptions = &WrappedOptions{
	Options:       nil,
	AllowRecovery: false,
}

var defOptMu = sync.RWMutex{}

var ErrInvalidOptions = errors.New("invalid pogreb options")

func castOptions(pogrebopts ...any) (*WrappedOptions, error) {
	inner, pgoptOk := pogrebopts[0].(pogreb.Options)
	innerPtr, pgoptPtrOk := pogrebopts[0].(*pogreb.Options)
	wrapped, pgoptWrappedOk := pogrebopts[0].(*WrappedOptions)
	wrappedLiteral, pgoptWrappedLiteralOk := pogrebopts[0].(WrappedOptions)
	var ret *WrappedOptions
	//goland:noinspection GoDfaConstantCondition
	switch {
	case !pgoptOk && !pgoptWrappedOk && !pgoptPtrOk && !pgoptWrappedLiteralOk:
		return nil, ErrInvalidOptions
	case pgoptOk:
		ret = &WrappedOptions{
			Options:       &inner,
			AllowRecovery: false,
		}
	case pgoptPtrOk:
		ret = &WrappedOptions{
			Options:       innerPtr,
			AllowRecovery: false,
		}
	case pgoptWrappedLiteralOk:
		ret = &wrappedLiteral
	case pgoptWrappedOk:
		ret = wrapped
	}

	return ret, nil
}

// SetDefaultPogrebOptions options will set the options used for all subsequent pogreb stores that are initialized.
func SetDefaultPogrebOptions(pogrebopts ...any) (err error) {
	defOptMu.Lock()
	defaultPogrebOptions, err = castOptions(pogrebopts...)
	defOptMu.Unlock()
	return
}

func normalizeOptions(opts ...any) *WrappedOptions {
	if len(opts) == 0 {
		defOptMu.RLock()
		defOpt := defaultPogrebOptions
		defOptMu.RUnlock()
		return defOpt
	}

	opt, err := castOptions(opts[0])
	if err != nil {
		println("bad options: " + err.Error())
		return nil
	}
	return opt
}
