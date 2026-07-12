// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package additional

import "github.com/autobrr/upbrr/internal/trackers/ruletypes"

type Result = ruletypes.Result
type ExtraCheck = ruletypes.ExtraCheck
type LanguageRule = ruletypes.LanguageRule
type RuleSet = ruletypes.RuleSet

func Pass() Result              { return ruletypes.Pass() }
func Fail(reason string) Result { return ruletypes.Fail(reason) }
