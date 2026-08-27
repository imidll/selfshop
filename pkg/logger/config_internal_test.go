package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_resolveFormat(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name   string
		format Format
		devmod bool
		want   bool
	}{
		{
			name:   "console",
			format: FormatConsole,
			devmod: false,
			want:   true,
		},
		{
			name:   "console in dev",
			format: FormatConsole,
			devmod: true,
			want:   true,
		},
		{
			name:   "auto in dev",
			format: FormatAuto,
			devmod: true,
			want:   true,
		},
		{
			name:   "auto in prod",
			format: FormatAuto,
			devmod: false,
			want:   false,
		},
		{
			name:   "unknown in dev",
			format: Format(42),
			devmod: true,
			want:   false,
		},
		{
			name:   "unknown in prod",
			format: Format(42),
			devmod: false,
			want:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			config := Config{Format: tc.format}

			// Act
			got := config.resolveFormat(tc.devmod)

			// Assert
			assert.Equal(t, tc.want, got)
		})
	}
}
