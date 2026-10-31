// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

// Package grpc queries game servers through an authenticated Steam CM session.
package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

	cm "github.com/woozymasta/steamcm"
	"github.com/woozymasta/steamcm/gameservers"
)

// Options configures authentication and a CM server-list operation.
type Options struct {
	// Filter is the encoded master-server filter, including its leading slash.
	Filter string
	// Records selects full or address-only server records.
	Records string
	// AuthMode chooses QR, Steam Guard code, or mobile approval authentication.
	AuthMode string
	// Account selects the Steam account for password authentication.
	Account string
	// DeviceName labels this client during Steam authentication.
	DeviceName string
	// TokenFile optionally stores a reusable refresh token.
	TokenFile string
	// CMEndpoint overrides directory lookup with one host:port endpoint.
	CMEndpoint string
	// PlanIn resumes collection from a serialized discovery plan.
	PlanIn string
	// PlanOut receives the generated discovery plan when non-empty.
	PlanOut string
	// Timeout bounds authentication, directory lookup, and server queries.
	Timeout time.Duration
	// MaxRequests caps discovery requests; zero derives a budget from Limit.
	MaxRequests int
	// MaxDepth caps recursive discovery splits.
	MaxDepth int
	// RequestInterval spaces later request starts; zero disables pacing.
	RequestInterval time.Duration
	// InitialBurst allows this many requests before RequestInterval pacing.
	InitialBurst int
	// AppID identifies the game whose servers are requested.
	AppID uint32
	// Limit caps unique server records returned to the caller.
	Limit uint32
}

// Result contains server records and progress metadata from a CM operation.
type Result struct {
	// Servers contains full discovery records or address-only endpoints.
	Servers any
	// Status contains discovery progress details when discovery was used.
	Status any
	// Complete reports whether all planned work was exhausted.
	Complete *bool
	// Filter is the effective encoded filter.
	Filter string
	// PlanPath identifies the input or output discovery plan, when present.
	PlanPath string
	// Count is the number of retained server records.
	Count int
	// Requests is the number of CM query calls performed.
	Requests int
	// Elapsed is the operation duration.
	Elapsed time.Duration
}

