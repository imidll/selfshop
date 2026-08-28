package appkit

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var ErrUnmarshalNilDrainDelay = errors.New("can't unmarshal a nil drain delay")

type DrainDelay time.Duration

func (d DrainDelay) Duration() time.Duration { return time.Duration(d) }

func (d DrainDelay) Wait(ctx context.Context) error {
	t := time.NewTimer(time.Duration(d))
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("drain wait: %w", ctx.Err())
	}
}

func (d DrainDelay) String() string { return d.Duration().String() }

func (d *DrainDelay) Set(s string) error { return d.UnmarshalText([]byte(s)) }
func (d *DrainDelay) Get() any           { return *d }

func (d DrainDelay) MarshalText() ([]byte, error) { return []byte(d.String()), nil }

func (d *DrainDelay) UnmarshalText(text []byte) error {
	if d == nil {
		return ErrUnmarshalNilDrainDelay
	}
	if len(text) == 0 {
		*d = 0
		return nil
	}
	parsed, err := time.ParseDuration(string(text))
	if err != nil {
		return fmt.Errorf("invalid drain delay %q: %w", text, err)
	}
	*d = DrainDelay(parsed)
	return nil
}
