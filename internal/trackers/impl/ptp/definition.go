// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package ptp

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/autobrr/upbrr/internal/config"
	"github.com/autobrr/upbrr/internal/trackers"
	"github.com/autobrr/upbrr/pkg/api"
)

type Definition struct{}

func New() *Definition {
	return &Definition{}
}

func (d *Definition) Name() string {
	return "PTP"
}

func (d *Definition) MetadataPolicy() *trackers.TrackerMetadataPolicy {
	return &trackers.TrackerMetadataPolicy{Requirements: []trackers.MetadataRequirement{{Scope: trackers.MetadataScopeAny, AnyOf: []trackers.MetadataField{trackers.MetadataFieldIMDBIDOnly}, Severity: api.RuleFailureSeverityWarning}}}
}

func (d *Definition) UploadArtifactPolicy() *trackers.UploadArtifactPolicy {
	return &trackers.UploadArtifactPolicy{Source: "PTP"}
}

func (d *Definition) DataLookupConfigured(cfg config.Config) bool {
	for name, entry := range cfg.Trackers.Trackers {
		if strings.EqualFold(strings.TrimSpace(name), "PTP") {
			return strings.TrimSpace(entry.PTPAPIUser) != "" && strings.TrimSpace(entry.PTPAPIKey) != ""
		}
	}
	return false
}

func (d *Definition) DataLookupPolicy() *trackers.DataLookupPolicy {
	return &trackers.DataLookupPolicy{Cooldown: time.Minute}
}

func (d *Definition) BannedGroups() []string {
	return []string{"aXXo", "BMDru", "BRrip", "CM8", "CrEwSaDe", "CTFOH", "d3g", "DNL", "FaNGDiNG0", "HD2DVD", "HDT", "HDTime", "ION10", "iPlanet", "KiNGDOM", "mHD", "mSD", "nHD", "nikt0", "nSD", "NhaNc3", "OFT", "PRODJi", "SANTi", "SPiRiT", "STUTTERSHIT", "ViSION", "VXT", "WAF", "x0r", "YIFY", "LAMA", "WORLD"}
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

	assets := trackers.DescriptionAssets{}
	if req.Assets != nil {
		assets = *req.Assets
	} else {
		resolvedAssets, err := trackers.ResolveDescriptionAssets(ctx, req.Tracker, req.Meta, req.Repo, req.Logger)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return trackers.DescriptionResult{}, fmt.Errorf("trackers: %w", err)
			}
			if req.Logger != nil {
				req.Logger.Warnf("trackers: PTP description assets failed: %v", err)
			}
		} else {
			assets = resolvedAssets
		}
	}

	description := strings.TrimSpace(assets.Description)
	if !assets.Final {
		description = buildDescription(req.Meta, req.TrackerConfig, req.AppConfig, assets)
	}
	if strings.TrimSpace(description) == "" && req.Logger != nil {
		req.Logger.Infof("trackers: PTP preparation description empty")
	}

	return trackers.DescriptionResult{
		Group:       "ptp",
		Description: description,
	}, nil
}
