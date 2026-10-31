// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package rest

import (
	"testing"

	steamserverlist "github.com/woozymasta/steamserverlist"
)

func TestRestDiscoveryOptionsDeriveBudgets(t *testing.T) {
	options := restDiscoveryOptions(120000, 0, 0, 0, 0, 0)
	if options.MaxPlanQueries != 14 || options.MaxRequests != 27 {
		t.Fatalf("derived REST budgets = (%d leaves, %d requests), want (14, 27)",
			options.MaxPlanQueries, options.MaxRequests)
	}
}

func TestRecordCountAndTruncateNamedServerList(t *testing.T) {
	servers := steamserverlist.Servers{{}, {}, {}}
	if got := recordCount(servers); got != 3 {
		t.Fatalf("recordCount() = %d, want 3", got)
	}
	if got := truncateResult(servers, 2).(steamserverlist.Servers); len(got) != 2 {
		t.Fatalf("truncateResult() returned %d servers, want 2", len(got))
	}
}
