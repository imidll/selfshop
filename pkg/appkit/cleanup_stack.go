package appkit

import (
	"context"
	"errors"
	"fmt"
	"slices"
)

type CleanupStack struct {
	items []cleanupItem
}

type cleanupItem struct {
	name string
	fn   func(context.Context) error
}

func (c *CleanupStack) Add(
	name string, fn func(context.Context) error,
) {
	if fn == nil {
		return
	}
	c.items = append(c.items, cleanupItem{name: name, fn: fn})
}

func (c *CleanupStack) Finalize(ctx context.Context) error {
	var err error
	for _, item := range slices.Backward(c.items) {
		if itemErr := item.fn(ctx); itemErr != nil {
			err = errors.Join(
				err,
				fmt.Errorf("%s: %w", item.name, itemErr),
			)
		}
	}
	return err
}
