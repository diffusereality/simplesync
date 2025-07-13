package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	slog.Info("SimpleSync starting", "args", os.Args)

	if len(os.Args) < 2 {
		slog.Error("usage: simplesync <repository>")
		os.Exit(1)
	}

	if err := run(os.Args[1]); err != nil {
		slog.Error("application failed", "err", err)
		os.Exit(1)
	}
}

func run(repository string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigs)

	s, err := NewSyncer(repository)
	if err != nil {
		return fmt.Errorf("failed to initialize syncer: %w", err)
	}

	defer func() {
		if err := s.cleanup(); err != nil {
			slog.Error("failed to cleanup", "err", err)
		}
	}()

	if err := s.cloneRepo(ctx); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	return s.syncLoop(ctx, ticker.C, sigs)
}