// Run authenticates to Steam CM and queries or discovers game servers.
func Run(
	parent context.Context,
	options Options,
	stdin *os.File,
	stderr io.Writer,
) (Result, error) {
	started := time.Now()
	ctx, cancel := context.WithTimeout(parent, options.Timeout)
	defer cancel()

	endpoints, err := resolveEndpoints(ctx, options.CMEndpoint, options.Timeout)
	if err != nil {
		return Result{}, err
	}

	storage := tokenStorage{path: options.TokenFile}
	login, err := authenticateCM(
		ctx,
		endpoints,
		options.AuthMode,
		options.Account,
		options.DeviceName,
		storage,
		stdin,
		stderr,
		options.Timeout,
	)
	if err != nil {
		return Result{}, err
	}

	defer func() { _ = login.client.Close() }()

	client, err := gameservers.New(login.session)
	if err != nil {
		return Result{}, err
	}

	result := Result{Filter: options.Filter, Requests: 1}

	// Keep ordinary requests to one CM page.
	// The richer MasterServerQuery response supplies metadata;
	// the historical GetServerList method may fail.
	forceDiscovery := options.PlanIn != "" ||
		options.MaxRequests != 0 ||
		options.MaxDepth != 0 ||
		options.RequestInterval != 0 ||
		options.InitialBurst != 0

	if options.Limit <= gameservers.DefaultDiscoveryLimit && !forceDiscovery {
		result.Servers, err = cmQuery(
			ctx,
			client,
			options.Filter,
			options.AppID,
			options.Limit,
			options.Records,
		)
	} else {
		budget := options.MaxRequests
		if budget == 0 && options.PlanIn == "" {
			leaves := derivedLeafBudget(options.Limit, gameservers.DefaultDiscoveryLimit)
			budget = min(2*leaves-1, 256)
		}

		discoveryOptions := gameservers.DiscoveryOptions{
			MaxRequests:        budget,
			MaxDepth:           options.MaxDepth,
			MinRequestInterval: options.RequestInterval,
			InitialBurst:       options.InitialBurst,
		}

		if options.PlanIn != "" {
			// A plan stores its own root filter and AppID,
			// so do not silently combine it with a newly constructed query.
			var plan gameservers.DiscoveryPlan
			plan, err = readPlan(options.PlanIn)
			if err == nil && plan.Root.AppID != options.AppID {
				err = errors.New("AppID does not match the saved CM plan")
			}

			if err == nil {
				result.Filter = plan.Root.Filter
				result.Servers, result.Status, result.Requests, err = cmCollect(
					ctx, client, plan, options.Records,
					gameservers.CollectionOptions{
						MaxRequests:        budget,
						MinRequestInterval: options.RequestInterval,
						InitialBurst:       options.InitialBurst,
					},
				)
			}
		} else {
			root := gameservers.QueryRequest{
				AppID:  options.AppID,
				Filter: options.Filter,
				Limit:  min(options.Limit, gameservers.DefaultDiscoveryLimit),
			}

			var plan gameservers.DiscoveryPlan
			result.Servers, plan, result.Status, result.Requests, err = cmDiscover(
				ctx, client, root, discoveryOptions, options.Records,
			)
			if err == nil && options.PlanOut != "" {
				err = writePlan(options.PlanOut, plan)
			}
		}

		if err == nil {
			result.Complete = cmStatusComplete(result.Status)
		}
	}

	if err != nil {
		return Result{}, err
	}

	result.Servers = truncateResult(result.Servers, int(options.Limit))
	result.Count = recordCount(result.Servers)
	result.PlanPath = options.PlanOut
	if options.PlanIn != "" {
		result.PlanPath = options.PlanIn
	}
	result.Elapsed = time.Since(started)

	return result, nil
}

// resolveEndpoints validates an explicit endpoint or reads Steam's CM directory.
func resolveEndpoints(ctx context.Context, address string, timeout time.Duration) ([]cm.Endpoint, error) {
	if address != "" {
		if err := validateEndpoint(address); err != nil {
			return nil, err
		}

		return []cm.Endpoint{{Address: address}}, nil
	}

	directory := cm.Directory{HTTPClient: &http.Client{Timeout: timeout}}
	endpoints, err := directory.Get(ctx, cm.DirectoryQuery{
		CMType:   "websockets",
		MaxCount: 10,
		QoSLevel: 2,
	})
	if err != nil {
		return nil, fmt.Errorf("query Steam CM directory: %w", err)
	}

	return endpoints, nil
}

// cmQuery runs one CM query and projects either full records or addresses.
func cmQuery(
	ctx context.Context,
	client *gameservers.Client,
	filter string,
	appID, limit uint32,
	mode string,
) (any, error) {
	if mode == "addresses" {
		return client.Query(ctx, gameservers.QueryRequest{AppID: appID, Filter: filter, Limit: limit})
	}

	result, err := client.DiscoverServers(
		ctx,
		gameservers.QueryRequest{AppID: appID, Filter: filter, Limit: limit},
		gameservers.DiscoveryOptions{MaxRequests: 1},
	)
	if err != nil {
		return nil, err
	}

	return result.Servers, nil
}

// cmDiscover captures detailed response records only when full output is needed.
func cmDiscover(
	ctx context.Context,
	client *gameservers.Client,
	root gameservers.QueryRequest,
	options gameservers.DiscoveryOptions,
	mode string,
) (any, gameservers.DiscoveryPlan, any, int, error) {
	if mode == "full" {
		result, err := client.DiscoverServers(ctx, root, options)
		return result.Servers, result.Plan, result.Status, result.Status.Requests, err
	}

	result, err := client.Discover(ctx, root, options)

	return result.Endpoints, result.Plan, result.Status, result.Status.Requests, err
}

