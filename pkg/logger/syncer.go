package logger

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
)

//go:generate mockgen -destination=mocks/syncer.go -package=loggermocks . SafeSyncer
type SafeSyncer interface {
	Sync() error
}

func SafeSync(l SafeSyncer) error {
	if l == nil {
		return nil
	}

	err := l.Sync()
	if err == nil {
		return nil
	}

	var pe *os.PathError
	if !errors.As(err, &pe) {
		return fmt.Errorf("sync: unexpected error: %w", err)
	}

	switch {
	case errors.Is(pe.Err, syscall.EINVAL),
		errors.Is(pe.Err, syscall.ENOTTY):
		return nil
	}
	msg := strings.ToLower(pe.Err.Error())

	if strings.Contains(msg, "invalid argument") ||
		strings.Contains(msg, "inappropriate ioctl") {
		return nil
	}
	return fmt.Errorf("sync: %w", err)
}
