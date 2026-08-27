package appkit_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/imidll/selfshop/pkg/appkit"
)

func TestRunmode_String(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name string
		mode appkit.Runmode
		want string
	}{
		{
			name: "dev",
			mode: appkit.RunmodeDev,
			want: "dev",
		},
		{
			name: "prod",
			mode: appkit.RunmodeProd,
			want: "prod",
		},
		{
			name: "unknown",
			mode: appkit.Runmode(42),
			want: "unknown(42)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			got := tc.mode.String()

			// Assert
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestRunmode_MarshalText(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name string
		mode appkit.Runmode
		want []byte
	}{
		{
			name: "dev",
			mode: appkit.RunmodeDev,
			want: []byte("dev"),
		},
		{
			name: "prod",
			mode: appkit.RunmodeProd,
			want: []byte("prod"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			got, err := tc.mode.MarshalText()

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestRunmode_UnmarshalText(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name    string
		text    string
		want    appkit.Runmode
		wantErr string
	}{
		{
			name: "dev",
			text: "dev",
			want: appkit.RunmodeDev,
		},
		{
			name: "prod",
			text: "prod",
			want: appkit.RunmodeProd,
		},
		{
			name: "empty means prod",
			text: "",
			want: appkit.RunmodeProd,
		},
		{
			name: "uppercase dev",
			text: "DEV",
			want: appkit.RunmodeDev,
		},
		{
			name: "uppercase prod",
			text: "PROD",
			want: appkit.RunmodeProd,
		},
		{
			name: "mixed case dev",
			text: "DeV",
			want: appkit.RunmodeDev,
		},
		{
			name:    "unknown",
			text:    "test",
			wantErr: `unrecognized runmode: "test"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var got appkit.Runmode

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

func TestRunmode_UnmarshalText_NilReceiver(t *testing.T) {
	// Arrange
	var mode *appkit.Runmode

	// Act
	err := mode.UnmarshalText([]byte("dev"))

	// Assert
	require.ErrorIs(t, err, appkit.ErrUnmarshalNilRunmode)
}

func TestRunmode_Set(t *testing.T) {
	// Arrange
	var mode appkit.Runmode

	// Act
	err := mode.Set("dev")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, appkit.RunmodeDev, mode)
}

func TestRunmode_Get(t *testing.T) {
	// Arrange
	mode := appkit.RunmodeDev

	// Act
	got := mode.Get()

	// Assert
	assert.Equal(t, appkit.RunmodeDev, got)
}

func TestParseRunmode(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name    string
		text    string
		want    appkit.Runmode
		wantErr string
	}{
		{
			name: "dev",
			text: "dev",
			want: appkit.RunmodeDev,
		},
		{
			name: "prod",
			text: "prod",
			want: appkit.RunmodeProd,
		},
		{
			name: "empty means prod",
			text: "",
			want: appkit.RunmodeProd,
		},
		{
			name: "uppercase dev",
			text: "DEV",
			want: appkit.RunmodeDev,
		},
		{
			name: "uppercase prod",
			text: "PROD",
			want: appkit.RunmodeProd,
		},
		{
			name: "mixed case",
			text: "PrOd",
			want: appkit.RunmodeProd,
		},
		{
			name:    "unknown",
			text:    "test",
			wantErr: `unrecognized runmode: "test"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			got, err := appkit.ParseRunmode(tc.text)

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
