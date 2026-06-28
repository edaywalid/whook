// Command whook runs the self-hostable webhook gateway.
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/edaywalid/whook/internal/app"
	"github.com/edaywalid/whook/internal/config"
)

// version is stamped at build time with -ldflags "-X main.version=...". It falls
// back to the module version recorded in the binary for `go install` builds.
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print the whook version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println(buildVersion())
		return
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("starting whook", "version", buildVersion())

	cfg, err := config.Load()
	if err != nil {
		logger.Error("loading config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx, cfg, logger); err != nil {
		logger.Error("gateway exited", "error", err)
		os.Exit(1)
	}
}

// buildVersion returns the stamped version, or the module version from the build
// info when whook was installed with `go install` and no ldflags were set.
func buildVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}
