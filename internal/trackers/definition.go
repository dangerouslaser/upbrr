// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package trackers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/autobrr/upbrr/internal/config"
	"github.com/autobrr/upbrr/internal/services/db"
	"github.com/autobrr/upbrr/internal/trackers/datatypes"
	"github.com/autobrr/upbrr/internal/trackers/ruletypes"
	"github.com/autobrr/upbrr/pkg/api"
)

// ErrSubmitted2FARejected marks tracker auth failure after a supplied manual 2FA code was rejected.
var ErrSubmitted2FARejected = errors.New("trackers: submitted 2FA rejected")

// AuthResolutionError reports tracker-owned remote auth classification to the generic coordinator.
type AuthResolutionError struct {
	Reason           string
	AuthRequired     bool
	ConfirmedInvalid bool
	Transient        bool
	Err              error
}

func (e *AuthResolutionError) Error() string {
	if e == nil {
		return "tracker auth resolution failed"
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Reason
}

func (e *AuthResolutionError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// UploadRequest supplies the prepared release and runtime dependencies needed by
// one tracker definition to build or submit an upload.
type UploadRequest struct {
	// Tracker is the normalized tracker name receiving the upload.
	Tracker string
	// Meta is the prepared release snapshot used throughout this upload attempt.
	Meta api.PreparedMetadata
	// TrackerConfig is the effective configuration for Tracker.
	TrackerConfig config.TrackerConfig
	// AppConfig is the effective application configuration snapshot.
	AppConfig config.Config
	// Logger receives tracker workflow progress and diagnostics.
	Logger api.Logger
	// Repo provides persisted metadata and image selections.
	Repo db.MetadataRepository
	// Images uploads description images when tracker policy requires rehosting.
	Images api.ImageHostingService
	// Assets contains pre-resolved description text and selected images, when available.
	Assets *DescriptionAssets
	// Registry exposes capabilities owned by other tracker definitions.
	Registry *Registry
}

// Definition is the required upload contract for a registered tracker.
type Definition interface {
	Name() string
	Upload(ctx context.Context, req UploadRequest) (api.UploadSummary, error)
}

// KindProvider declares a tracker's protocol family.
type KindProvider interface {
	TrackerKind() Kind
}

// BaseURLProvider declares a tracker's default endpoint.
type BaseURLProvider interface {
	DefaultBaseURL() string
}

// LocalizedMetadataProvider declares a locale consumed by tracker-owned naming or description behavior.
type LocalizedMetadataProvider interface {
	LocalizedMetadataLocale() string
}

// DescriptionGroupProvider declares a tracker-specific description override group.
type DescriptionGroupProvider interface {
	DescriptionGroup() string
}

// AuthSessionResolver validates or refreshes tracker-owned auth material.
type AuthSessionResolver func(context.Context, config.TrackerConfig, string, api.TrackerAuthLoginRequest) error

// AuthSessionProvider declares tracker-owned remote auth behavior.
type AuthSessionProvider interface {
	AuthSessionResolver() AuthSessionResolver
}

// AuthCapabilityProvider declares tracker-owned auth support metadata.
type AuthCapabilityProvider interface {
	AuthCapability() api.TrackerAuthCapability
}

// DupeSearcher searches a tracker for releases matching prepared metadata.
type DupeSearcher interface {
	Search(ctx context.Context, meta api.PreparedMetadata, tracker string) ([]api.DupeEntry, []string, error)
}

// DupeSearcherFactory constructs a tracker-owned searcher from runtime deps.
type DupeSearcherFactory interface {
	NewDupeSearcher(cfg config.Config, httpClient *http.Client, logger api.Logger) DupeSearcher
}

// RuleProvider declares tracker-owned validation rules.
type RuleProvider interface {
	Rules() *ruletypes.RuleSet
}

// ArtifactPolicy declares tracker-owned torrent artifact constraints.
type ArtifactPolicy struct {
	// MaxPieceSizeMiB is the largest permitted torrent piece size; zero imposes no limit.
	MaxPieceSizeMiB int
	// MaxTorrentBytes is the largest permitted encoded torrent size; zero imposes no limit.
	MaxTorrentBytes int64
}

// ArtifactPolicyProvider declares tracker-owned torrent artifact policy.
type ArtifactPolicyProvider interface {
	ArtifactPolicy() *ArtifactPolicy
}

// DataLookupRequest contains tracker metadata lookup inputs.
type DataLookupRequest struct {
	// TrackerID is the tracker-side torrent or release identifier when already known.
	TrackerID string
	// Meta is the prepared release whose tracker metadata is requested.
	Meta api.PreparedMetadata
	// SearchName overrides the release name used for tracker search when non-empty.
	SearchName string
	// OnlyID limits lookup work to resolving tracker identity where supported.
	OnlyID bool
	// KeepImages requests preservation of images found in tracker descriptions.
	KeepImages bool
}

// DataLookup resolves tracker-owned metadata for a release.
type DataLookup interface {
	Lookup(ctx context.Context, req DataLookupRequest) (datatypes.Result, error)
}

// DataLookupFactory constructs a tracker-owned lookup from runtime deps.
type DataLookupFactory interface {
	NewDataLookup(cfg config.Config, httpClient *http.Client, logger api.Logger) DataLookup
}

// DataLookupConfigProvider validates tracker-data lookup credentials.
type DataLookupConfigProvider interface {
	DataLookupConfigured(cfg config.Config) bool
}

// DataLookupPolicy declares tracker-specific lookup orchestration behavior.
type DataLookupPolicy struct {
	// Cooldown is the minimum delay applied around tracker lookup operations.
	Cooldown time.Duration
	// DeferWhenCollectingImages postpones lookup while the caller is still collecting images.
	DeferWhenCollectingImages bool
}

// DataLookupPolicyProvider declares tracker-owned lookup orchestration policy.
type DataLookupPolicyProvider interface {
	DataLookupPolicy() *DataLookupPolicy
}

// BannedGroupsProvider declares tracker-owned static banned release groups.
type BannedGroupsProvider interface {
	BannedGroups() []string
}

// BannedGroupPolicy declares a tracker-owned dynamic blacklist source.
type BannedGroupPolicy struct {
	// EndpointPath is appended to the configured tracker base URL.
	EndpointPath string
	// DefaultEndpoint is used when tracker configuration supplies no base URL.
	DefaultEndpoint string
	// TRaSHGuideURL supplies an optional external banned-group source.
	TRaSHGuideURL string
	// RequireAPIKey disables remote refresh when no API key is configured.
	RequireAPIKey bool
	// RawAPIKeyFallback allows the configured APIKey field when no specialized key exists.
	RawAPIKeyFallback bool
}

// BannedGroupPolicyProvider declares dynamic banned-group retrieval behavior.
type BannedGroupPolicyProvider interface {
	BannedGroupPolicy() *BannedGroupPolicy
}

// MetadataPolicyProvider declares tracker-owned metadata requirements.
type MetadataPolicyProvider interface {
	MetadataPolicy() *TrackerMetadataPolicy
}

// UploadArtifactPolicy declares tracker torrent personalization fields.
type UploadArtifactPolicy struct {
	// Source replaces the torrent info dictionary's private-tracker source field.
	Source string
	// DefaultAnnounce is used when tracker configuration has no announce URL.
	DefaultAnnounce string
	// UseMyAnnounce selects the tracker configuration's personal announce URL.
	UseMyAnnounce bool
	// RequireAnnounce prevents artifact preparation without an announce URL.
	RequireAnnounce bool
}

// UploadArtifactPolicyProvider declares tracker-owned personalization policy.
type UploadArtifactPolicyProvider interface {
	UploadArtifactPolicy() *UploadArtifactPolicy
}

// DupePolicy declares tracker-specific duplicate comparison semantics.
type DupePolicy struct {
	DolbyVisionImpliesHDR           bool
	MatchAggregateSize              bool
	ContainsFilenameMatch           bool
	NormalizeMTVName                bool
	TrackTrumpableID                bool
	MatchDVDReleaseGroup            bool
	RequireReleaseGroup             bool
	RejectEpisodeResolutionMismatch bool
	NormalizeDDPlusName             bool
	SDMatchesHD                     bool
	CompareDVDResolution            bool
	AllowSizeVariance1080           bool
}

// DupePolicyProvider declares tracker-owned duplicate comparison policy.
type DupePolicyProvider interface {
	DupePolicy() *DupePolicy
}

// AudioPolicy declares tracker-specific multi-language upload constraints.
type AudioPolicy struct {
	// AllowedLanguages contains normalized languages accepted for foreign audio.
	AllowedLanguages []string
	// BlockEnglishOriginalWithForeign rejects foreign tracks when English is original audio.
	BlockEnglishOriginalWithForeign bool
}

// AudioPolicyProvider declares tracker-owned audio constraints.
type AudioPolicyProvider interface {
	AudioPolicy() *AudioPolicy
}

// ImageHostPolicy declares tracker-owned accepted image hosts and activation gates.
type ImageHostPolicy struct {
	// AllowedHosts lists normalized image hosts accepted in descriptions.
	AllowedHosts []string
	// DisableWithoutRehost disables the policy unless image rehosting is enabled.
	DisableWithoutRehost bool
	// DisableWithoutAPI disables the policy unless tracker image API credentials exist.
	DisableWithoutAPI bool
	// ConditionalHost is enabled only when its associated runtime condition is met.
	ConditionalHost string
	// EnableWithLostimg enables ConditionalHost when LostImg is configured.
	EnableWithLostimg bool
	// EnableWhenConfigured enables ConditionalHost when that uploader is configured.
	EnableWhenConfigured bool
}

// ImageHostPolicyProvider declares tracker-owned image-host restrictions.
type ImageHostPolicyProvider interface {
	ImageHostPolicy() *ImageHostPolicy
}

// ClaimChecker evaluates tracker-owned active-claim rules.
type ClaimChecker interface {
	HasClaim(ctx context.Context, meta api.PreparedMetadata) (bool, error)
	FailureReason(meta api.PreparedMetadata) string
}

// ClaimCheckerFactory constructs a tracker-owned claim checker.
type ClaimCheckerFactory interface {
	NewClaimChecker(cfg config.Config, logger api.Logger) ClaimChecker
}

// ClaimPolicy declares generic claim orchestration required by a tracker.
type ClaimPolicy struct {
	// APIBacked reports that claim evaluation requires a remote tracker lookup.
	APIBacked bool
}

// ClaimPolicyProvider declares tracker-owned claim orchestration policy.
type ClaimPolicyProvider interface {
	ClaimPolicy() *ClaimPolicy
}

// Descriptor binds a tracker definition to its optional capabilities.
type Descriptor struct {
	Name             string
	Kind             Kind
	BaseURL          string
	Definition       Definition
	DupeSearcher     DupeSearcher
	DupeFactory      DupeSearcherFactory
	Rules            *ruletypes.RuleSet
	Artifact         *ArtifactPolicy
	DataFactory      DataLookupFactory
	DataPolicy       *DataLookupPolicy
	BannedGroups     []string
	BannedPolicy     *BannedGroupPolicy
	Metadata         *TrackerMetadataPolicy
	UploadArtifact   *UploadArtifactPolicy
	DupePolicy       *DupePolicy
	AudioPolicy      *AudioPolicy
	ImageHost        *ImageHostPolicy
	ClaimFactory     ClaimCheckerFactory
	ClaimPolicy      *ClaimPolicy
	AuthResolver     AuthSessionResolver
	AuthCapability   *api.TrackerAuthCapability
	MetadataLocale   string
	DescriptionGroup string
}
