// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package unit3d

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/autobrr/upbrr/internal/config"
	"github.com/autobrr/upbrr/internal/trackers"
	trackerdata "github.com/autobrr/upbrr/internal/trackers/data"
	"github.com/autobrr/upbrr/pkg/api"
)

type dupeSearcher struct {
	cfg     config.Config
	tracker *trackerdata.Client
}

func (d *Definition) NewDupeSearcher(cfg config.Config, httpClient *http.Client, logger api.Logger) trackers.DupeSearcher {
	return &dupeSearcher{cfg: cfg, tracker: trackerdata.NewClient(cfg, logger, httpClient)}
}

func (s *dupeSearcher) Search(ctx context.Context, meta api.PreparedMetadata, tracker string) ([]api.DupeEntry, []string, error) {
	if strings.TrimSpace(trackerdata.TrackerAPIKey(s.cfg, tracker)) == "" {
		return nil, []string{"skip: missing api_key for tracker"}, nil
	}
	params := buildDupeSearchParams(meta, tracker)
	if len(params) == 0 {
		return nil, []string{"missing required metadata for dupe search"}, nil
	}
	entries, warning, err := s.tracker.SearchTorrents(ctx, tracker, params, strings.TrimSpace(meta.DiscType) != "")
	if err != nil {
		return nil, nil, fmt.Errorf("dupechecking: %w", err)
	}
	if warning != "" {
		return entries, []string{warning}, nil
	}
	return entries, nil, nil
}
