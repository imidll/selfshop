package appkit

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanupStack_Add(t *testing.T) {
	// Arrange
	var stack CleanupStack

	fn := func(context.Context) error {
		return nil
	}

	// Act
	stack.Add("cleanup", fn)

	// Assert
	require.Len(t, stack.items, 1)
	assert.Equal(t, "cleanup", stack.items[0].name)
	assert.NotNil(t, stack.items[0].fn)
}

func TestCleanupStack_Add_NilFunction(t *testing.T) {
	// Arrange
	var stack CleanupStack

	// Act
	stack.Add("cleanup", nil)

	// Assert
	assert.Empty(t, stack.items)
}

func TestCleanupStack_Add_Multiple(t *testing.T) {
	// Arrange
	var stack CleanupStack

	// Act
	stack.Add("first", func(context.Context) error {
		return nil
	})
	stack.Add("second", func(context.Context) error {
		return nil
	})

	// Assert
	require.Len(t, stack.items, 2)
	assert.Equal(t, "first", stack.items[0].name)
	assert.Equal(t, "second", stack.items[1].name)
}

func TestCleanupStack_Finalize_Empty(t *testing.T) {
	// Arrange
	var stack CleanupStack

	// Act
	err := stack.Finalize(context.Background())

	// Assert
	require.NoError(t, err)
}

func TestCleanupStack_Finalize_ExecutesInReverseOrder(t *testing.T) {
	// Arrange
	var stack CleanupStack
	var got []string

	stack.Add("first", func(context.Context) error {
		got = append(got, "first")
		return nil
	})
	stack.Add("second", func(context.Context) error {
		got = append(got, "second")
		return nil
	})
	stack.Add("third", func(context.Context) error {
		got = append(got, "third")
		return nil
	})

	// Act
	err := stack.Finalize(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{
		"third",
		"second",
		"first",
	}, got)
}

func TestCleanupStack_Finalize_PassesContext(t *testing.T) {
	// Arrange
	var stack CleanupStack

	type contextKey string
	const key contextKey = "cleanup"

	wantCtxValue := "value"
	ctx := context.WithValue(context.Background(), key, wantCtxValue)

	var gotCtxValue string

	stack.Add("cleanup", func(ctx context.Context) error {
		value, ok := ctx.Value(key).(string)
		require.True(t, ok)

		gotCtxValue = value
		return nil
	})

	// Act
	err := stack.Finalize(ctx)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, wantCtxValue, gotCtxValue)
}

func TestCleanupStack_Finalize_ReturnsNilWhenAllSucceed(t *testing.T) {
	// Arrange
	var stack CleanupStack
	var calls int

	stack.Add("first", func(context.Context) error {
		calls++
		return nil
	})
	stack.Add("second", func(context.Context) error {
		calls++
		return nil
	})

	// Act
	err := stack.Finalize(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
}

func TestCleanupStack_Finalize_WrapsErrorWithName(t *testing.T) {
	// Arrange
	var stack CleanupStack

	cleanupErr := errors.New("close failed")

	stack.Add("database", func(context.Context) error {
		return cleanupErr
	})

	// Act
	err := stack.Finalize(context.Background())

	// Assert
	require.Error(t, err)
	require.EqualError(t, err, "database: close failed")
	require.ErrorIs(t, err, cleanupErr)
}

func TestCleanupStack_Finalize_JoinsMultipleErrors(t *testing.T) {
	// Arrange
	var stack CleanupStack

	firstErr := errors.New("first failed")
	secondErr := errors.New("second failed")
	thirdErr := errors.New("third failed")

	stack.Add("first", func(context.Context) error {
		return firstErr
	})
	stack.Add("second", func(context.Context) error {
		return secondErr
	})
	stack.Add("third", func(context.Context) error {
		return thirdErr
	})

	// Act
	err := stack.Finalize(context.Background())

	// Assert
	require.Error(t, err)

	require.ErrorIs(t, err, firstErr)
	require.ErrorIs(t, err, secondErr)
	require.ErrorIs(t, err, thirdErr)

	assert.Equal(
		t,
		"third: third failed\nsecond: second failed\nfirst: first failed",
		err.Error(),
	)
}

func TestCleanupStack_Finalize_ContinuesAfterError(t *testing.T) {
	// Arrange
	var stack CleanupStack
	var got []string

	firstErr := errors.New("first failed")

	stack.Add("first", func(context.Context) error {
		got = append(got, "first")
		return firstErr
	})
	stack.Add("second", func(context.Context) error {
		got = append(got, "second")
		return nil
	})
	stack.Add("third", func(context.Context) error {
		got = append(got, "third")
		return errors.New("third failed")
	})

	// Act
	err := stack.Finalize(context.Background())

	// Assert
	require.Error(t, err)

	assert.Equal(t, []string{
		"third",
		"second",
		"first",
	}, got)

	require.ErrorIs(t, err, firstErr)
	require.Contains(t, err.Error(), "third: third failed")
	require.Contains(t, err.Error(), "first: first failed")
}

func TestCleanupStack_Finalize_ExecutesAllItemsWhenErrorsOccur(t *testing.T) {
	// Arrange
	var stack CleanupStack
	var calls int

	stack.Add("first", func(context.Context) error {
		calls++
		return errors.New("first failed")
	})
	stack.Add("second", func(context.Context) error {
		calls++
		return nil
	})
	stack.Add("third", func(context.Context) error {
		calls++
		return errors.New("third failed")
	})

	// Act
	err := stack.Finalize(context.Background())

	// Assert
	require.Error(t, err)
	assert.Equal(t, 3, calls)
}
