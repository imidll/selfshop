package appkit_test

import (
	"testing"

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
