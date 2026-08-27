package logger_test

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/imidll/selfshop/pkg/logger"
	loggermocks "github.com/imidll/selfshop/pkg/logger/mocks"
)

func TestSafeSync(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name    string
		syncErr error
		wantErr string
	}{
		{
			name:    "sync succeeds",
			syncErr: nil,
		},
		{
			name: "EINVAL is ignored",
			syncErr: &os.PathError{
				Op:   "sync",
				Path: "test",
				Err:  syscall.EINVAL,
			},
		},
		{
			name: "ENOTTY is ignored",
			syncErr: &os.PathError{
				Op:   "sync",
				Path: "test",
				Err:  syscall.ENOTTY,
			},
		},
		{
			name: "invalid argument is ignored",
			syncErr: &os.PathError{
				Op:   "sync",
				Path: "test",
				Err:  errors.New("invalid argument"),
			},
		},
		{
			name: "inappropriate ioctl is ignored",
			syncErr: &os.PathError{
				Op:   "sync",
				Path: "test",
				Err:  errors.New("inappropriate ioctl for device"),
			},
		},
		{
			name: "other path error is returned",
			syncErr: &os.PathError{
				Op:   "sync",
				Path: "test",
				Err:  syscall.EIO,
			},
			wantErr: "sync: sync test: input/output error",
		},
		{
			name:    "non path error is wrapped",
			syncErr: errors.New("something went wrong"),
			wantErr: "sync: unexpected error: something went wrong",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			syncer := loggermocks.NewMockSafeSyncer(ctrl)
			syncer.EXPECT().Sync().Return(tc.syncErr)

			// Act
			err := logger.SafeSync(syncer)

			// Assert
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestSafeSync_Nil(t *testing.T) {
	// Act
	err := logger.SafeSync(nil)

	// Assert
	require.NoError(t, err)
}

func TestSafeSync_WrappedErrors(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name string
		err  error
	}{
		{
			name: "wrapped EINVAL",
			err:  fmt.Errorf("wrapped: %w", syscall.EINVAL),
		},
		{
			name: "wrapped ENOTTY",
			err:  fmt.Errorf("wrapped: %w", syscall.ENOTTY),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			syncer := loggermocks.NewMockSafeSyncer(ctrl)
			syncer.EXPECT().Sync().Return(&os.PathError{
				Op:   "sync",
				Path: "test",
				Err:  tc.err,
			})

			// Act
			err := logger.SafeSync(syncer)

			// Assert
			require.NoError(t, err)
		})
	}
}
