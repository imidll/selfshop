package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/imidll/selfshop/pkg/appkit"
	"github.com/imidll/selfshop/pkg/config"
	"github.com/imidll/selfshop/pkg/logger"
)

type RootConfig struct {
	App appkit.Config `koanf:"app"`
	Log logger.Config `koanf:"log"`
}

var (
	appcfg *RootConfig
	applog *slog.Logger
)

func main() {
	var err error
	appcfg, err = config.New[RootConfig]("INIT", defaultValues)
	require(err, "config")

	var syncer logger.SafeSyncer
	defer func() {
		require(logger.SafeSync(syncer), "logger")
	}()
	applog, syncer = logger.New(
		appcfg.Log,
		appcfg.App.IsDevmod(), buildAttrs,
	)

	applog.With(logger.AlwaysKey, true).Info(
		"Hi!",
		"app_name", appcfg.App.Name,
		"run_mode", appcfg.App.Runmode,
	)
	appkit.ExitOnError(run, os.Exit)
}

func run() error { return nil }

func require(err error, msg string) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "fatal: %v\n", fmt.Errorf("%s: %w", msg, err))
	os.Exit(1)
}
