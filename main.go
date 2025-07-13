package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var applyOrder = []string{
	"Namespace",
	"ResourceQuota",
	"LimitRange",
	"PodSecurityPolicy",
	"Secret",
	"ConfigMap",
	"StorageClass",
	"PersistentVolume",
	"PersistentVolumeClaim",
	"ServiceAccount",
	"CustomResourceDefinition",
	"ClusterRole",
	"ClusterRoleBinding",
	"Role",
	"RoleBinding",
	"Service",
	"DaemonSet",
	"Deployment",
	"StatefulSet",
	"Job",
	"CronJob",
	"Ingress",
	"APIService",
	"MutatingWebhookConfiguration",
	"ValidatingWebhookConfiguration",
	"AdmissionConfiguration",
}

func main() {
	slog.Info("SimpleSync", "args", os.Args)

	if len(os.Args) < 2 {
		slog.Error("missing required argument")
		os.Exit(1)
	}

	repository := os.Args[1]

	ctx, cancel := context.WithCancel(context.Background())
	ticker := time.NewTicker(10 * time.Second)
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	s, err := NewSyncer(repository)
	if err != nil {
		slog.Error("failed to initialize syncer", "err", err)
		os.Exit(1)
	}

	s.cloneRepo(ctx)

	for {
		select {
		case sig := <-sigs:
			slog.Info("os signals", "sigs", sig)

			cancel()
			s.Close()

			return
		case t := <-ticker.C:
			slog.Info("ticking", "tick", t)

			if err := s.pull(ctx); err != nil {
				slog.Error("failed to pull repository", "err", err)
				return
			}

			if err := s.loadManifests(ctx); err != nil {
				slog.Error("failed to load manifests", "err", err)
				return

			}

			if err := s.applyManifests(ctx); err != nil {
				slog.Error("failed to apply manifests", "err", err)
				return

			}

		}
	}

}
