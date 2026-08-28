package health

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	// Arrange
	checker1 := &testChecker{name: "database", ready: true}
	checker2 := &testChecker{name: "redis", ready: false}

	// Act
	got := New(checker1, checker2)

	// Assert
	require.NotNil(t, got)
	require.Len(t, got.checkers, 2)
	assert.Same(t, checker1, got.checkers[0])
	assert.Same(t, checker2, got.checkers[1])
}

func TestNew_WithoutCheckers(t *testing.T) {
	// Act
	got := New()

	// Assert
	require.NotNil(t, got)
	assert.Empty(t, got.checkers)
}

func TestNew_CopiesCheckers(t *testing.T) {
	// Arrange
	checkers := []Checker{
		&testChecker{name: "database", ready: true},
		&testChecker{name: "redis", ready: true},
	}

	// Act
	readiness := New(checkers...)
	checkers[0] = &testChecker{name: "replaced", ready: false}

	// Assert
	require.Len(t, readiness.checkers, 2)
	assert.Equal(t, "database", readiness.checkers[0].Named())
	assert.Equal(t, "redis", readiness.checkers[1].Named())
}

func TestReadiness_MarkNotReady(t *testing.T) {
	// Arrange
	readiness := New(
		&testChecker{name: "database", ready: true},
	)

	require.True(t, readiness.Ready())

	// Act
	readiness.MarkNotReady()

	// Assert
	assert.False(t, readiness.Ready())
}

func TestReadiness_Register(t *testing.T) {
	// Arrange
	readiness := New()

	checker1 := &testChecker{name: "database", ready: true}
	checker2 := &testChecker{name: "redis", ready: true}

	// Act
	readiness.Register(checker1, checker2)

	// Assert
	require.Len(t, readiness.checkers, 2)
	assert.Same(t, checker1, readiness.checkers[0])
	assert.Same(t, checker2, readiness.checkers[1])
}

func TestReadiness_ExistingCheckers(t *testing.T) {
	// Arrange
	readiness := New(
		&testChecker{name: "database", ready: true},
		&testChecker{name: "redis", ready: false},
	)

	// Act
	got := readiness.ExistingCheckers()

	// Assert
	assert.Equal(t, map[string]bool{
		"database": true,
		"redis":    false,
	}, got)
}

func TestReadiness_ExistingCheckers_Empty(t *testing.T) {
	// Arrange
	readiness := New()

	// Act
	got := readiness.ExistingCheckers()

	// Assert
	require.NotNil(t, got)
	assert.Empty(t, got)
}

func TestReadiness_ExistingCheckers_ReturnsCopy(t *testing.T) {
	// Arrange
	readiness := New(
		&testChecker{name: "database", ready: true},
	)

	// Act
	got := readiness.ExistingCheckers()
	got["database"] = false
	got["redis"] = true

	// Assert
	assert.Equal(t, map[string]bool{
		"database": true,
	}, readiness.ExistingCheckers())
}

func TestReadiness_Ready(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name     string
		closed   bool
		checkers []Checker
		want     bool
	}{
		{
			name: "no checkers",
			want: false,
		},
		{
			name: "all checkers ready",
			checkers: []Checker{
				&testChecker{name: "database", ready: true},
				&testChecker{name: "redis", ready: true},
			},
			want: true,
		},
		{
			name: "one checker not ready",
			checkers: []Checker{
				&testChecker{name: "database", ready: true},
				&testChecker{name: "redis", ready: false},
			},
			want: false,
		},
		{
			name:   "closed",
			closed: true,
			checkers: []Checker{
				&testChecker{name: "database", ready: true},
			},
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			readiness := New(tc.checkers...)

			if tc.closed {
				readiness.MarkNotReady()
			}

			// Act
			got := readiness.Ready()

			// Assert
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestReadiness_ReadyHandler(t *testing.T) {
	// Arrange
	testCases := [...]struct {
		name     string
		checkers []Checker
		wantCode int
	}{
		{
			name: "ready",
			checkers: []Checker{
				&testChecker{name: "database", ready: true},
				&testChecker{name: "redis", ready: true},
			},
			wantCode: http.StatusOK,
		},
		{
			name: "not ready",
			checkers: []Checker{
				&testChecker{name: "database", ready: true},
				&testChecker{name: "redis", ready: false},
			},
			wantCode: http.StatusServiceUnavailable,
		},
		{
			name:     "no checkers",
			checkers: nil,
			wantCode: http.StatusServiceUnavailable,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			readiness := New(tc.checkers...)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/ready", nil)

			// Act
			readiness.ReadyHandler(recorder, request)

			// Assert
			assert.Equal(t, tc.wantCode, recorder.Code)
		})
	}
}

func TestReadiness_ReadyHandler_AfterMarkNotReady(t *testing.T) {
	// Arrange
	readiness := New(
		&testChecker{name: "database", ready: true},
	)

	readiness.MarkNotReady()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ready", nil)

	// Act
	readiness.ReadyHandler(recorder, request)

	// Assert
	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}

func TestReadiness_AliveHandler(t *testing.T) {
	// Arrange
	readiness := New()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/live", nil)

	// Act
	readiness.AliveHandler(recorder, request)

	// Assert
	assert.Equal(t, http.StatusOK, recorder.Code)
}

type testChecker struct {
	name  string
	ready bool
}

func (c *testChecker) Named() string { return c.name }
func (c *testChecker) Ready() bool   { return c.ready }
