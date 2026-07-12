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
	name             string
	baseURL          string
	uploadArtifact   *trackers.UploadArtifactPolicy
	claimPolicy      *trackers.ClaimPolicy
	descriptionGroup string
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
	return &Definition{name: strings.ToUpper(strings.TrimSpace(name))}
}

// NewWithProfile constructs a Unit3D definition from an explicitly composed
// site profile.
func NewWithProfile(profile Profile) *Definition {
	return &Definition{
		name:             strings.ToUpper(strings.TrimSpace(profile.Name)),
		baseURL:          strings.TrimSpace(profile.BaseURL),
		uploadArtifact:   profile.UploadArtifact,
		claimPolicy:      profile.ClaimPolicy,
		descriptionGroup: strings.ToLower(strings.TrimSpace(profile.DescriptionGroup)),
	}
}

func (d *Definition) DescriptionGroup() string { return d.descriptionGroup }

func (d *Definition) ClaimPolicy() *trackers.ClaimPolicy {
	if d.claimPolicy == nil {
		return nil
	}
	policy := *d.claimPolicy
	return &policy
}

func (d *Definition) UploadArtifactPolicy() *trackers.UploadArtifactPolicy {
	if d.uploadArtifact == nil {
		return nil
	}
	policy := *d.uploadArtifact
	return &policy
}

func (d *Definition) MetadataPolicy() *trackers.TrackerMetadataPolicy {
	return &trackers.TrackerMetadataPolicy{Requirements: []trackers.MetadataRequirement{{Scope: trackers.MetadataScopeAny, AnyOf: []trackers.MetadataField{trackers.MetadataFieldTMDB}}}}
}

// Rules declares validation required by every Unit3D upload.
func (d *Definition) Rules() *ruletypes.RuleSet {
	return &ruletypes.RuleSet{RequireValidMISetting: true}
}

func (d *Definition) Name() string {
	return d.name
}

func (d *Definition) TrackerKind() trackers.Kind { return trackers.KindUnit3D }

func (d *Definition) Upload(ctx context.Context, req trackers.UploadRequest) (api.UploadSummary, error) {
	if d.baseURL != "" || strings.TrimSpace(req.TrackerConfig.URL) != "" {
		if strings.TrimSpace(req.TrackerConfig.URL) == "" {
			req.TrackerConfig.URL = d.baseURL
		}
		return uploadUnit3D(ctx, req)
	}
	select {
	case <-ctx.Done():
		return api.UploadSummary{}, fmt.Errorf("context canceled: %w", ctx.Err())
	default:
	}
	if req.Logger != nil {
		req.Logger.Infof("trackers: %s upload not implemented (unit3d scaffold)", d.name)
	}
	return api.UploadSummary{}, internalerrors.ErrNotImplemented
}

func (d *Definition) BuildUploadDryRun(ctx context.Context, req trackers.UploadRequest) (api.TrackerDryRunEntry, error) {
	if d.baseURL != "" || strings.TrimSpace(req.TrackerConfig.URL) != "" {
		if strings.TrimSpace(req.TrackerConfig.URL) == "" {
			req.TrackerConfig.URL = d.baseURL
		}
		return buildUploadDryRunUnit3D(ctx, req)
	}
	select {
	case <-ctx.Done():
		return api.TrackerDryRunEntry{}, fmt.Errorf("context canceled: %w", ctx.Err())
	default:
	}
	if req.Logger != nil {
		req.Logger.Infof("trackers: dry-run decision=not_implemented tracker=%s", d.name)
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
		req.Logger.Debugf("trackers: %s building unit3d description", d.name)
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
				req.Logger.Warnf("trackers: description assets failed tracker=%s err=%s", d.name, redaction.RedactValue(err.Error(), nil))
			}
			assets = trackers.DescriptionAssets{}
		}
	}
	description := strings.TrimSpace(assets.Description)
	if !assets.Final {
		description, err = buildUnit3DDescription(ctx, d.name, req.Meta, req.AppConfig, req.TrackerConfig, req.Logger, assets.Description, assets.MenuImages, assets.Screenshots)
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
		installSiteProfile(profile.Name, profile.Site)
		definition := NewWithProfile(profile)
		rules := definition.Rules()
		if profile.Rules != nil {
			cloned := *profile.Rules
			cloned.RequireValidMISetting = true
			rules = &cloned
		}
		if err := registry.RegisterDescriptor(trackers.Descriptor{Name: profile.Name, BaseURL: profile.BaseURL, Definition: definition, DupeFactory: definition, Rules: rules, DupePolicy: profile.DupePolicy, UploadArtifact: profile.UploadArtifact, Metadata: definition.MetadataPolicy(), BannedPolicy: profile.BannedPolicy, BannedGroups: append([]string(nil), profile.BannedGroups...), ImageHost: profile.ImageHost, ClaimPolicy: profile.ClaimPolicy, DescriptionGroup: profile.DescriptionGroup}); err != nil {
			return fmt.Errorf("trackers: %w", err)
		}
	}
	return nil
}
