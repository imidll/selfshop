package logger

import (
	"context"
	"log/slog"
)

type dualHandler struct {
	normal slog.Handler
	always slog.Handler // unsampled
}

func NewDualHandler(
	normal, always slog.Handler,
) *dualHandler {
	return &dualHandler{normal: normal, always: always}
}

const AlwaysKey = "__always"

func (h *dualHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.normal.Enabled(ctx, l) || h.always.Enabled(ctx, l)
}

func (h *dualHandler) Handle(ctx context.Context, r slog.Record) error {
	if hasAlwaysFlag(r) {
		return h.always.Handle(ctx, r)
	}
	return h.normal.Handle(ctx, r)
}

func (h *dualHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &dualHandler{
		normal: h.normal.WithAttrs(attrs),
		always: h.always.WithAttrs(attrs),
	}
}

func (h *dualHandler) WithGroup(name string) slog.Handler {
	return &dualHandler{
		normal: h.normal.WithGroup(name),
		always: h.always.WithGroup(name),
	}
}

func hasAlwaysFlag(r slog.Record) bool {
	found := false
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == AlwaysKey &&
			a.Value.Equal(slog.BoolValue(true)) {
			found = true
			return false
		}
		return true
	})
	return found
}
