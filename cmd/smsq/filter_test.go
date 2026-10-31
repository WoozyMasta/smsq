// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package main

import "testing"

func TestBuildFilterAddsExplicitAppID(t *testing.T) {
	_, encoded, appID, err := buildFilter(filterFlags{
		AppID:  730,
		Filter: []string{"map=de_dust2", "gametype=battleye"},
		NOR:    []string{"password=1"},
	})
	if err != nil {
		t.Fatalf("buildFilter() error = %v", err)
	}

	const want = "appid\\730\\map\\de_dust2\\gametype\\battleye\\nor\\1\\password\\1"
	if encoded != want {
		t.Fatalf("filter = %q, want %q", encoded, want)
	}
	if appID != 730 {
		t.Fatalf("AppID = %d, want 730", appID)
	}
}

func TestBuildFilterRejectsInvalidConditions(t *testing.T) {
	for name, flags := range map[string]filterFlags{
		"missing app ID":           {Filter: []string{"map=de_dust2"}},
		"missing raw AppID source": {RawFilter: `\map\de_dust2`},
		"unknown appid filter":     {AppID: 730, Filter: []string{"appid=730"}},
		"conflicting exact maps":   {AppID: 730, Filter: []string{"map=de_dust2", "map=de_inferno"}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, _, err := buildFilter(flags); err == nil {
				t.Fatal("buildFilter() succeeded, want an error")
			}
		})
	}
}

func TestBuildFilterUsesFlagAppIDForRawPredicates(t *testing.T) {
	const raw = `\gamedataor\tag1,tag2\custom\value`
	_, encoded, appID, err := buildFilter(filterFlags{AppID: 730, RawFilter: raw})
	if err != nil {
		t.Fatalf("buildFilter() error = %v", err)
	}

	const want = `\appid\730\gamedataor\tag1,tag2\custom\value`
	if encoded != want {
		t.Fatalf("raw filter = %q, want %q", encoded, want)
	}
	if appID != 730 {
		t.Fatalf("AppID = %d, want 730", appID)
	}
}

func TestBuildFilterUsesEmbeddedRawAppID(t *testing.T) {
	const raw = `\appid\730\map\de_dust2`
	_, encoded, appID, err := buildFilter(filterFlags{RawFilter: raw})
	if err != nil {
		t.Fatalf("buildFilter() error = %v", err)
	}
	if encoded != raw {
		t.Fatalf("raw filter = %q, want %q", encoded, raw)
	}
	if appID != 730 {
		t.Fatalf("AppID = %d, want 730", appID)
	}
}

func TestBuildFilterPrefersEmbeddedRawAppID(t *testing.T) {
	const raw = `\appid\10\map\de_dust2`
	_, encoded, appID, err := buildFilter(filterFlags{AppID: 730, RawFilter: raw})
	if err != nil {
		t.Fatalf("buildFilter() error = %v", err)
	}
	if encoded != raw {
		t.Fatalf("raw filter = %q, want %q", encoded, raw)
	}
	if appID != 10 {
		t.Fatalf("AppID = %d, want embedded AppID 10", appID)
	}
}

func TestBuildFilterRejectsInvalidEmbeddedRawAppID(t *testing.T) {
	for _, raw := range []string{`\appid\0`, `\appid\not-a-number`} {
		t.Run(raw, func(t *testing.T) {
			if _, _, _, err := buildFilter(filterFlags{AppID: 730, RawFilter: raw}); err == nil {
				t.Fatal("buildFilter() accepted an invalid embedded AppID")
			}
		})
	}
}
