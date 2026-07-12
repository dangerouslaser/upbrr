// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package bhd

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/autobrr/upbrr/internal/config"
	"github.com/autobrr/upbrr/internal/trackers"
	"github.com/autobrr/upbrr/pkg/api"
)

type Definition struct{}

func New() *Definition {
	return &Definition{}
}

func (d *Definition) Name() string {
	return "BHD"
}

func (d *Definition) MetadataPolicy() *trackers.TrackerMetadataPolicy {
	return &trackers.TrackerMetadataPolicy{RequireKnownCategory: true, Requirements: []trackers.MetadataRequirement{{Scope: trackers.MetadataScopeMovie, AnyOf: []trackers.MetadataField{trackers.MetadataFieldIMDB}}}}
}

func (d *Definition) UploadArtifactPolicy() *trackers.UploadArtifactPolicy {
	return &trackers.UploadArtifactPolicy{Source: "BHD"}
}

func (d *Definition) AudioPolicy() *trackers.AudioPolicy {
	return &trackers.AudioPolicy{BlockEnglishOriginalWithForeign: true}
}

func (d *Definition) DupePolicy() *trackers.DupePolicy {
	return &trackers.DupePolicy{MatchAggregateSize: true, NormalizeDDPlusName: true, SDMatchesHD: true, CompareDVDResolution: true, AllowSizeVariance1080: true}
}

func (d *Definition) BannedGroups() []string {
	return []string{
		"Sicario", "TOMMY", "x0r", "nikt0", "FGT", "d3g", "MeGusta", "YIFY", "tigole", "TEKNO3D",
		"C4K", "RARBG", "4K4U", "EASports", "ReaLHD", "Telly", "AOC", "WKS", "SasukeducK", "CRUCiBLE",
		"iFT", "ProRes", "MezRips", "Flights", "BiTOR", "iVy", "QxR", "SyncUP", "OFT", "TGS",
	}
}

func (d *Definition) DataLookupConfigured(cfg config.Config) bool {
	entry, _ := bhdConfig(cfg)
	return len(strings.TrimSpace(entry.APIKey)) >= minDataTokenLength && len(strings.TrimSpace(entry.BhdRSSKey)) >= minDataTokenLength
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

	var err error
	var assets trackers.DescriptionAssets
	if req.Assets != nil {
		assets = *req.Assets
	} else {
		assets, err = trackers.ResolveDescriptionAssets(ctx, req.Tracker, req.Meta, req.Repo, req.Logger)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return trackers.DescriptionResult{}, fmt.Errorf("trackers: %w", err)
			}
			if req.Logger != nil {
				req.Logger.Warnf("trackers: BHD description assets failed: %v", err)
			}
			assets = trackers.DescriptionAssets{}
		}
	}

	description := buildDescription(req.Meta, req.AppConfig, assets)
	return trackers.DescriptionResult{
		Group:       "bhd",
		Description: description,
	}, nil
}
