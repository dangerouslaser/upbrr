// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package additional

import "github.com/autobrr/upbrr/pkg/api"

func DiscType(value string) bool                  { return isDiscType(value) }
func Resolution(meta api.PreparedMetadata) string { return resolveResolution(meta) }
func ResolutionBelow(value, minimum string) bool {
	return resolutionOrder[value] < resolutionOrder[minimum]
}
func Type(meta api.PreparedMetadata) string          { return resolveType(meta) }
func Group(meta api.PreparedMetadata) string         { return resolveGroup(meta) }
func DolbyVisionOnly(meta api.PreparedMetadata) bool { return isDolbyVisionOnly(meta) }
func Genres(meta api.PreparedMetadata) []string      { return collectGenres(meta) }
func Keywords(meta api.PreparedMetadata) []string    { return collectKeywords(meta) }
func AdultContent(meta api.PreparedMetadata) bool    { return isAdultContent(meta) }
func Anime(meta api.PreparedMetadata) bool           { return isAnime(meta) }
func Animation(meta api.PreparedMetadata) bool       { return isAnimation(meta) }
func Normalize(values []string) []string             { return normalizeStrings(values) }
func Contains(values, targets []string) bool         { return containsAny(values, targets) }
