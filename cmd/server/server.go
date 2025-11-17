package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"log/slog"

	"github.com/0x2e/fusion/api"
	"github.com/0x2e/fusion/conf"
	"github.com/0x2e/fusion/repo"
	"github.com/0x2e/fusion/service/pull"
)

func main() {
	config, err := conf.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		return
	}

	// Parse log level from config
	logLevel := parseLogLevel(config.LogLevel)

	// Setup logger based on debug mode and log level
	if conf.Debug {
		l := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
		slog.SetDefault(l)

		go func() {
			if err := http.ListenAndServe("localhost:6060", nil); err != nil {
				slog.Error("pprof server", "error", err)
				return
			}
		}()
	} else {
		l := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}))
		slog.SetDefault(l)
	}

	slog.Info("Starting Fusion", "log_level", logLevel.String())
	repo.Init(config.DB)

	go pull.NewPuller(repo.NewFeed(repo.DB), repo.NewItem(repo.DB), config.AutoFetchFullContent).Run()

	api.Run(api.Params{
		Host:                 config.Host,
		Port:                 config.Port,
		PasswordHash:         config.PasswordHash,
		UseSecureCookie:      config.SecureCookie,
		TLSCert:              config.TLSCert,
		TLSKey:               config.TLSKey,
		AutoFetchFullContent: config.AutoFetchFullContent,
	})
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		slog.Warn("Invalid log level, using INFO", "level", level)
		return slog.LevelInfo
	}
}
