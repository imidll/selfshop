package appkit_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/imidll/selfshop/pkg/appkit"
)

func TestExit_WhenNoError(t *testing.T) {
	// Arrange
	exitCalled := false
	exitFunc := func(code int) { exitCalled = true }

	// Act
	appkit.ExitOnError(func() error { return nil }, exitFunc)

	// Assert
	assert.False(t, exitCalled)
}

func TestExit_WhenError(t *testing.T) {
	// Arrange
	exitCalled := false
	exitCode := 0
	exitFunc := func(code int) { exitCalled = true; exitCode = code }

	// Act
	appkit.ExitOnError(func() error { return errors.New("something went wrong") }, exitFunc)

	// Assert
	assert.Equal(t, 1, exitCode)
	require.True(t, exitCalled)
}
