// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package grpc

import (
	"path/filepath"
	"testing"

	"github.com/woozymasta/steamcm/auth"
)

func TestTokenStorageRoundTrip(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "token.json")
	token := &auth.Token{
		SteamID:      76561198000000000,
		AccountName:  "test-account",
		RefreshToken: "header.payload.signature",
		GuardData:    "guard-context",
	}

	plain := tokenStorage{path: path}
	if err := plain.write(token); err != nil {
		t.Fatalf("write plaintext token: %v", err)
	}
	got, err := plain.read()
	if err != nil {
		t.Fatalf("read plaintext token: %v", err)
	}
	assertStoredToken(t, got, token)
}

// assertStoredToken compares all persisted token fields without logging secrets.
func assertStoredToken(t *testing.T, got, want *auth.Token) {
	t.Helper()
	if got == nil ||
		got.SteamID != want.SteamID ||
		got.AccountName != want.AccountName ||
		got.RefreshToken != want.RefreshToken ||
		got.GuardData != want.GuardData {
		t.Fatalf("stored token = %#v, want %#v", got, want)
	}
}
