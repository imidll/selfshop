package config

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_DefaultValues(t *testing.T) {
	// Arrange
	type Server struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	}

	type Config struct {
		Name    string        `koanf:"name"`
		Debug   bool          `koanf:"debug"`
		Timeout time.Duration `koanf:"timeout"`
		Server  Server        `koanf:"server"`
	}

	defaults := map[string]any{
		"name":    "selfshop",
		"debug":   true,
		"timeout": "5s",
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	}

	// Act
	got, err := New[Config]("app", defaults)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, "selfshop", got.Name)
	assert.True(t, got.Debug)
	assert.Equal(t, 5*time.Second, got.Timeout)
	assert.Equal(t, "localhost", got.Server.Host)
	assert.Equal(t, 8080, got.Server.Port)
}

func TestNew_EnvironmentOverridesDefaults(t *testing.T) {
	// Arrange
	type Config struct {
		Name  string `koanf:"name"`
		Port  int    `koanf:"port"`
		Debug bool   `koanf:"debug"`
	}

	t.Setenv("APP_NAME", "from-env")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("APP_DEBUG", "true")

	defaults := map[string]any{
		"name":  "from-defaults",
		"port":  8080,
		"debug": false,
	}

	// Act
	got, err := New[Config]("app", defaults)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, "from-env", got.Name)
	assert.Equal(t, 9090, got.Port)
	assert.True(t, got.Debug)
}

func TestNew_EnvironmentNestedKeys(t *testing.T) {
	// Arrange
	type Server struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	}

	type HTTP struct {
		Port int `koanf:"port"`
	}

	type Config struct {
		Server Server `koanf:"server"`
		HTTP   HTTP   `koanf:"http"`
	}

	t.Setenv("APP_SERVER__HOST", "localhost")
	t.Setenv("APP_SERVER__PORT", "8080")
	t.Setenv("APP_HTTP__PORT", "9090")

	// Act
	got, err := New[Config]("app", nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, "localhost", got.Server.Host)
	assert.Equal(t, 8080, got.Server.Port)
	assert.Equal(t, 9090, got.HTTP.Port)
}

func TestNew_EnvironmentPrefixIsCaseInsensitive(t *testing.T) {
	// Arrange
	type Config struct {
		Name string `koanf:"name"`
	}

	t.Setenv("APP_NAME", "test")

	// Act
	got, err := New[Config]("aPp", nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, "test", got.Name)
}

func TestNew_EnvironmentPrefixTrimsTrailingUnderscores(t *testing.T) {
	// Arrange
	type Config struct {
		Name string `koanf:"name"`
	}

	t.Setenv("APP_NAME", "test")

	// Act
	got, err := New[Config]("app___", nil)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, "test", got.Name)
}

func TestNew_UnknownDefaultKey(t *testing.T) {
	// Arrange
	type Config struct {
		Name string `koanf:"name"`
	}

	defaults := map[string]any{
		"name":    "test",
		"unknown": "value",
	}

	// Act
	got, err := New[Config]("app", defaults)

	// Assert
	require.Nil(t, got)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "unknown")
	assert.Contains(t, err.Error(), "invalid keys")
}

func TestNew_UnknownNestedDefaultKey(t *testing.T) {
	// Arrange
	type Server struct {
		Host string `koanf:"host"`
	}

	type Config struct {
		Server Server `koanf:"server"`
	}

	defaults := map[string]any{
		"server.host":    "localhost",
		"server.unknown": "value",
	}

	// Act
	got, err := New[Config]("app", defaults)

	// Assert
	require.Nil(t, got)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "unknown")
	assert.Contains(t, err.Error(), "invalid keys")
}

func TestNew_UnknownEnvironmentKey(t *testing.T) {
	// Arrange
	type Config struct {
		Name string `koanf:"name"`
	}

	t.Setenv("APP_NAME", "test")
	t.Setenv("APP_UNKNOWN", "value")

	// Act
	got, err := New[Config]("app", nil)

	// Assert
	require.Nil(t, got)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "unknown")
	assert.Contains(t, err.Error(), "invalid keys")
}

func TestNew_UnknownNestedEnvironmentKey(t *testing.T) {
	// Arrange
	type Server struct {
		Host string `koanf:"host"`
	}

	type Config struct {
		Server Server `koanf:"server"`
	}

	t.Setenv("APP_SERVER__HOST", "localhost")
	t.Setenv("APP_SERVER__UNKNOWN", "value")

	// Act
	got, err := New[Config]("app", nil)

	// Assert
	require.Nil(t, got)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "unknown")
	assert.Contains(t, err.Error(), "invalid keys")
}

func TestNew_ExcludedField(t *testing.T) {
	// Arrange
	type Config struct {
		Name    string `koanf:"name"`
		Private string `koanf:"-"`
	}

	defaults := map[string]any{"name": "test"}

	// Act
	got, err := New[Config]("app", defaults)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, "test", got.Name)
	assert.Empty(t, got.Private)
}

