// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package unit3d

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/autobrr/upbrr/internal/config"
	internalerrors "github.com/autobrr/upbrr/internal/errors"
	"github.com/autobrr/upbrr/internal/redaction"
	"github.com/autobrr/upbrr/internal/trackers"
	"github.com/autobrr/upbrr/internal/trackers/ruletypes"
	"github.com/autobrr/upbrr/pkg/api"
)

// IsConfiguredTracker reports whether tracker is either a known Unit3D site or
// a custom config entry whose auth shape identifies the Unit3D protocol.
func IsConfiguredTracker(cfg config.Config, tracker string) bool {
	return IsConfiguredTrackerWithRegistry(cfg, tracker, nil)
}

// IsConfiguredTrackerWithRegistry reports whether tracker is registered as Unit3D or is a custom Unit3D-shaped config entry.
func IsConfiguredTrackerWithRegistry(cfg config.Config, tracker string, registry *trackers.Registry) bool {
	key := strings.ToUpper(strings.TrimSpace(tracker))
	if key == "" {
		return false
	}
	if kind, ok := registry.LookupKind(key); ok {
		return kind == trackers.KindUnit3D
	}
	if trackers.IsUnit3DTracker(key) {
		return true
	}
	if trackers.IsKnownTracker(key) {
		return false
	}
	entry, ok := cfg.Trackers.Trackers[key]
	if !ok {
		for name, candidate := range cfg.Trackers.Trackers {
			if strings.EqualFold(name, key) {
				entry, ok = candidate, true
				break
			}
		}
	}
	if !ok {
		return false
	}
	return strings.TrimSpace(entry.APIKey) != "" &&
		strings.TrimSpace(entry.AnnounceURL) != "" &&
		strings.TrimSpace(entry.Username) == "" &&
		strings.TrimSpace(entry.Password) == "" &&
		strings.TrimSpace(entry.Passkey) == "" &&
		strings.TrimSpace(entry.PTPAPIUser) == "" &&
		strings.TrimSpace(entry.PTPAPIKey) == ""
}

type Definition struct {
	profile Profile
}

// Profile declares Unit3D site identity and default endpoint metadata.
// Runtime tracker config remains authoritative for endpoint overrides.
type Profile struct {
	Name             string
	BaseURL          string
	Site             SiteProfile
	Rules            *ruletypes.RuleSet
	DupePolicy       *trackers.DupePolicy
	UploadArtifact   *trackers.UploadArtifactPolicy
	BannedPolicy     *trackers.BannedGroupPolicy
	BannedGroups     []string
	ImageHost        *trackers.ImageHostPolicy
	ClaimPolicy      *trackers.ClaimPolicy
	DescriptionGroup string
}

func New(name string) *Definition {
	return NewWithProfile(Profile{Name: name})
}

// NewWithProfile constructs a Unit3D definition from an explicitly composed
// site profile.
func NewWithProfile(profile Profile) *Definition {
	profile.Name = strings.ToUpper(strings.TrimSpace(profile.Name))
	profile.BaseURL = strings.TrimSpace(profile.BaseURL)
	profile.DescriptionGroup = strings.ToLower(strings.TrimSpace(profile.DescriptionGroup))
	profile.BannedGroups = append([]string(nil), profile.BannedGroups...)
	return &Definition{profile: profile}
}

func (d *Definition) DescriptionGroup() string { return d.profile.DescriptionGroup }

func (d *Definition) DefaultBaseURL() string { return d.profile.BaseURL }

func (d *Definition) ClaimPolicy() *trackers.ClaimPolicy {
	if d.profile.ClaimPolicy == nil {
		return nil
	}
	policy := *d.profile.ClaimPolicy
	return &policy
}

func (d *Definition) UploadArtifactPolicy() *trackers.UploadArtifactPolicy {
	if d.profile.UploadArtifact == nil {
		return nil
	}
	policy := *d.profile.UploadArtifact
	return &policy
}

func (d *Definition) MetadataPolicy() *trackers.TrackerMetadataPolicy {
	return &trackers.TrackerMetadataPolicy{
		Requirements: []trackers.MetadataRequirement{{Scope: trackers.MetadataScopeAny, AnyOf: []trackers.MetadataField{trackers.MetadataFieldTMDB}}},
	}
}

// Rules declares validation required by every Unit3D upload.
func (d *Definition) Rules() *ruletypes.RuleSet {
	if d.profile.Rules == nil {
		return &ruletypes.RuleSet{RequireValidMISetting: true}
	}
	rules := *d.profile.Rules
	rules.RequireValidMISetting = true
	return &rules
}

func (d *Definition) BannedGroups() []string { return append([]string(nil), d.profile.BannedGroups...) }

