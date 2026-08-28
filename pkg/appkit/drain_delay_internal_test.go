package appkit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDrainDelay_Duration(t *testing.T) {
	// Arrange
	delay := DrainDelay(1500 * time.Millisecond)

	// Act
	got := delay.Duration()

	// Assert
	assert.Equal(t, 1500*time.Millisecond, got)
}

func TestDrainDelay_Wait(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name    string
		delay   DrainDelay
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "timer completes",
			delay:   0,
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name:    "context canceled",
			delay:   DrainDelay(time.Hour),
			ctx:     canceledContext(),
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			err := tc.delay.Wait(tc.ctx)

			// Assert
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, context.Canceled)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestDrainDelay_Wait_ContextDeadlineExceeded(t *testing.T) {
	// Arrange
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	delay := DrainDelay(time.Hour)

	// Act
	err := delay.Wait(ctx)

	// Assert
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, "drain wait: context canceled", err.Error())
}

func TestDrainDelay_String(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name  string
		delay DrainDelay
		want  string
	}{
		{
			name:  "zero",
			delay: 0,
			want:  "0s",
		},
		{
			name:  "milliseconds",
			delay: DrainDelay(1500 * time.Millisecond),
			want:  "1.5s",
		},
		{
			name:  "seconds",
			delay: DrainDelay(5 * time.Second),
			want:  "5s",
		},
		{
			name:  "negative",
			delay: DrainDelay(-2 * time.Second),
			want:  "-2s",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			got := tc.delay.String()

			// Assert
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDrainDelay_Set(t *testing.T) {
	// Arrange
	var delay DrainDelay

	// Act
	err := delay.Set("2.5s")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, DrainDelay(2500*time.Millisecond), delay)
}

func TestDrainDelay_Set_Invalid(t *testing.T) {
	// Arrange
	var delay DrainDelay

	// Act
	err := delay.Set("invalid")

	// Assert
	require.EqualError(
		t,
		err,
		`invalid drain delay "invalid": time: invalid duration "invalid"`,
	)
}

func TestDrainDelay_Get(t *testing.T) {
	// Arrange
	delay := DrainDelay(5 * time.Second)

	// Act
	got := delay.Get()

	// Assert
	assert.Equal(t, DrainDelay(5*time.Second), got)
}

func TestDrainDelay_MarshalText(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name  string
		delay DrainDelay
		want  []byte
	}{
		{
			name:  "zero",
			delay: 0,
			want:  []byte("0s"),
		},
		{
			name:  "duration",
			delay: DrainDelay(1500 * time.Millisecond),
			want:  []byte("1.5s"),
		},
		{
			name:  "negative",
			delay: DrainDelay(-2 * time.Second),
			want:  []byte("-2s"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			got, err := tc.delay.MarshalText()

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDrainDelay_UnmarshalText(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name    string
		text    string
		want    DrainDelay
		wantErr string
	}{
		{
			name: "duration",
			text: "5s",
			want: DrainDelay(5 * time.Second),
		},
		{
			name: "milliseconds",
			text: "1500ms",
			want: DrainDelay(1500 * time.Millisecond),
		},
		{
			name: "composite duration",
			text: "1h2m3s",
			want: DrainDelay(time.Hour + 2*time.Minute + 3*time.Second),
		},
		{
			name: "negative duration",
			text: "-2s",
			want: DrainDelay(-2 * time.Second),
		},
		{
			name: "empty means zero",
			text: "",
			want: 0,
		},
		{
			name:    "invalid",
			text:    "invalid",
			wantErr: `invalid drain delay "invalid": time: invalid duration "invalid"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			delay := DrainDelay(42 * time.Second)

			// Act
			err := delay.UnmarshalText([]byte(tc.text))

			// Assert
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				assert.Equal(t, DrainDelay(42*time.Second), delay)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, delay)
		})
	}
}

func TestDrainDelay_UnmarshalText_NilReceiver(t *testing.T) {
	// Arrange
	var delay *DrainDelay

	// Act
	err := delay.UnmarshalText([]byte("5s"))

	// Assert
	require.ErrorIs(t, err, ErrUnmarshalNilDrainDelay)
}

func TestDrainDelay_UnmarshalText_InvalidDoesNotModifyValue(t *testing.T) {
	// Arrange
	delay := DrainDelay(5 * time.Second)

	// Act
	err := delay.UnmarshalText([]byte("invalid"))

	// Assert
	require.Error(t, err)
	assert.Equal(t, DrainDelay(5*time.Second), delay)
}

func TestDrainDelay_Set_Empty(t *testing.T) {
	// Arrange
	delay := DrainDelay(5 * time.Second)

	// Act
	err := delay.Set("")

	// Assert
	require.NoError(t, err)
	assert.Zero(t, delay)
}

func TestDrainDelay_MarshalUnmarshalText(t *testing.T) {
	// Arrange
	want := DrainDelay(2500 * time.Millisecond)

	// Act
	data, err := want.MarshalText()
	require.NoError(t, err)

	var got DrainDelay
	err = got.UnmarshalText(data)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func canceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}
