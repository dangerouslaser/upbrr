// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package tvc

import (
	"context"
	"strings"

	"github.com/autobrr/upbrr/internal/trackers"
	"github.com/autobrr/upbrr/pkg/api"
)

type definition struct{}

func New() trackers.Definition  { return definition{} }
func (definition) Name() string { return "TVC" }

func (definition) Search(_ context.Context, meta api.PreparedMetadata, _ string) ([]api.DupeEntry, []string, error) {
	resolution := strings.ToLower(strings.TrimSpace(meta.Release.Resolution))
	if strings.Contains(resolution, "2160") || strings.EqualFold(meta.Type, "REMUX") || strings.TrimSpace(meta.DiscType) != "" {
		return nil, []string{"skip: TVC disallows UHD/disc/remux content"}, nil
	}
	return nil, []string{"skip: TVC dupe search currently unavailable; manual check required"}, nil
}

func (definition) Upload(ctx context.Context, req trackers.UploadRequest) (api.UploadSummary, error) {
	return upload(ctx, req)
}

func (definition) BuildUploadDryRun(ctx context.Context, req trackers.UploadRequest) (api.TrackerDryRunEntry, error) {
	return buildUploadDryRun(ctx, req)
}

func (definition) BuildDescription(ctx context.Context, req trackers.DescriptionRequest) (trackers.DescriptionResult, error) {
	assets, err := trackers.ResolveDescriptionAssets(ctx, req.Tracker, req.Meta, req.Repo, req.Logger)
	if err != nil {
		assets = trackers.DescriptionAssets{}
	}
	description := buildDescription(req.Meta, req.TrackerConfig, assets)
	return trackers.DescriptionResult{Group: "tvc", Description: description}, nil
}