func (d *Definition) BannedGroupPolicy() *trackers.BannedGroupPolicy { return d.profile.BannedPolicy }

func (d *Definition) DupePolicy() *trackers.DupePolicy { return d.profile.DupePolicy }

func (d *Definition) ImageHostPolicy() *trackers.ImageHostPolicy { return d.profile.ImageHost }

func (d *Definition) Name() string {
	return d.profile.Name
}

func (d *Definition) TrackerKind() trackers.Kind { return trackers.KindUnit3D }

func (d *Definition) Upload(ctx context.Context, req trackers.UploadRequest) (api.UploadSummary, error) {
	if d.profile.BaseURL != "" || strings.TrimSpace(req.TrackerConfig.URL) != "" {
		if strings.TrimSpace(req.TrackerConfig.URL) == "" {
			req.TrackerConfig.URL = d.profile.BaseURL
		}
		return uploadUnit3D(ctx, req, d.profile.Site)
	}
	select {
	case <-ctx.Done():
		return api.UploadSummary{}, fmt.Errorf("context canceled: %w", ctx.Err())
	default:
	}
	if req.Logger != nil {
		req.Logger.Infof("trackers: %s upload not implemented (unit3d scaffold)", d.profile.Name)
	}
	return api.UploadSummary{}, internalerrors.ErrNotImplemented
}

func (d *Definition) BuildUploadDryRun(ctx context.Context, req trackers.UploadRequest) (api.TrackerDryRunEntry, error) {
	if d.profile.BaseURL != "" || strings.TrimSpace(req.TrackerConfig.URL) != "" {
		if strings.TrimSpace(req.TrackerConfig.URL) == "" {
			req.TrackerConfig.URL = d.profile.BaseURL
		}
		return buildUploadDryRunUnit3D(ctx, req, d.profile.Site)
	}
	select {
	case <-ctx.Done():
		return api.TrackerDryRunEntry{}, fmt.Errorf("context canceled: %w", ctx.Err())
	default:
	}
	if req.Logger != nil {
		req.Logger.Infof("trackers: dry-run decision=not_implemented tracker=%s", d.profile.Name)
	}
	return api.TrackerDryRunEntry{}, internalerrors.ErrNotImplemented
}

func (d *Definition) BuildDescription(ctx context.Context, req trackers.DescriptionRequest) (trackers.DescriptionResult, error) {
	select {
	case <-ctx.Done():
		return trackers.DescriptionResult{}, fmt.Errorf("context canceled: %w", ctx.Err())
	default:
	}
	if req.Logger != nil {
		req.Logger.Debugf("trackers: %s building unit3d description", d.profile.Name)
	}
	var err error
	assets := trackers.DescriptionAssets{}
	if req.Assets != nil {
		assets = *req.Assets
	} else {
		assets, err = trackers.ResolveDescriptionAssets(ctx, req.Tracker, req.Meta, req.Repo, req.Logger)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return trackers.DescriptionResult{}, fmt.Errorf("trackers: %w", err)
			}
			if req.Logger != nil {
				req.Logger.Warnf("trackers: description assets failed tracker=%s err=%s", d.profile.Name, redaction.RedactValue(err.Error(), nil))
			}
			assets = trackers.DescriptionAssets{}
		}
	}
	description := strings.TrimSpace(assets.Description)
	if !assets.Final {
		description, err = buildUnit3DDescription(
			ctx,
			d.profile.Name,
			req.Meta,
			req.AppConfig,
			req.TrackerConfig,
			req.Logger,
			assets.Description,
			assets.MenuImages,
			assets.Screenshots,
			d.profile.Site,
		)
		if err != nil {
			return trackers.DescriptionResult{}, err
		}
	}
	return trackers.DescriptionResult{
		Group:       "unit3d",
		Description: description,
	}, nil
}

func Register(registry *trackers.Registry, trackersList []string) error {
	if registry == nil {
		return nil
	}
	for _, name := range trackersList {
		if err := registry.Register(New(name)); err != nil {
			return fmt.Errorf("trackers: %w", err)
		}
	}
	return nil
}

// RegisterProfiles explicitly registers composed Unit3D site profiles.
func RegisterProfiles(registry *trackers.Registry, profiles []Profile) error {
	if registry == nil {
		return nil
	}
	for _, profile := range profiles {
		if strings.TrimSpace(profile.Name) == "" {
			return errors.New("trackers: unit3d profile has empty name")
		}
		definition := NewWithProfile(profile)
		if err := registry.Register(definition); err != nil {
			return fmt.Errorf("trackers: %w", err)
		}
	}
	return nil
}