func TestNew_DurationDecodeHook(t *testing.T) {
	// Arrange
	type Config struct {
		Timeout time.Duration `koanf:"timeout"`
	}

	defaults := map[string]any{"timeout": "5s"}

	// Act
	got, err := New[Config]("app", defaults)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, 5*time.Second, got.Timeout)
}

func TestNew_TimeDecodeHook(t *testing.T) {
	// Arrange
	type Config struct {
		StartAt time.Time `koanf:"start_at"`
	}

	defaults := map[string]any{"start_at": "2026-08-27T12:00:00Z"}

	// Act
	got, err := New[Config]("app", defaults)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(
		t,
		time.Date(2026, 8, 27, 12, 0, 0, 0, time.UTC),
		got.StartAt,
	)
}

func TestNew_InvalidValue(t *testing.T) {
	// Arrange
	type Config struct {
		Port int `koanf:"port"`
	}

	defaults := map[string]any{"port": "not-a-number"}

	// Act
	got, err := New[Config]("app", defaults)

	// Assert
	require.Nil(t, got)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "unmarshal")
}

func TestNew_InvalidTime(t *testing.T) {
	// Arrange
	type Config struct {
		StartAt time.Time `koanf:"start_at"`
	}

	defaults := map[string]any{"start_at": "not-a-time"}

	// Act
	got, err := New[Config]("app", defaults)

	// Assert
	require.Nil(t, got)
	require.Error(t, err)

	assert.Contains(t, err.Error(), "unmarshal")
}

func TestNew_NonStructType(t *testing.T) {
	// Act
	got, err := New[string]("app", nil)

	// Assert
	require.Nil(t, got)
	require.EqualError(t, err, "type parameter T must be a struct, got string")
}

func TestNew_NilType(t *testing.T) {
	// Act
	got, err := New[any]("app", nil)

	// Assert
	require.Nil(t, got)
	require.EqualError(t, err, "type parameter T must be a struct")
}

func TestNew_PointerToStruct(t *testing.T) {
	// Arrange
	type Config struct {
		Name string `koanf:"name"`
	}

	defaults := map[string]any{"name": "test"}

	// Act
	got, err := New[*Config]("app", defaults)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)
	require.NotNil(t, *got)

	assert.Equal(t, "test", (*got).Name)
}

func TestNew_Validator(t *testing.T) {
	// Arrange
	type Config struct {
		Name string `koanf:"name"`
	}

	// Act
	got, err := New[Config]("app", map[string]any{"name": "test"})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, "test", got.Name)
}

func TestNew_ValidatorSuccess(t *testing.T) {
	// Act
	got, err := New[validConfig]("app", map[string]any{
		"name": "test",
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, "test", got.Name)
}

func TestNew_ValidatorError(t *testing.T) {
	// Act
	got, err := New[invalidConfig]("app", map[string]any{
		"name": "test",
	})

	// Assert
	require.Nil(t, got)
	require.EqualError(t, err, "semantic validation: invalid config")
}

func TestEnvKeyTransform(t *testing.T) {
	// Arrange
	transform := envKeyTransform("APP_", delim)

	testCases := [...]struct {
		name  string
		key   string
		value string
		want  string
	}{
		{
			name:  "simple",
			key:   "APP_NAME",
			value: "test",
			want:  "name",
		},
		{
			name:  "nested",
			key:   "APP_SERVER__HOST",
			value: "localhost",
			want:  "server.host",
		},
		{
			name:  "multiple nested levels",
			key:   "APP_SERVER__HTTP__PORT",
			value: "8080",
			want:  "server.http.port",
		},
		{
			name:  "leading underscores",
			key:   "APP___NAME",
			value: "test",
			want:  "name",
		},
		{
			name:  "without prefix",
			key:   "OTHER_NAME",
			value: "test",
			want:  "other_name",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			gotKey, gotValue := transform(tc.key, tc.value)

			// Assert
			assert.Equal(t, tc.want, gotKey)
			assert.Equal(t, tc.value, gotValue)
		})
	}
}

func TestDecoderConfig(t *testing.T) {
	// Act
	got := decoderConfig()

	// Assert
	require.NotNil(t, got)

	assert.False(t, got.IgnoreUntaggedFields)
	assert.True(t, got.ErrorUnused)
	assert.False(t, got.WeaklyTypedInput)
	assert.NotNil(t, got.DecodeHook)
}

type validConfig struct {
	Name string `koanf:"name"`
}

func (validConfig) Validate() error {
	return nil
}

type invalidConfig struct {
	Name string `koanf:"name"`
}

func (invalidConfig) Validate() error {
	return errors.New("invalid config")
}

func TestValidatorInterface(t *testing.T) {
	// Arrange
	var validator Validator = &validConfig{}

	// Act
	err := validator.Validate()

	// Assert
	require.NoError(t, err)
}
