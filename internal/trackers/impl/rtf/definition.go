// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package rtf

import (
	"context"

	"github.com/autobrr/upbrr/internal/trackers"
	"github.com/autobrr/upbrr/pkg/api"
)

type Definition struct{}

func New() *Definition          { return &Definition{} }
func (Definition) Name() string { return "RTF" }

func (Definition) DupePolicy() *trackers.DupePolicy {
	return &trackers.DupePolicy{ContainsFilenameMatch: true}
}

func (Definition) Upload(ctx context.Context, req trackers.UploadRequest) (api.UploadSummary, error) {
	return upload(ctx, req)
}

func (Definition) BuildUploadDryRun(ctx context.Context, req trackers.UploadRequest) (api.TrackerDryRunEntry, error) {
	return buildUploadDryRun(ctx, req)
}

func (Definition) BuildDescription(ctx context.Context, req trackers.DescriptionRequest) (trackers.DescriptionResult, error) {
	var (
		err    error
		assets trackers.DescriptionAssets
	)
	if req.Assets != nil {
		assets = *req.Assets
	} else {
		assets, err = trackers.ResolveDescriptionAssets(ctx, req.Tracker, req.Meta, req.Repo, req.Logger)
		if err != nil {
			assets = trackers.DescriptionAssets{}
		}
	}
	description := buildDescription(assets)
	return trackers.DescriptionResult{Group: "rtf", Description: description}, nil
}
