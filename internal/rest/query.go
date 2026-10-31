// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

// Package rest lists and discovers game servers through the Steam Web API.
package rest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/netip"
	"os"
	"time"

	steamserverlist "github.com/woozymasta/steamserverlist"
)

// restPageLimit is the maximum number of rows returned by one list request.
const restPageLimit uint32 = 10000

// Options configures a Web API list or discovery operation.
type Options struct {
	// APIKey authenticates requests to the Steam Web API.
	APIKey string
	// FilterString is the encoded filter used by single-page listing.
	FilterString string
	// Records selects full or address-only server records.
	Records string
	// PlanIn resumes collection from this serialized discovery plan.
	PlanIn string
	// PlanOut receives the generated discovery plan when non-empty.
	PlanOut string
	// Filter is the parsed filter used to create or validate a discovery plan.
	Filter steamserverlist.QueryFilter
	// MaxRequests caps Web API calls; zero derives a budget from Limit.
	MaxRequests int
	// MaxPlanQueries caps discovery leaves; zero derives a budget from Limit.
	MaxPlanQueries int
	// MaxBufferedRecords limits records retained as discovery witnesses.
	MaxBufferedRecords int
	// MaxDepth caps recursive discovery splits.
	MaxDepth int
	// MinValueCount rejects split values with insufficient sample support.
	MinValueCount int
	// Timeout bounds the complete operation and each HTTP request.
	Timeout time.Duration
	// Limit is the maximum number of unique server records to return.
	Limit uint32
}

// Result contains records and discovery metadata returned by a Web API query.
type Result struct {
	// Servers contains the selected record projection.
	Servers any
	// Status contains discovery progress details when discovery was used.
	Status any
	// Complete reports whether discovery exhausted all planned work.
	Complete *bool
	// Filter is the encoded filter used for the operation.
	Filter string
	// PlanPath identifies the input or output discovery plan, when present.
	PlanPath string
	// Count is the number of records retained after applying Limit.
	Count int
	// Requests is the number of Web API requests performed.
	Requests int
	// Elapsed is the operation duration.
	Elapsed time.Duration
}

// Query lists servers directly or discovers them through a reusable plan.
func Query(parent context.Context, options Options) (Result, error) {
	started := time.Now()
	if options.PlanIn != "" && options.FilterString != "" {
		return Result{}, errors.New("a discovery plan cannot be combined with a query filter")
	}

	ctx, cancel := context.WithTimeout(parent, options.Timeout)
	defer cancel()

	client, err := steamserverlist.New(
		options.APIKey,
		steamserverlist.WithHTTPClient(&http.Client{Timeout: options.Timeout}),
	)
	if err != nil {
		return Result{}, fmt.Errorf("create Web API client: %w", err)
	}

	result := Result{Filter: options.FilterString, Requests: 1}
	limit := options.Limit
	// One-page listing is enough up to the API cap
	// unless the caller requests explicit discovery controls or supplies a saved plan.
	forceDiscovery := options.PlanIn != "" ||
		options.MaxRequests != 0 ||
		options.MaxPlanQueries != 0 ||
		options.MaxBufferedRecords != 0 ||
		options.MaxDepth != 0 ||
		options.MinValueCount != 0

	if limit <= restPageLimit && !forceDiscovery {
		result.Servers, err = restList(ctx, client, options.FilterString, limit, options.Records)
	} else {
		planOptions := restDiscoveryOptions(
			limit,
			options.MaxRequests,
			options.MaxPlanQueries,
			options.MaxBufferedRecords,
			options.MaxDepth,
			options.MinValueCount,
		)

		if options.PlanIn != "" {
			// The saved plan owns its filters;
			// the caller only supplies execution controls
			// such as the request budget and record projection.
			var plan steamserverlist.DiscoveryPlan
			plan, err = readPlan(options.PlanIn)
			if err == nil {
				queries := plan.Queries()
				if len(queries) > 0 {
					result.Filter = queries[0].Filter
				}

				result.Servers, result.Status, result.Requests, err = restCollect(
					ctx, client, plan, options.MaxRequests, options.Records,
				)
				if err == nil {
					complete := restStatusComplete(result.Status)
					result.Complete = &complete
				}
			}
		} else {
			var plan steamserverlist.DiscoveryPlan
			result.Servers, plan, result.Status, result.Requests, err = restDiscover(
				ctx, client, options.Filter, planOptions, options.Records,
			)
			if err == nil {
				complete := restStatusComplete(result.Status)
				result.Complete = &complete
				if options.PlanOut != "" {
					err = writePlan(options.PlanOut, plan)
				}
			}
		}
	}
	if err != nil {
		return Result{}, err
	}

	result.Servers = truncateResult(result.Servers, int(limit))
	result.Count = recordCount(result.Servers)
	result.PlanPath = options.PlanOut
	if options.PlanIn != "" {
		result.PlanPath = options.PlanIn
	}
	result.Elapsed = time.Since(started)

	return result, nil
}

