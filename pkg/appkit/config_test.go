package appkit_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/imidll/selfshop/pkg/appkit"
)

func TestConfig_IsDevmod(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name    string
		runmode appkit.Runmode
		want    bool
	}{
		{
			name:    "dev",
			runmode: appkit.RunmodeDev,
			want:    true,
		},
		{
			name:    "prod",
			runmode: appkit.RunmodeProd,
			want:    false,
		},
		{
			name:    "unknown",
			runmode: appkit.Runmode(42),
			want:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			config := appkit.Config{Runmode: tc.runmode}

			// Act
			got := config.IsDevmod()

			// Assert
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestShutdownConfig_TotalTimeout(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name       string
		timeout    time.Duration
		drainDelay appkit.DrainDelay
		want       time.Duration
	}{
		{
			name:       "zero values",
			timeout:    0,
			drainDelay: 0,
			want:       0,
		},
		{
			name:       "timeout only",
			timeout:    10 * time.Second,
			drainDelay: 0,
			want:       10 * time.Second,
		},
		{
			name:       "drain delay only",
			timeout:    0,
			drainDelay: appkit.DrainDelay(5 * time.Second),
			want:       5 * time.Second,
		},
		{
			name:       "timeout and drain delay",
			timeout:    30 * time.Second,
			drainDelay: appkit.DrainDelay(5 * time.Second),
			want:       35 * time.Second,
		},
		{
			name:       "milliseconds",
			timeout:    1500 * time.Millisecond,
			drainDelay: appkit.DrainDelay(750 * time.Millisecond),
			want:       2250 * time.Millisecond,
		},
		{
			name:       "negative drain delay",
			timeout:    10 * time.Second,
			drainDelay: appkit.DrainDelay(-2 * time.Second),
			want:       8 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			config := appkit.ShutdownConfig{
				Timeout:    tc.timeout,
				DrainDelay: tc.drainDelay,
			}

			// Act
			got := config.TotalTimeout()

			// Assert
			assert.Equal(t, tc.want, got)
		})
	}
}
