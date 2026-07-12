// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package unit3d

import (
	"context"

	"github.com/autobrr/upbrr/internal/config"
	"github.com/autobrr/upbrr/internal/trackers"
	"github.com/autobrr/upbrr/pkg/api"
)

// SiteProfile contains optional site-owned Unit3D payload callbacks.
type SiteProfile struct {
	BuildName              func(meta api.PreparedMetadata, cfg config.TrackerConfig) string
	BuildDescription       func(ctx context.Context, meta api.PreparedMetadata, appConfig config.Config, trackerConfig config.TrackerConfig, logger api.Logger, keptDescription string, menuImages []api.ScreenshotImage, screenshots []api.ScreenshotImage) (string, error)
	ResolveKeywords        func(meta api.PreparedMetadata) string
	ResolveTypeID          func(meta api.PreparedMetadata) string
	ResolveResolutionID    func(meta api.PreparedMetadata) string
	ResolveCategoryID      func(meta api.PreparedMetadata) string
	ApplyAdditionalPayload func(req trackers.UploadRequest, data map[string]string)
	FinalizeDescription    func(description string, meta api.PreparedMetadata) string
}

func firstSiteProfile(profiles []SiteProfile) SiteProfile {
	if len(profiles) == 0 {
		return SiteProfile{}
	}
	return profiles[0]
}
