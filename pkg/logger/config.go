package logger

import (
	"time"

	"go.uber.org/zap/zapcore"
)

type Config struct {
	MinLevel zapcore.Level `koanf:"min_level"`
	Format   Format        `koanf:"format"`
	Sampler  SamplerConfig `koanf:"sampler"`
}

type SamplerConfig struct {
	AlwaysLevel zapcore.Level `koanf:"always_level"`
	Interval    time.Duration `koanf:"interval"`
	First       int           `koanf:"first"`
	Thereafter  int           `koanf:"thereafter"`
}

func (c Config) resolveFormat(
	devmod bool,
) bool {
	return c.Format == FormatConsole || (c.Format == FormatAuto && devmod)
}
