// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package ant

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/autobrr/upbrr/internal/trackers"
	"github.com/autobrr/upbrr/internal/trackers/ruletypes"
	"github.com/autobrr/upbrr/pkg/api"
)

type Definition struct{}

func New() *Definition {
	return &Definition{}
}

func (d *Definition) Name() string {
	return "ANT"
}

func (d *Definition) Rules() *ruletypes.RuleSet {
	return &ruletypes.RuleSet{RequireMovieOnly: true}
}

func (d *Definition) ArtifactPolicy() *trackers.ArtifactPolicy {
	return &trackers.ArtifactPolicy{MaxPieceSizeMiB: 128, MaxTorrentBytes: 250 << 10}
}

func (d *Definition) BannedGroups() []string {
	groups := slices.Collect(maps.Keys(antBannedReleaseGroups))
	slices.Sort(groups)
	return groups
}

func (d *Definition) MetadataPolicy() *trackers.TrackerMetadataPolicy {
	return &trackers.TrackerMetadataPolicy{RequireKnownCategory: true, Requirements: []trackers.MetadataRequirement{{Scope: trackers.MetadataScopeMovie, AnyOf: []trackers.MetadataField{trackers.MetadataFieldTMDB}}}}
}

func (d *Definition) UploadArtifactPolicy() *trackers.UploadArtifactPolicy {
	return &trackers.UploadArtifactPolicy{Source: "ANT"}
}

func (d *Definition) DupePolicy() *trackers.DupePolicy {
	return &trackers.DupePolicy{DolbyVisionImpliesHDR: true}
}

func (d *Definition) AudioPolicy() *trackers.AudioPolicy {
	return &trackers.AudioPolicy{AllowedLanguages: []string{"english"}, BlockEnglishOriginalWithForeign: true}
}

func (d *Definition) Upload(ctx context.Context, req trackers.UploadRequest) (api.UploadSummary, error) {
	return upload(ctx, req)
}

func (d *Definition) BuildUploadDryRun(ctx context.Context, req trackers.UploadRequest) (api.TrackerDryRunEntry, error) {
	return buildUploadDryRun(ctx, req)
}

func (d *Definition) BuildDescription(ctx context.Context, req trackers.DescriptionRequest) (trackers.DescriptionResult, error) {
	select {
	case <-ctx.Done():
		return trackers.DescriptionResult{}, fmt.Errorf("context canceled: %w", ctx.Err())
	default:
	}

	assets, err := trackers.ResolveDescriptionAssetsWithPrepared(ctx, req.Tracker, req.Meta, req.Repo, req.Logger, req.Assets)
	if err != nil {
		assets = trackers.DescriptionAssets{}
	}
	assets.Description = trackers.StripDefaultDescriptionSignature(assets.Description)

	description := buildDescription(trackers.UploadRequest{
		Tracker:       req.Tracker,
		Meta:          req.Meta,
		TrackerConfig: req.TrackerConfig,
		AppConfig:     req.AppConfig,
		Logger:        req.Logger,
		Repo:          req.Repo,
	}, assets)

	return trackers.DescriptionResult{
		Group:       "ant",
		Description: strings.TrimSpace(description),
	}, nil
}
