package pogreb

import (
	"encoding/json"

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
	*pogreb.Options
	// AllowRecovery allows the database to be recovered if a lockfile is detected upon running Init.
	AllowRecovery bool
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

// SetDefaultPogrebOptions options will set the options used for all subsequent pogreb stores that are initialized.
func SetDefaultPogrebOptions(pogrebopts ...any) {
	inner, pgoptOk := pogrebopts[0].(pogreb.Options)
	innerPtr, pgoptPtrOk := pogrebopts[0].(*pogreb.Options)
	wrapped, pgoptWrappedOk := pogrebopts[0].(*WrappedOptions)
	wrappedLiteral, pgoptWrappedLiteralOk := pogrebopts[0].(WrappedOptions)
	//goland:noinspection GoDfaConstantCondition
	switch {
	case !pgoptOk && !pgoptWrappedOk && !pgoptPtrOk && !pgoptWrappedLiteralOk:
		panic("invalid pogreb options")
	case pgoptOk:
		defaultPogrebOptions = &WrappedOptions{
			Options:       &inner,
			AllowRecovery: false,
		}
	case pgoptPtrOk:
		defaultPogrebOptions = &WrappedOptions{
			Options:       innerPtr,
			AllowRecovery: false,
		}
	case pgoptWrappedLiteralOk:
		defaultPogrebOptions = &wrappedLiteral
	case pgoptWrappedOk:
		defaultPogrebOptions = wrapped
	}
}

func normalizeOptions(opts ...any) *WrappedOptions {
	var pogrebopts *WrappedOptions
	pgInner, pgOK := opts[0].(pogreb.Options)
	pgWrapped, pgWrappedOK := opts[0].(WrappedOptions)
	//goland:noinspection GoDfaConstantCondition
	switch {
	case !pgOK && !pgWrappedOK:
		return nil
	case pgOK:
		pogrebopts = &WrappedOptions{
			Options:       &pgInner,
			AllowRecovery: false,
		}
	case pgWrappedOK:
		pogrebopts = &pgWrapped
	}
	return pogrebopts
}
