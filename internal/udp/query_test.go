// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package udp

import (
	"context"
	"testing"
	"time"

	"github.com/woozymasta/msqp"
)

func TestQueryRequiresPositiveLimitAndMaster(t *testing.T) {
	if _, _, _, err := Query(
		context.Background(),
		[]string{"master:27011"},
		"",
		msqp.Region(0),
		0,
		0,
		time.Second,
	); err == nil {
		t.Fatal("Query() error = nil for zero limit")
	}

	if _, _, _, err := Query(
		context.Background(),
		nil,
		"",
		msqp.Region(0),
		1,
		0,
		time.Second,
	); err == nil {
		t.Fatal("Query() error = nil without masters")
	}
}
