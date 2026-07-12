// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package azfamily

import (
	"context"
	"errors"
	"fmt"

	"github.com/autobrr/upbrr/internal/trackers"
	"github.com/autobrr/upbrr/pkg/api"
)

type Definition struct {
	site siteDefinition
}

func New(name string) *Definition {
	return &Definition{site: siteFor(name)}
}

func (d *Definition) Name() string {
	return d.site.Name
}

func (d *Definition) UploadArtifactPolicy() *trackers.UploadArtifactPolicy {
	return &trackers.UploadArtifactPolicy{
		Source:          d.site.SourceFlag,
		DefaultAnnounce: d.site.DefaultAnnounceURL,
	}
}

func (d *Definition) MetadataPolicy() *trackers.TrackerMetadataPolicy {
	return &trackers.TrackerMetadataPolicy{RequireKnownCategory: true, Requirements: []trackers.MetadataRequirement{
		{Scope: trackers.MetadataScopeMovie, AnyOf: []trackers.MetadataField{trackers.MetadataFieldTMDBIDOnly, trackers.MetadataFieldIMDBIDOnly}},
		{Scope: trackers.MetadataScopeTV, AnyOf: []trackers.MetadataField{trackers.MetadataFieldTMDBIDOnly, trackers.MetadataFieldIMDBIDOnly, trackers.MetadataFieldTVDBIDOnly}},
	}}
}

func (d *Definition) BannedGroups() []string {
	if d.site.Name != "PHD" {
		return nil
	}
	return []string{
		"RARBG", "STUTTERSHIT", "LiGaS", "DDR", "Zeus", "TBS", "SWTYBLZ", "EASports", "C4K", "d3g", "MeGusta", "YTS", "YIFY", "Tigole", "x0r", "nikt0", "NhaNc3", "PRoDJi", "RDN", "SANTi", "FaNGDiNG0", "FRDS", "HD2DVD", "HDTime", "iPlanet", "KiNGDOM", "Leffe", "4K4U", "Xiaomi", "VisionXpert", "WKS",
	}
}

func (d *Definition) Upload(ctx context.Context, req trackers.UploadRequest) (api.UploadSummary, error) {
	return upload(ctx, applyTrackerConfig(d.site, req.TrackerConfig), req)
}

func (d *Definition) BuildUploadDryRun(ctx context.Context, req trackers.UploadRequest) (api.TrackerDryRunEntry, error) {
	return buildUploadDryRun(ctx, applyTrackerConfig(d.site, req.TrackerConfig), req)
}

func (d *Definition) BuildDescription(ctx context.Context, req trackers.DescriptionRequest) (trackers.DescriptionResult, error) {
	select {
	case <-ctx.Done():
		return trackers.DescriptionResult{}, fmt.Errorf("context canceled: %w", ctx.Err())
	default:
	}

	assets, err := trackers.ResolveDescriptionAssets(ctx, req.Tracker, req.Meta, req.Repo, req.Logger)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return trackers.DescriptionResult{}, fmt.Errorf("trackers: %w", err)
		}
		assets = trackers.DescriptionAssets{}
	}

	description := buildDescription(assets.Description)
	return trackers.DescriptionResult{
		Group:       "azfamily",
		Description: description,
	}, nil
}
