// Copyright (c) 2025-2026, Audionut and the autobrr contributors.
// SPDX-License-Identifier: GPL-2.0-or-later

package impl

import (
	"fmt"
	"strings"

	"github.com/autobrr/upbrr/internal/config"
	"github.com/autobrr/upbrr/internal/trackers"
	"github.com/autobrr/upbrr/internal/trackers/impl/ant"
	"github.com/autobrr/upbrr/internal/trackers/impl/ar"
	"github.com/autobrr/upbrr/internal/trackers/impl/asc"
	"github.com/autobrr/upbrr/internal/trackers/impl/azfamily"
	"github.com/autobrr/upbrr/internal/trackers/impl/bhd"
	"github.com/autobrr/upbrr/internal/trackers/impl/bhdtv"
	"github.com/autobrr/upbrr/internal/trackers/impl/bjs"
	"github.com/autobrr/upbrr/internal/trackers/impl/bt"
	"github.com/autobrr/upbrr/internal/trackers/impl/btn"
	"github.com/autobrr/upbrr/internal/trackers/impl/czt"
	"github.com/autobrr/upbrr/internal/trackers/impl/dc"
	"github.com/autobrr/upbrr/internal/trackers/impl/ff"
	"github.com/autobrr/upbrr/internal/trackers/impl/fl"
	"github.com/autobrr/upbrr/internal/trackers/impl/gpw"
	"github.com/autobrr/upbrr/internal/trackers/impl/hdb"
	"github.com/autobrr/upbrr/internal/trackers/impl/hds"
	"github.com/autobrr/upbrr/internal/trackers/impl/hdt"
	"github.com/autobrr/upbrr/internal/trackers/impl/is"
	"github.com/autobrr/upbrr/internal/trackers/impl/mtv"
	"github.com/autobrr/upbrr/internal/trackers/impl/nbl"
	"github.com/autobrr/upbrr/internal/trackers/impl/ptp"
	"github.com/autobrr/upbrr/internal/trackers/impl/pts"
	"github.com/autobrr/upbrr/internal/trackers/impl/rtf"
	"github.com/autobrr/upbrr/internal/trackers/impl/spd"
	"github.com/autobrr/upbrr/internal/trackers/impl/thr"
	"github.com/autobrr/upbrr/internal/trackers/impl/tl"
	"github.com/autobrr/upbrr/internal/trackers/impl/tvc"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/a4k"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/acm"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/aither"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/blu"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/cbr"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/dp"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/emuw"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/friki"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/hhd"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/ihd"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/itt"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/lcd"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/ldu"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/lst"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/lt"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/lume"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/mns"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/oe"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/otw"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/pt"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/ptt"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/r4e"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/ras"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/rf"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/rhd"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/sam"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/shri"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/sp"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/stc"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/tik"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/tlz"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/tos"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/ttr"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/ulcx"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/utp"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/yus"
	"github.com/autobrr/upbrr/internal/trackers/impl/unit3d/sites/znth"
	"github.com/autobrr/upbrr/internal/trackers/ruletypes"
)

func withRules(profile unit3d.Profile, rules *ruletypes.RuleSet) unit3d.Profile {
	profile.Rules = rules
	return profile
}

func withBannedGroups(profile unit3d.Profile, groups []string) unit3d.Profile {
	profile.BannedGroups = append([]string(nil), groups...)
	return profile
}

func NewRegistry() (*trackers.Registry, error) {
	registry := trackers.NewRegistry()
	profiles := []unit3d.Profile{
		withRules(withBannedGroups(a4k.Profile(), a4k.BannedGroups()), a4k.Rules()), acm.Profile(), withRules(aither.Profile(), aither.Rules()), withRules(withBannedGroups(blu.Profile(), blu.BannedGroups()), blu.Rules()), withBannedGroups(cbr.Profile(), cbr.BannedGroups()), withRules(withBannedGroups(dp.Profile(), dp.BannedGroups()), dp.Rules()), emuw.Profile(), friki.Profile(), withRules(withBannedGroups(hhd.Profile(), hhd.BannedGroups()), hhd.Rules()), ihd.Profile(), itt.Profile(), lcd.Profile(), ldu.Profile(), withBannedGroups(lt.Profile(), lt.BannedGroups()),
		withRules(lume.Profile(), lume.Rules()), withRules(lst.Profile(), lst.Rules()), withRules(mns.Profile(), mns.Rules()), pt.Profile(), withBannedGroups(ptt.Profile(), ptt.BannedGroups()), r4e.Profile(), withRules(withBannedGroups(ras.Profile(), ras.BannedGroups()), ras.Rules()), withRules(rf.Profile(), rf.Rules()), withRules(withBannedGroups(rhd.Profile(), rhd.BannedGroups()), rhd.Rules()), sam.Profile(),
		withRules(withBannedGroups(oe.Profile(), oe.BannedGroups()), oe.Rules()), withRules(withBannedGroups(otw.Profile(), otw.BannedGroups()), otw.Rules()), withRules(shri.Profile(), shri.Rules()), withRules(sp.Profile(), sp.Rules()), withRules(stc.Profile(), stc.Rules()), withRules(tik.Profile(), tik.Rules()), tlz.Profile(), withRules(withBannedGroups(tos.Profile(), tos.BannedGroups()), tos.Rules()), withRules(ttr.Profile(), ttr.Rules()), withRules(withBannedGroups(ulcx.Profile(), ulcx.BannedGroups()), ulcx.Rules()), withRules(znth.Profile(), znth.Rules()),
		utp.Profile(), withBannedGroups(yus.Profile(), yus.BannedGroups()),
	}
	if err := unit3d.RegisterProfiles(registry, profiles); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(hdb.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(mtv.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(ant.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(ar.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(asc.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(bhd.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(bhdtv.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(bjs.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(btn.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(bt.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(czt.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(dc.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(ff.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(fl.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(gpw.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(hds.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(hdt.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(is.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(nbl.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(ptp.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(pts.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(rtf.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(spd.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(thr.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(tl.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	if err := registry.Register(tvc.New()); err != nil {
		return nil, fmt.Errorf("trackers: %w", err)
	}
	for _, name := range []string{"AZ", "CZ", "PHD"} {
		if err := registry.Register(azfamily.New(name)); err != nil {
			return nil, fmt.Errorf("trackers: %w", err)
		}
	}
	registry.SetPriorityOrder([]string{"aither", "ulcx", "lst", "blu", "oe", "btn", "bhd", "hdb", "ant", "rf", "otw", "yus", "dp", "sp", "ptp"})
	return registry, nil
}

// NewRegistryWithConfig composes built-in definitions and configured custom
// Unit3D trackers. Runtime config URLs remain authoritative in the Unit3D client.
func NewRegistryWithConfig(cfg config.Config) (*trackers.Registry, error) {
	registry, err := NewRegistry()
	if err != nil {
		return nil, err
	}
	for name := range cfg.Trackers.Trackers {
		normalized := strings.ToUpper(strings.TrimSpace(name))
		if normalized == "" {
			continue
		}
		if _, exists := registry.LookupDescriptor(normalized); exists {
			continue
		}
		if !unit3d.IsConfiguredTrackerWithRegistry(cfg, normalized, registry) {
			continue
		}
		if err := registry.Register(unit3d.New(normalized)); err != nil {
			return nil, fmt.Errorf("trackers: register custom unit3d %s: %w", normalized, err)
		}
	}
	return registry, nil
}
