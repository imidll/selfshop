package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestNew_ProductionFormat(t *testing.T) {
	// Arrange
	config := Config{
		MinLevel: zapcore.InfoLevel,
		Format:   FormatAuto,
		Sampler: SamplerConfig{
			AlwaysLevel: zapcore.ErrorLevel,
			Interval:    time.Hour,
			First:       1,
			Thereafter:  0,
		},
	}

	output := captureStdout(t, func() {
		logger, syncer := New(config, false)

		logger.Info("hello")

		_ = syncer.Sync()
	})

	// Assert
	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &got))

	assert.Equal(t, "hello", got["msg"])
	assert.Equal(t, "info", strings.ToLower(got["level"].(string)))
}

func TestNew_DevFormat(t *testing.T) {
	// Arrange
	config := Config{
		MinLevel: zapcore.InfoLevel,
		Format:   FormatAuto,
		Sampler: SamplerConfig{
			AlwaysLevel: zapcore.ErrorLevel,
			Interval:    time.Hour,
			First:       1,
			Thereafter:  0,
		},
	}

	output := captureStdout(t, func() {
		logger, syncer := New(config, true)

		logger.Info("hello")

		_ = syncer.Sync()
	})

	// Assert
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "hello")
}

func TestNew_Sampling(t *testing.T) {
	// Arrange
	config := Config{
		MinLevel: zapcore.InfoLevel,
		Format:   FormatJSON,
		Sampler: SamplerConfig{
			AlwaysLevel: zapcore.ErrorLevel,
			Interval:    time.Hour,
			First:       2,
			Thereafter:  0,
		},
	}

	output := captureStdout(t, func() {
		logger, _ := New(config, false)

		logger.Info("sampled")
		logger.Info("sampled")
		logger.Info("sampled")
	})

	// Assert
	lines := nonEmptyLines(output)

	assert.Len(t, lines, 2)
}

func TestNew_AlwaysLevelIsNotSampled(t *testing.T) {
	// Arrange
	config := Config{
		MinLevel: zapcore.InfoLevel,
		Format:   FormatJSON,
		Sampler: SamplerConfig{
			AlwaysLevel: zapcore.ErrorLevel,
			Interval:    time.Hour,
			First:       1,
			Thereafter:  0,
		},
	}

	output := captureStdout(t, func() {
		logger, syncer := New(config, false)

		logger.Error("error-1")
		logger.Error("error-2")

		_ = syncer.Sync()
	})

	// Assert
	lines := nonEmptyLines(output)

	require.Len(t, lines, 2)
	assert.Contains(t, lines[0], "error-1")
	assert.Contains(t, lines[1], "error-2")
}

func TestNew_AlwaysFlag(t *testing.T) {
	// Arrange
	config := Config{
		MinLevel: zapcore.InfoLevel,
		Format:   FormatJSON,
		Sampler: SamplerConfig{
			AlwaysLevel: zapcore.ErrorLevel,
			Interval:    time.Hour,
			First:       1,
			Thereafter:  0,
		},
	}

	output := captureStdout(t, func() {
		logger, syncer := New(config, false)

		logger.Info("always-1", slog.Bool(AlwaysKey, true))
		logger.Info("always-2", slog.Bool(AlwaysKey, true))

		_ = syncer.Sync()
	})

	// Assert
	lines := nonEmptyLines(output)

	require.Len(t, lines, 2)
	assert.Contains(t, lines[0], "always-1")
	assert.Contains(t, lines[1], "always-2")
}

func TestNew_MinLevel(t *testing.T) {
	// Arrange
	config := Config{
		MinLevel: zapcore.ErrorLevel,
		Format:   FormatJSON,
		Sampler: SamplerConfig{
			AlwaysLevel: zapcore.ErrorLevel,
			Interval:    time.Hour,
			First:       1,
			Thereafter:  0,
		},
	}

	output := captureStdout(t, func() {
		logger, syncer := New(config, false)

		logger.Info("ignored")
		logger.Error("error")

		_ = syncer.Sync()
	})

	// Assert
	lines := nonEmptyLines(output)

	require.Len(t, lines, 1)
	assert.NotContains(t, lines[0], "ignored")
	assert.Contains(t, lines[0], "error")
}

func TestNew_WithArgs(t *testing.T) {
	// Arrange
	config := Config{
		MinLevel: zapcore.InfoLevel,
		Format:   FormatJSON,
		Sampler: SamplerConfig{
			AlwaysLevel: zapcore.ErrorLevel,
			Interval:    time.Hour,
			First:       1,
			Thereafter:  0,
		},
	}

	output := captureStdout(t, func() {
		logger, syncer := New(config, false, "service", "api")

		logger.Info("hello")

		_ = syncer.Sync()
	})

	// Assert
	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &got))

	assert.Equal(t, "hello", got["msg"])
	assert.Equal(t, "api", got["service"])
}

func TestNewLogEncoder_Production(t *testing.T) {
	// Arrange
	encoder := newLogEncoder(false)

	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Date(2026, 8, 27, 12, 34, 56, 123_000_000, time.UTC),
		Message: "hello",
	}

	// Act
	buf, err := encoder.EncodeEntry(entry, nil)

	// Assert
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &got))

	assert.Equal(t, "hello", got["msg"])
	assert.Equal(t, "2026-08-27T12:34:56Z", got["ts"])
}

func TestNewLogEncoder_Development(t *testing.T) {
	// Arrange
	encoder := newLogEncoder(true)

	entry := zapcore.Entry{
		Level:   zapcore.InfoLevel,
		Time:    time.Date(2026, 8, 27, 12, 34, 56, 123_000_000, time.UTC),
		Message: "hello",
	}

	// Act
	buf, err := encoder.EncodeEntry(entry, nil)

	// Assert
	require.NoError(t, err)

	output := buf.String()

	assert.Contains(t, output, "12:34:56.123")
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "hello")
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout

	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	fn()

	require.NoError(t, w.Close())

	var buf bytes.Buffer

	_, err = io.Copy(&buf, r)
	require.NoError(t, err)
	require.NoError(t, r.Close())

	return buf.String()
}

func nonEmptyLines(output string) (lines []string) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
