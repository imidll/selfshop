package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/imidll/selfshop/pkg/appkit"
	"github.com/imidll/selfshop/pkg/config"
	"github.com/imidll/selfshop/pkg/health"
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
	notfctx, uncatch := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM, syscall.SIGINT,
	)
	defer uncatch()

	var err error
	appcfg, err = config.New[RootConfig]("INIT", defaultValues)
	require(err, "config")

	var tcplis net.Listener
	tcplis, err = net.Listen("tcp", net.JoinHostPort("::", "8080"))
	require(err, "tcp")

	var syncer logger.SafeSyncer
	defer func() {
		require(logger.SafeSync(syncer), "logger")
	}()
	applog, syncer = logger.New(
		appcfg.Log,
		appcfg.App.IsDevmod(), buildAttrs,
	)

	applog.With(logger.AlwaysKey, true).Info(
		"hi!",
		"app_name", appcfg.App.Name,
		"run_mode", appcfg.App.Runmode,
	)
	appkit.ExitOnError(func() error {
		return run(notfctx, uncatch, tcplis)
	}, os.Exit)
}

func run(
	notfctx context.Context,
	uncatch context.CancelFunc, tcplis net.Listener,
) (
	err error,
) {
	h := health.New()
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health/alive", h.AliveHandler)
	mux.HandleFunc("GET /health/ready", h.ReadyHandler)

	critlog := applog.With(logger.AlwaysKey, true)

	httpsrv := new(http.Server{
		Addr:              tcplis.Addr().String(),
		ReadTimeout:       30 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
		WriteTimeout:      30 * time.Second,
	})
	httpsrv.ErrorLog = slog.NewLogLogger(critlog.Handler(), slog.LevelWarn)
	httpsrv.Handler = mux
	httpsrv.DisableGeneralOptionsHandler = true

	// root context for the entire application
	// it is propagated to long-lived components and canceled during shutdown
	rootctx, abort := context.WithCancelCause(context.Background())
	defer abort(nil)

	g, workctx := errgroup.WithContext(rootctx)

	var cleanup appkit.CleanupStack
	defer func() {
		critlog.Info(
			"shutdown initiated, draining active requests",
			"drain_delay", appcfg.App.Shutdown.DrainDelay,
		)
		// stop intercepting signals so a second SIGINT/SIGTERM
		// terminates the process immediately
		uncatch()
		h.MarkNotReady() // fail readiness checks to drain incoming traffic

		// now tell all long-lived components to stop
		abort(errors.New("application shutdown"))

		downctx, finish := context.WithTimeout(
			context.Background(),
			appcfg.App.Shutdown.TotalTimeout(),
		)
		defer finish()

		// give time for readiness check to propagate
		if waitErr := appcfg.App.Shutdown.DrainDelay.Wait(
			downctx,
		); waitErr != nil {
			err = errors.Join(err, fmt.Errorf("delay: %w", waitErr))
		}

		if err = errors.Join(
			err,
			wraperr(httpsrv.Shutdown(downctx), "http server shutdown"),
			wraperr(cleanup.Finalize(downctx), "cleanup"),
		); err != nil {
			critlog.Error("graceful shutdown failed", "error", err)
		}
		critlog.Info("application exited")
	}()

	g.Go(func() error {
		critlog.Info("HTTP server starting", "addr", httpsrv.Addr)
		if serveErr := httpsrv.Serve(tcplis); serveErr != nil &&
			!errors.Is(serveErr, http.ErrServerClosed) {
			return fmt.Errorf("http serve: %w", serveErr)
		}
		critlog.Info("HTTP server stopped accepting new connections")
		return nil
	})

	select {
	case <-notfctx.Done():
	case <-workctx.Done():
		cause := context.Cause(workctx)
		if cause == nil {
			cause = workctx.Err()
		}
		if !errors.Is(cause, context.Canceled) {
			critlog.Error("critical component failed", "error", cause)
		}
	}
	return err
}

func require(err error, msg string) {
	if err == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "fatal: %v\n", fmt.Errorf("%s: %w", msg, err))
	os.Exit(1)
}

func wraperr(err error, msg string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", msg, err)
}
