package logger_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/imidll/selfshop/pkg/logger"
)

func TestFormat_String(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name   string
		format logger.Format
		want   string
	}{
		{
			name:   "auto",
			format: logger.FormatAuto,
			want:   "auto",
		},
		{
			name:   "json",
			format: logger.FormatJSON,
			want:   "json",
		},
		{
			name:   "console",
			format: logger.FormatConsole,
			want:   "console",
		},
		{
			name:   "unknown",
			format: logger.Format(42),
			want:   "unknown(42)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			got := tc.format.String()

			// Assert
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFormat_MarshalText(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name   string
		format logger.Format
		want   []byte
	}{
		{
			name:   "auto",
			format: logger.FormatAuto,
			want:   []byte("auto"),
		},
		{
			name:   "json",
			format: logger.FormatJSON,
			want:   []byte("json"),
		},
		{
			name:   "console",
			format: logger.FormatConsole,
			want:   []byte("console"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			got, err := tc.format.MarshalText()

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFormat_UnmarshalText(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name    string
		text    string
		want    logger.Format
		wantErr string
	}{
		{
			name: "auto",
			text: "auto",
			want: logger.FormatAuto,
		},
		{
			name: "json",
			text: "json",
			want: logger.FormatJSON,
		},
		{
			name: "console",
			text: "console",
			want: logger.FormatConsole,
		},
		{
			name: "empty means json",
			text: "",
			want: logger.FormatJSON,
		},
		{
			name: "uppercase auto",
			text: "AUTO",
			want: logger.FormatAuto,
		},
		{
			name: "uppercase json",
			text: "JSON",
			want: logger.FormatJSON,
		},
		{
			name: "uppercase console",
			text: "CONSOLE",
			want: logger.FormatConsole,
		},
		{
			name: "mixed case auto",
			text: "AuTo",
			want: logger.FormatAuto,
		},
		{
			name: "mixed case json",
			text: "JsOn",
			want: logger.FormatJSON,
		},
		{
			name: "mixed case console",
			text: "CoNsOlE",
			want: logger.FormatConsole,
		},
		{
			name:    "unknown format",
			text:    "text",
			wantErr: `unrecognized log format: "text"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var got logger.Format

			// Act
			err := got.UnmarshalText([]byte(tc.text))

			// Assert
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestFormat_UnmarshalText_NilReceiver(t *testing.T) {
	// Arrange
	var format *logger.Format

	// Act
	err := format.UnmarshalText([]byte("json"))

	// Assert
	require.ErrorIs(t, err, logger.ErrUnmarshalNilFormat)
}

func TestFormat_Set(t *testing.T) {
	// Arrange
	var format logger.Format

	// Act
	err := format.Set("console")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, logger.FormatConsole, format)
}

func TestFormat_Get(t *testing.T) {
	// Arrange
	format := logger.FormatJSON

	// Act
	got := format.Get()

	// Assert
	assert.Equal(t, logger.FormatJSON, got)
}

func TestParseFormat(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name    string
		text    string
		want    logger.Format
		wantErr string
	}{
		{
			name: "auto",
			text: "auto",
			want: logger.FormatAuto,
		},
		{
			name: "json",
			text: "json",
			want: logger.FormatJSON,
		},
		{
			name: "console",
			text: "console",
			want: logger.FormatConsole,
		},
		{
			name: "case insensitive",
			text: "JSON",
			want: logger.FormatJSON,
		},
		{
			name:    "unknown",
			text:    "xml",
			wantErr: `unrecognized log format: "xml"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			got, err := logger.ParseFormat(tc.text)

			// Assert
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
