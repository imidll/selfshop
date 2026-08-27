package appkit

import (
	"bytes"
	"fmt"
)

var ErrUnmarshalNilRunmode = fmt.Errorf("can't unmarshal a nil runmode")

type Runmode int8

const (
	RunmodeDev Runmode = iota - 1
	RunmodeProd

	_minRunmode = RunmodeDev
	_maxRunmode = RunmodeProd

	InvalidRunmode = _maxRunmode + 1
)

func ParseRunmode(text string) (Runmode, error) {
	var r Runmode
	err := r.UnmarshalText([]byte(text))
	return r, err
}

func (r Runmode) String() string {
	switch r {
	case RunmodeDev:
		return "dev"
	case RunmodeProd:
		return "prod"
	default:
		return fmt.Sprintf("unknown(%d)", r)
	}
}

func (r *Runmode) Set(s string) error { return r.UnmarshalText([]byte(s)) }
func (r *Runmode) Get() any           { return *r }

func (r Runmode) MarshalText() ([]byte, error) {
	return []byte(r.String()), nil
}

func (r *Runmode) UnmarshalText(text []byte) error {
	if r == nil {
		return ErrUnmarshalNilRunmode
	}
	if r.unmarshalText(text) ||
		r.unmarshalText(bytes.ToLower(text)) {
		return nil
	}
	return fmt.Errorf("unrecognized runmode: %q", text)
}

func (r *Runmode) unmarshalText(text []byte) bool {
	switch string(text) {
	case "dev":
		*r = RunmodeDev
	case "prod", "":
		*r = RunmodeProd
	default:
		return false
	}
	return true
}
