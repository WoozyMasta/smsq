// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package grpc

import "testing"

func TestValidateEndpoint(t *testing.T) {
	for _, address := range []string{
		"cmp1-fra2.steamserver.net:27023",
		"192.0.2.1:27023",
		"[2001:db8::1]:27023",
	} {
		if err := validateEndpoint(address); err != nil {
			t.Errorf("validateEndpoint(%q) error = %v", address, err)
		}
	}
	for _, address := range []string{
		"https://cmp1-fra2.steamserver.net:27023",
		"cmp1-fra2.steamserver.net",
		"host:0",
		"host:65536",
		"127.0.0.1:0",
		":27023",
	} {
		if err := validateEndpoint(address); err == nil {
			t.Errorf("validateEndpoint(%q) error = nil", address)
		}
	}
}
