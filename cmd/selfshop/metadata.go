package main

import (
	"log/slog"
	"os"
)

// build metadata, overridden via ldflags at build time.
var (
	version = "dev"
	comhash = "undefined"
	buildAt = "none"
)

var buildAttrs = slog.Group(
	"program_info",
	slog.String("build_at", buildAt),
	slog.Int("os_pid", os.Getpid()),
	slog.String("version", version),
	slog.String("comhash", comhash),
)
