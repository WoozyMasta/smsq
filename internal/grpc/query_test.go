// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package grpc

import "testing"

func TestDerivedLeafBudget(t *testing.T) {
	if got := derivedLeafBudget(120000, 10000); got != 14 {
		t.Fatalf("derivedLeafBudget() = %d, want 14", got)
	}
}
