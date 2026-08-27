package logger

import (
	"log/slog"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
)

func New(c Config, devmod bool, args ...any) (*slog.Logger, SafeSyncer) {
	lvl := zap.NewAtomicLevelAt(c.MinLevel)
	out := zapcore.Lock(os.Stdout)
	enc := newLogEncoder(c.resolveFormat(devmod))

	loCore := zapcore.NewCore(
		enc, out,
		zap.LevelEnablerFunc(func(l zapcore.Level) bool { return lvl.Enabled(l) && l < c.Sampler.AlwaysLevel }),
	)
	hiCore := zapcore.NewCore(
		enc, out,
		zap.LevelEnablerFunc(func(l zapcore.Level) bool { return lvl.Enabled(l) && l >= c.Sampler.AlwaysLevel }),
	)

	alwaysCore := zapcore.NewCore(enc, out, lvl)
	normalCore := zapcore.NewTee(
		hiCore,
		zapcore.NewSamplerWithOptions(
			loCore,
			c.Sampler.Interval,
			c.Sampler.First, c.Sampler.Thereafter,
		),
	)

	return slog.New(&dualHandler{
		normal: zapslog.NewHandler(normalCore), // sampled
		always: zapslog.NewHandler(alwaysCore),
	}).With(args...), alwaysCore
}

func newLogEncoder(devfmt bool) (enc zapcore.Encoder) {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeDuration = zapcore.MillisDurationEncoder
	cfg.EncodeTime = zapcore.RFC3339TimeEncoder
	enc = zapcore.NewJSONEncoder(cfg)
	if devfmt {
		cfg := zap.NewDevelopmentEncoderConfig()
		cfg.EncodeTime = func(t time.Time, e zapcore.PrimitiveArrayEncoder) { e.AppendString(t.Format("15:04:05.000")) }
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		enc = zapcore.NewConsoleEncoder(cfg)
	}
	return enc
}
