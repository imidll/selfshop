package logger

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDualHandler(t *testing.T) {
	// Arrange
	normal := newSpyHandler()
	always := newSpyHandler()

	// Act
	got := NewDualHandler(normal, always)

	// Assert
	require.NotNil(t, got)
	assert.Same(t, normal, got.normal)
	assert.Same(t, always, got.always)
}

func TestDualHandler_Enabled(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name         string
		normalLevels map[slog.Level]bool
		alwaysLevels map[slog.Level]bool
		level        slog.Level
		want         bool
	}{
		{
			name:         "normal enabled",
			normalLevels: map[slog.Level]bool{slog.LevelInfo: true},
			alwaysLevels: map[slog.Level]bool{slog.LevelInfo: false},
			level:        slog.LevelInfo,
			want:         true,
		},
		{
			name:         "always enabled",
			normalLevels: map[slog.Level]bool{slog.LevelInfo: false},
			alwaysLevels: map[slog.Level]bool{slog.LevelInfo: true},
			level:        slog.LevelInfo,
			want:         true,
		},
		{
			name:         "both enabled",
			normalLevels: map[slog.Level]bool{slog.LevelInfo: true},
			alwaysLevels: map[slog.Level]bool{slog.LevelInfo: true},
			level:        slog.LevelInfo,
			want:         true,
		},
		{
			name:         "both disabled",
			normalLevels: map[slog.Level]bool{slog.LevelInfo: false},
			alwaysLevels: map[slog.Level]bool{slog.LevelInfo: false},
			level:        slog.LevelInfo,
			want:         false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			normal := newSpyHandler()
			always := newSpyHandler()

			normal.enabled = tc.normalLevels
			always.enabled = tc.alwaysLevels

			handler := NewDualHandler(normal, always)

			// Act
			got := handler.Enabled(context.Background(), tc.level)

			// Assert
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDualHandler_Handle(t *testing.T) {
	// Arrange
	errNormal := errors.New("normal")
	errAlways := errors.New("always")

	testCases := [...]struct {
		name        string
		recordAttrs []slog.Attr
		wantHandler string
		wantErr     error
	}{
		{
			name:        "normal",
			recordAttrs: nil,
			wantHandler: "normal",
			wantErr:     errNormal,
		},
		{
			name:        "always",
			recordAttrs: []slog.Attr{slog.Bool(AlwaysKey, true)},
			wantHandler: "always",
			wantErr:     errAlways,
		},
		{
			name: "always flag after other attrs",
			recordAttrs: []slog.Attr{
				slog.String("message", "test"),
				slog.Int("attempt", 1),
				slog.Bool(AlwaysKey, true),
			},
			wantHandler: "always",
			wantErr:     errAlways,
		},
		{
			name:        "always false",
			recordAttrs: []slog.Attr{slog.Bool(AlwaysKey, false)},
			wantHandler: "normal",
			wantErr:     errNormal,
		},
		{
			name:        "always non-bool",
			recordAttrs: []slog.Attr{slog.String(AlwaysKey, "true")},
			wantHandler: "normal",
			wantErr:     errNormal,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			normal := newSpyHandler()
			always := newSpyHandler()

			normal.handleErr = errNormal
			always.handleErr = errAlways

			handler := NewDualHandler(normal, always)

			record := slog.NewRecord(time.Time{}, slog.LevelInfo, "test", 0)
			record.AddAttrs(tc.recordAttrs...)

			// Act
			err := handler.Handle(context.Background(), record)

			// Assert
			require.ErrorIs(t, err, tc.wantErr)

			switch tc.wantHandler {
			case "normal":
				assert.Equal(t, 1, normal.handleCalls)
				assert.Equal(t, 0, always.handleCalls)
			case "always":
				assert.Equal(t, 0, normal.handleCalls)
				assert.Equal(t, 1, always.handleCalls)
			}
		})
	}
}

func TestDualHandler_WithAttrs(t *testing.T) {
	// Arrange
	normal := newSpyHandler()
	always := newSpyHandler()

	handler := NewDualHandler(normal, always)

	attrs := []slog.Attr{
		slog.String("service", "api"),
		slog.Int("version", 2),
	}

	// Act
	got := handler.WithAttrs(attrs)

	// Assert
	require.IsType(t, &dualHandler{}, got)

	gotDual, ok := got.(*dualHandler)
	require.True(t, ok)

	assert.Same(t, normal.withAttrsResult, gotDual.normal)
	assert.Same(t, always.withAttrsResult, gotDual.always)

	assert.Equal(t, 1, normal.withAttrsCalls)
	assert.Equal(t, 1, always.withAttrsCalls)
	assert.Equal(t, attrs, normal.withAttrsAttrs)
	assert.Equal(t, attrs, always.withAttrsAttrs)
}

func TestDualHandler_WithGroup(t *testing.T) {
	// Arrange
	normal := newSpyHandler()
	always := newSpyHandler()

	handler := NewDualHandler(normal, always)

	// Act
	got := handler.WithGroup("request")

	// Assert
	require.IsType(t, &dualHandler{}, got)

	gotDual, ok := got.(*dualHandler)
	require.True(t, ok)

	assert.Same(t, normal.withGroupResult, gotDual.normal)
	assert.Same(t, always.withGroupResult, gotDual.always)

	assert.Equal(t, 1, normal.withGroupCalls)
	assert.Equal(t, 1, always.withGroupCalls)
	assert.Equal(t, "request", normal.withGroupName)
	assert.Equal(t, "request", always.withGroupName)
}

func TestHasAlwaysFlag(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name  string
		attrs []slog.Attr
		want  bool
	}{
		{
			name:  "flag is true",
			attrs: []slog.Attr{slog.Bool(AlwaysKey, true)},
			want:  true,
		},
		{
			name:  "flag is false",
			attrs: []slog.Attr{slog.Bool(AlwaysKey, false)},
			want:  false,
		},
		{
			name:  "flag is absent",
			attrs: []slog.Attr{slog.String("message", "test")},
			want:  false,
		},
		{
			name:  "same key with wrong value",
			attrs: []slog.Attr{slog.String(AlwaysKey, "true")},
			want:  false,
		},
		{
			name: "true flag after other attrs",
			attrs: []slog.Attr{
				slog.String("message", "test"),
				slog.Int("id", 42),
				slog.Bool(AlwaysKey, true),
			},
			want: true,
		},
		{
			name: "true flag among false flags",
			attrs: []slog.Attr{
				slog.Bool(AlwaysKey, false),
				slog.Bool(AlwaysKey, true),
			},
			want: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			record := slog.NewRecord(time.Time{}, slog.LevelInfo, "test", 0)
			record.AddAttrs(tc.attrs...)

			// Act
			got := hasAlwaysFlag(record)

			// Assert
			assert.Equal(t, tc.want, got)
		})
	}
}

type spyHandler struct {
	enabled map[slog.Level]bool

	handleCalls int
	handleErr   error

	withAttrsCalls  int
	withAttrsAttrs  []slog.Attr
	withAttrsResult slog.Handler

	withGroupCalls  int
	withGroupName   string
	withGroupResult slog.Handler
}

func newSpyHandler() *spyHandler {
	handler := &spyHandler{
		enabled: make(map[slog.Level]bool),
	}

	handler.withAttrsResult = handler
	handler.withGroupResult = handler

	return handler
}

func (h *spyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return h.enabled[level]
}

func (h *spyHandler) Handle(_ context.Context, _ slog.Record) error {
	h.handleCalls++
	return h.handleErr
}

func (h *spyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	h.withAttrsCalls++
	h.withAttrsAttrs = attrs
	return h.withAttrsResult
}

func (h *spyHandler) WithGroup(name string) slog.Handler {
	h.withGroupCalls++
	h.withGroupName = name
	return h.withGroupResult
}