// cmCollect resumes an existing plan and optionally retains full response records.
func cmCollect(
	ctx context.Context,
	client *gameservers.Client,
	plan gameservers.DiscoveryPlan,
	mode string,
	options gameservers.CollectionOptions,
) (any, any, int, error) {
	if mode == "full" {
		result, err := client.CollectServers(ctx, plan, options)
		return result.Servers, result.Status, result.Status.Requests, err
	}

	result, err := client.Collect(ctx, plan, options)
	return result.Endpoints, result.Status, result.Status.Requests, err
}

// readPlan bounds the input before decoding and validates its request graph.
func readPlan(path string) (gameservers.DiscoveryPlan, error) {
	info, err := os.Stat(path)
	if err != nil {
		return gameservers.DiscoveryPlan{}, fmt.Errorf("stat CM plan: %w", err)
	}

	if info.Size() > 1<<20 {
		return gameservers.DiscoveryPlan{}, errors.New("CM plan exceeds 1 MiB")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return gameservers.DiscoveryPlan{}, fmt.Errorf("read CM plan: %w", err)
	}

	var plan gameservers.DiscoveryPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return gameservers.DiscoveryPlan{}, fmt.Errorf("decode CM plan: %w", err)
	}

	if err := plan.Validate(); err != nil {
		return gameservers.DiscoveryPlan{}, err
	}

	return plan, nil
}

// writePlan persists the generated discovery plan as readable JSON.
func writePlan(path string, plan gameservers.DiscoveryPlan) error {
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("encode CM plan: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write CM plan: %w", err)
	}

	return nil
}

// cmStatusComplete reports whether discovery exhausted every pending request.
func cmStatusComplete(value any) *bool {
	status, ok := value.(gameservers.DiscoveryStatus)
	if !ok {
		return nil
	}

	return new(status.Complete)
}

// recordCount returns the number of records in either supported projection.
func recordCount(value any) int {
	switch records := value.(type) {
	case []gameservers.Server:
		return len(records)

	case []gameservers.DiscoveryServer:
		return len(records)

	case []gameservers.Endpoint:
		return len(records)

	default:
		return 0
	}
}

// truncateResult enforces the requested total limit after list or discovery.
func truncateResult(value any, limit int) any {
	switch records := value.(type) {
	case []gameservers.Server:
		if len(records) > limit {
			return records[:limit]
		}

	case []gameservers.DiscoveryServer:
		if len(records) > limit {
			return records[:limit]
		}

	case []gameservers.Endpoint:
		if len(records) > limit {
			return records[:limit]
		}
	}

	return value
}

// derivedLeafBudget estimates useful discovery partitions for the requested cap.
func derivedLeafBudget(limit, pageLimit uint32) int {
	const usefulPagePercent = 89
	// Allow for uneven partitions so discovery does not assume every query fills.
	leaves := (uint64(limit)*100 +
		uint64(pageLimit)*usefulPagePercent - 1) /
		(uint64(pageLimit) * usefulPagePercent)

	if leaves == 0 {
		return 1
	}
	if leaves > 256 {
		return 256
	}

	return int(leaves)
}

// validateEndpoint accepts a CM host:port address but rejects URL-shaped input.
func validateEndpoint(address string) error {
	if strings.Contains(address, "://") {
		return errors.New("-cm-endpoint expects host:port, not a URL")
	}

	if endpoint, err := netip.ParseAddrPort(address); err == nil {
		if endpoint.Port() == 0 {
			return errors.New("-cm-endpoint port must be in 1..65535")
		}

		return nil
	}

	host, portText, err := net.SplitHostPort(address)
	if err != nil || host == "" {
		return errors.New("-cm-endpoint must be host:port")
	}

	port, err := strconv.ParseUint(portText, 10, 16)
	if err != nil || port == 0 {
		return errors.New("-cm-endpoint port must be in 1..65535")
	}

	return nil
}
