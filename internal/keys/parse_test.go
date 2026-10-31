// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package keys

import "testing"

func TestParseNormalizesBooleanOperands(t *testing.T) {
	tests := []struct {
		value string
		want  string
	}{
		{value: "0", want: "0"},
		{value: "1", want: "1"},
		{value: "y", want: "1"},
		{value: "YES", want: "1"},
		{value: "true", want: "1"},
		{value: "n", want: "0"},
		{value: "No", want: "0"},
		{value: "FALSE", want: "0"},
	}

	for _, test := range tests {
		t.Run(test.value, func(t *testing.T) {
			got, err := Parse([]string{"dedicated=" + test.value})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if len(got) != 1 || got[0].Value != test.want {
				t.Fatalf("Parse() = %#v, want dedicated=%s", got, test.want)
			}
		})
	}
}

func TestParseKeepsNonBooleanValueCase(t *testing.T) {
	got, err := Parse([]string{"gametype=BattleEye"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(got) != 1 || got[0].Value != "BattleEye" {
		t.Fatalf("Parse() = %#v, want gametype=BattleEye", got)
	}
}

func TestParseRejectsUnsupportedBooleanSpelling(t *testing.T) {
	if _, err := Parse([]string{"dedicated=on"}); err == nil {
		t.Fatal("Parse() accepted dedicated=on")
	}
}

func TestParseRejectsUnverifiedCollapseAddressFilter(t *testing.T) {
	if _, err := Parse([]string{"collapse_addr_hash=1"}); err == nil {
		t.Fatal("Parse() accepted collapse_addr_hash")
	}
}
