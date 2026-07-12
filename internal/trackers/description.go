// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package trackers

import (
	"context"

	"github.com/autobrr/upbrr/internal/config"
	"github.com/autobrr/upbrr/internal/services/db"
	"github.com/autobrr/upbrr/pkg/api"
)

// DescriptionRequest supplies one tracker's description-building inputs.
type DescriptionRequest struct {
	Tracker       string
	Meta          api.PreparedMetadata
	TrackerConfig config.TrackerConfig
	AppConfig     config.Config
	Logger        api.Logger
	Repo          db.MetadataRepository
	Assets        *DescriptionAssets
}

// DescriptionResult contains rendered description text and optional diagnostics.
type DescriptionResult struct {
	Group       string
	Description string
}

// DescriptionBuilder renders a tracker-specific upload description.
type DescriptionBuilder interface {
	BuildDescription(ctx context.Context, req DescriptionRequest) (DescriptionResult, error)
}
