// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package btn

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
	return "BTN"
}

func (d *Definition) MetadataPolicy() *trackers.TrackerMetadataPolicy {
	return &trackers.TrackerMetadataPolicy{
		RequireKnownCategory: true,
		Requirements: []trackers.MetadataRequirement{
			{Scope: trackers.MetadataScopeTV, AnyOf: []trackers.MetadataField{trackers.MetadataFieldIMDB, trackers.MetadataFieldTVDB}},
		},
	}
}

func (d *Definition) UploadArtifactPolicy() *trackers.UploadArtifactPolicy {
	return &trackers.UploadArtifactPolicy{Source: "BTN", RequireAnnounce: true}
}

func (d *Definition) DataLookupConfigured(cfg config.Config) bool {
	return len(config.ResolveBTNAPIToken(cfg)) >= 25
}

func (d *Definition) DataLookupPolicy() *trackers.DataLookupPolicy {
	return &trackers.DataLookupPolicy{DeferWhenCollectingImages: true}
}

func (d *Definition) BannedGroups() []string {
	return []string{
		"3LTON", "4yEo", "7VFr33104D", "AFG", "AniHLS", "AnimeRG", "AniURL", "DeadFish", "ELiTE", "eSc",
		"EVO", "FGT", "FUM", "GalaxyTV", "GRANiTEN", "HAiKU", "Hi10", "ION10", "JFF", "JIVE", "LOAD", "MeGusta",
		"mSD", "NhaNc3", "NOIVTC", "PHOENiX", "PlaySD", "playXD", "Pr1M371M3", "RAPiDCOWS", "REsuRRecTioN", "RMTeam",
		"ROBOTS", "RUBiK", "SPASM", "Telly", "TM", "URANiME", "ViSiON", "W45Ps", "xRed", "XS", "ZKBL", "ZmN", "ZMNT", "[Oj]",
	}
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

	assets, err := trackers.ResolveDescriptionAssets(ctx, req.Tracker, req.Meta, req.Repo, req.Logger)
	if err != nil {
		assets = trackers.DescriptionAssets{}
	}

	description := strings.TrimSpace(assets.Description)
	if description == "" {
		description = strings.TrimSpace(req.Meta.DescriptionOverride)
	}
	if description == "" {
		description = "No description provided."
	}

	return trackers.DescriptionResult{
		Group:       "btn",
		Description: description,
	}, nil
}

func validateBTNRequest(req trackers.UploadRequest) error {
	if !strings.EqualFold(strings.TrimSpace(req.Meta.ExternalIDs.Category), "TV") && !strings.EqualFold(strings.TrimSpace(req.Meta.MediaInfoCategory), "TV") {
		return errors.New("trackers: BTN only supports TV uploads")
	}
	if strings.TrimSpace(config.ResolveBTNAPIToken(req.AppConfig)) == "" {
		return errors.New("trackers: BTN requires trackers.BTN.api_key")
	}
	return nil
}

var btnInternalGroups = []string{
	"BTW",
	"ESPNtb",
	"HiSD",
	"HRiP",
	"iPRiP",
	"iT00NZ",
	"JJ",
	"LoTV",
	"NTb",
	"PreBS",
	"RAWR",
	"TTVa",
	"TVSmash",
}

func isBTNInternalGroup(meta api.PreparedMetadata) bool {
	if strings.TrimSpace(meta.Tag) == "" {
		return false
	}

	group := strings.ToLower(strings.TrimPrefix(meta.Tag, "-"))
	for _, value := range btnInternalGroups {
		if strings.ToLower(value) == group {
			return true
		}
	}
	return false
}
