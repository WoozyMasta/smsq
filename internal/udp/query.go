// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

// Package udp queries legacy Steam master servers.
package udp

import (
	"context"
	"errors"
	"fmt"
	"net/netip"
	"time"

	"github.com/woozymasta/msqp"
)

const maxInitialAddressCapacity = 10000

// Query reads unique server addresses from the configured legacy masters.
// complete is false when the limit or page budget stops collection early.
func Query(
	ctx context.Context,
	masters []string,
	filter string,
	region msqp.Region,
	limit, maxPages uint,
	ioTimeout time.Duration,
) (addresses []netip.AddrPort, requests int, complete bool, err error) {
	if limit == 0 {
		return nil, 0, false, errors.New("UDP result limit must be positive")
	}
	if len(masters) == 0 {
		return nil, 0, false, errors.New("at least one UDP master is required")
	}

	capacity := int(min(uint64(limit), maxInitialAddressCapacity))
	addresses = make([]netip.AddrPort, 0, capacity)
	seen := make(map[netip.AddrPort]struct{}, capacity)
	complete = true

	for index, master := range masters {
		client, clientErr := msqp.NewClient(master, msqp.WithIOTimeout(ioTimeout))
		if clientErr != nil {
			return nil, requests, false, fmt.Errorf("connect UDP master %q: %w", master, clientErr)
		}

		masterRequests, exhausted, queryErr := collect(
			ctx, client, filter, region, limit-uint(len(addresses)), maxPages, seen, &addresses,
		)
		requests += masterRequests
		// A result is complete only when every queried master was exhausted.
		complete = complete && exhausted
		closeErr := client.Close()
		if queryErr != nil {
			return nil, requests, false, fmt.Errorf("query UDP master %q: %w", master, queryErr)
		}
		if closeErr != nil {
			return nil, requests, false, fmt.Errorf("close UDP master %q: %w", master, closeErr)
		}

		if uint(len(addresses)) >= limit {
			if index < len(masters)-1 {
				complete = false
			}

			break
		}
	}

	return addresses, requests, complete, nil
}

// collect follows one master's cursor, deduplicating endpoints across pages.
func collect(
	ctx context.Context,
	client *msqp.Client,
	filter string,
	region msqp.Region,
	remaining, maxPages uint,
	seen map[netip.AddrPort]struct{},
	addresses *[]netip.AddrPort,
) (requests int, exhausted bool, err error) {
	if remaining == 0 {
		return 0, false, nil
	}

	startCount := len(*addresses)
	request := msqp.Request{Cursor: msqp.InitialCursor(), Filter: filter, Region: region}
	for pageNumber := uint(0); uint(len(*addresses)-startCount) < remaining; pageNumber++ {
		if maxPages > 0 && pageNumber >= maxPages {
			return requests, false, nil
		}

		page, queryErr := client.Query(ctx, request)
		requests++
		if queryErr != nil {
			return requests, false, queryErr
		}

		for _, endpoint := range page.Servers {
			if _, exists := seen[endpoint]; exists {
				continue
			}

			seen[endpoint] = struct{}{}
			*addresses = append(*addresses, endpoint)
			if uint(len(*addresses)-startCount) >= remaining {
				break
			}
		}

		if page.Done {
			return requests, true, nil
		}

		// The next request must advance the cursor
		// or collection would loop over the same page indefinitely.
		if !page.NextCursor.Addr.IsValid() || page.NextCursor == request.Cursor {
			return requests, false, msqp.ErrCursorDidNotAdvance
		}

		request.Cursor = page.NextCursor
	}

	return requests, false, nil
}
