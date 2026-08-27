package logger

import (
	"bytes"
	"fmt"
)

var ErrUnmarshalNilFormat = fmt.Errorf("can't unmarshal a nil format")

type Format int8

const (
	FormatAuto Format = iota - 1
	FormatJSON
	FormatConsole

	_minFormat = FormatAuto
	_maxFormat = FormatConsole

	InvalidFormat = _maxFormat + 1
)

func ParseFormat(text string) (Format, error) {
	var f Format
	err := f.UnmarshalText([]byte(text))
	return f, err
}

func (f Format) String() string {
	switch f {
	case FormatConsole:
		return "console"
	case FormatAuto:
		return "auto"
	case FormatJSON:
		return "json"
	default:
		return fmt.Sprintf("unknown(%d)", f)
	}
}

func (f *Format) Set(s string) error { return f.UnmarshalText([]byte(s)) }
func (f *Format) Get() any           { return *f }

func (f Format) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

func (f *Format) UnmarshalText(text []byte) error {
	if f == nil {
		return ErrUnmarshalNilFormat
	}
	if f.unmarshalText(text) ||
		f.unmarshalText(bytes.ToLower(text)) {
		return nil
	}
	return fmt.Errorf("unrecognized log format: %q", text)
}

func (f *Format) unmarshalText(text []byte) bool {
	switch string(text) {
	case "auto":
		*f = FormatAuto
	case "json", "":
		*f = FormatJSON
	case "console":
		*f = FormatConsole
	default:
		return false
	}
	return true
}