// restList requests one Web API page using the requested record projection.
func restList(
	ctx context.Context,
	client *steamserverlist.Client,
	filter string,
	limit uint32,
	mode string,
) (any, error) {
	query := steamserverlist.ListQuery{Filter: filter, Limit: min(limit, restPageLimit)}
	switch mode {
	case "addresses":
		items, err := client.ListAddresses(ctx, query)
		return items, err

	default:
		items, err := client.List(ctx, query)
		return items, err
	}
}

// restDiscover executes adaptive discovery and returns its reusable plan.
func restDiscover(
	ctx context.Context,
	client *steamserverlist.Client,
	filter steamserverlist.QueryFilter,
	options steamserverlist.DiscoveryOptions,
	mode string,
) (any, steamserverlist.DiscoveryPlan, any, int, error) {
	switch mode {
	case "addresses":
		result, err := client.DiscoverAddresses(ctx, filter, options)
		return result.Items, result.Plan, result.Status, result.Status.Requests, err

	default:
		result, err := client.Discover(ctx, filter, options)
		return result.Items, result.Plan, result.Status, result.Status.Requests, err
	}
}

// restCollect executes an existing discovery plan within the request budget.
func restCollect(
	ctx context.Context,
	client *steamserverlist.Client,
	plan steamserverlist.DiscoveryPlan,
	maxRequests int,
	mode string,
) (any, any, int, error) {
	options := steamserverlist.CollectOptions{MaxRequests: maxRequests}
	switch mode {
	case "addresses":
		result, err := client.CollectAddresses(ctx, plan, options)
		return result.Items, result.Status, result.Status.Requests, err

	default:
		result, err := client.Collect(ctx, plan, options)
		return result.Items, result.Status, result.Status.Requests, err
	}
}

// restDiscoveryOptions derives conservative defaults while preserving overrides.
func restDiscoveryOptions(
	limit uint32,
	requests, leaves, buffered, depth, minCount int,
) steamserverlist.DiscoveryOptions {
	// The multiplier leaves headroom because real partitions do not fill evenly.
	calculatedLeaves := (uint64(limit)*100 +
		uint64(restPageLimit)*89 - 1) /
		(uint64(restPageLimit) * 89)

	if calculatedLeaves == 0 {
		calculatedLeaves = 1
	}
	if leaves == 0 {
		if calculatedLeaves > 256 {
			leaves = 256
		} else {
			leaves = int(calculatedLeaves)
		}
	}
	if requests == 0 {
		requests = min(2*leaves-1, 256)
	}

	return steamserverlist.DiscoveryOptions{
		Limit:              restPageLimit,
		MaxRequests:        requests,
		MaxPlanQueries:     leaves,
		MaxBufferedRecords: buffered,
		MaxDepth:           depth,
		MinValueCount:      minCount,
	}
}

// readPlan bounds the input before decoding and validates its query graph.
func readPlan(path string) (steamserverlist.DiscoveryPlan, error) {
	info, err := os.Stat(path)
	if err != nil {
		return steamserverlist.DiscoveryPlan{}, fmt.Errorf("stat discovery plan: %w", err)
	}

	if info.Size() > 1<<20 {
		return steamserverlist.DiscoveryPlan{}, errors.New("discovery plan exceeds 1 MiB")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return steamserverlist.DiscoveryPlan{}, fmt.Errorf("read discovery plan: %w", err)
	}

	var plan steamserverlist.DiscoveryPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return steamserverlist.DiscoveryPlan{}, fmt.Errorf("decode discovery plan: %w", err)
	}

	return plan, plan.Validate()
}

// restStatusComplete converts the API's status value to the CLI result flag.
func restStatusComplete(value any) bool {
	status, ok := value.(steamserverlist.DiscoveryStatus)
	return ok && status.Complete()
}

// writePlan persists a validated discovery plan as readable JSON.
func writePlan(path string, plan steamserverlist.DiscoveryPlan) error {
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("encode discovery plan: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write discovery plan: %w", err)
	}

	return nil
}

// recordCount returns the number of items for each supported projection.
func recordCount(value any) int {
	switch items := value.(type) {
	case steamserverlist.Servers:
		return len(items)

	case []steamserverlist.Server:
		return len(items)

	case []netip.AddrPort:
		return len(items)

	default:
		return 0
	}
}

// truncateResult enforces the caller's total-result limit after discovery.
func truncateResult(value any, limit int) any {
	switch items := value.(type) {
	case steamserverlist.Servers:
		if len(items) > limit {
			return items[:limit]
		}

	case []steamserverlist.Server:
		if len(items) > limit {
			return items[:limit]
		}

	case []netip.AddrPort:
		if len(items) > limit {
			return items[:limit]
		}
	}

	return value
}
