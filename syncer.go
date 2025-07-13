package main

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

const defaultManifestPriority = 1000

type Syncer struct {
	Repository string
	TempFolder string
	Manifests  []Manifest
}

func NewSyncer(repository string) (*Syncer, error) {
	tmpFolder, err := os.MkdirTemp("", "repo-*")
	if err != nil {
		return nil, err
	}

	slog.Info("temp folder created", "folder", tmpFolder)

	s := &Syncer{
		Repository: repository,
		TempFolder: tmpFolder,
		Manifests:  nil,
	}

	return s, nil
}

func (s *Syncer) Close() error {
	if s.TempFolder != "" {
		return os.RemoveAll(s.TempFolder)
	}
	return nil
}

func (s *Syncer) cloneRepo(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", s.Repository, s.TempFolder)
	if err := cmd.Run(); err != nil {
		slog.Error("failed to clone repository", "err", err)
		return err
	}

	return nil
}

func (s *Syncer) pull(ctx context.Context) error {
	slog.Info("pulling git updates")
	cmd := exec.CommandContext(ctx, "git", "pull")
	cmd.Dir = s.TempFolder

	if err := cmd.Run(); err != nil {
		slog.Error("failed to pull repository", "err", err)
		return fmt.Errorf("git pull failed: %w", err)
	}

	return nil
}

func (s *Syncer) sortManifests() {
	kindPriority := make(map[string]int, len(applyOrder))
	for i, kind := range applyOrder {
		kindPriority[kind] = i
	}

	sort.Slice(s.Manifests, func(i, j int) bool {
		return s.getManifestPriority(s.Manifests[i], kindPriority) <
			s.getManifestPriority(s.Manifests[j], kindPriority)
	})
}

func (s *Syncer) getManifestPriority(manifest Manifest, kindPriority map[string]int) int {
	priority, exists := kindPriority[manifest.Kind()]
	if !exists || (priority == 0 && manifest.Kind() != "Namespace") {
		return defaultManifestPriority
	}
	return priority
}

func (s *Syncer) loadManifests(context.Context) error {
	s.Manifests = nil

	err := filepath.WalkDir(s.getManifestsFolder(), func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		var m Manifest
		if err := yaml.Unmarshal(content, &m.Body); err != nil {
			return fmt.Errorf("unmarshalling %s: %w", path, err)
		}

		m.Path = path
		s.Manifests = append(s.Manifests, m)
		return nil
	})
	if err != nil {
		return err
	}

	s.sortManifests()
	slog.Info("loaded manifests", "len", len(s.Manifests))
	return nil
}

func (s *Syncer) applyManifests(ctx context.Context) error {
	for _, m := range s.Manifests {
		slog.Info("applying manifest", "path", m.Path)

		cmd := exec.CommandContext(ctx, "kubectl", "apply", "-f", m.Path)
		if result, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to apply manifest: %s, %w", string(result), err)
		}
	}

	return nil
}

func (s *Syncer) getManifestsFolder() string {
	return filepath.Join(s.TempFolder, "manifests")
}
