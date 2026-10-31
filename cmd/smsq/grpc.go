// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package main

import (
	"context"
	"io"
	"os"
	"strconv"

	cmclient "github.com/woozymasta/smsq/internal/grpc"
	"github.com/woozymasta/smsq/internal/table"
	"github.com/woozymasta/steamcm/gameservers"
)

// runGRPC adapts CLI options to Steam CM and formats the returned records.
func runGRPC(
	parent context.Context,
	filters filterFlags,
	query queryFlags,
	output outputFlags,
	command grpcCommand,
	stdin *os.File,
	stdout, stderr io.Writer,
) error {
	if err := planInputConflict(command.Plan.PlanIn, filters); err != nil {
		return err
	}

	if err := rawFilterDiscoveryConflict(
		filters,
		query.Limit,
		gameservers.DefaultDiscoveryLimit,
		command.Plan,
		command.Discovery.MaxRequests != 0 ||
			command.Discovery.MaxDepth != 0 ||
			command.Connection.RequestInterval != 0 ||
			command.Connection.InitialBurst != 0,
	); err != nil {
		return err
	}

	filter, filterString, appID, err := buildFilter(filters)
	if err != nil {
		return err
	}

	if filters.RawFilter == "" {
		filterString = filter.String()
		filterString = `\` + filterString
	}

	result, err := cmclient.Run(
		parent,
		cmclient.Options{
			AppID:           appID,
			Filter:          filterString,
			Limit:           query.Limit,
			Timeout:         query.Timeout,
			Records:         command.Discovery.Records,
			AuthMode:        command.Auth.Mode,
			Account:         command.Auth.Account,
			DeviceName:      command.Auth.DeviceName,
			TokenFile:       command.Token.File,
			CMEndpoint:      command.Connection.Endpoint,
			MaxRequests:     command.Discovery.MaxRequests,
			MaxDepth:        command.Discovery.MaxDepth,
			RequestInterval: command.Connection.RequestInterval,
			InitialBurst:    command.Connection.InitialBurst,
			PlanIn:          command.Plan.PlanIn,
			PlanOut:         command.Plan.PlanOut,
		},
		stdin,
		stderr,
	)
	if err != nil {
		return err
	}

	rows := grpcRows(result.Servers)
	document := resultDocument{
		Protocol:  "grpc",
		Filter:    result.Filter,
		Servers:   result.Servers,
		Count:     result.Count,
		Requests:  result.Requests,
		Complete:  result.Complete,
		Status:    result.Status,
		PlanFile:  result.PlanPath,
		ElapsedMS: result.Elapsed.Milliseconds(),
	}
	if output.Format == "table" {
		document.Servers = nil
	}

	return outputResult(output, document, rows, grpcAvailableColumns(result.Servers), stdout)
}

// grpcAvailableColumns selects fields represented by the concrete CM result type.
func grpcAvailableColumns(value any) []string {
	switch value.(type) {
	case []gameservers.Endpoint:
		return []string{table.ColumnAddress}

	case []gameservers.DiscoveryServer:
		return table.DiscoveryColumns()

	case []gameservers.Server:
		return table.SharedColumns()

	default:
		return nil
	}
}

// grpcRows converts CM record projections to the CLI's shared table columns.
func grpcRows(value any) []table.Row {
	switch servers := value.(type) {
	case []gameservers.Endpoint:
		rows := make([]table.Row, len(servers))
		for index, endpoint := range servers {
			rows[index].Address = endpoint.Addr.String()
		}
		return rows

	case []gameservers.Server:
		rows := make([]table.Row, len(servers))
		for index, server := range servers {
			rows[index] = table.Row{
				Address:         server.Addr.String(),
				Name:            server.Name,
				Map:             server.Map,
				Game:            server.Product,
				Version:         server.Version,
				GameDir:         server.GameDir,
				OS:              server.OS,
				GameType:        server.GameType,
				SteamID:         strconv.FormatUint(server.SteamID, 10),
				Region:          server.Region,
				Players:         server.Players,
				MaxPlayers:      server.MaxPlayers,
				Bots:            server.Bots,
				GamePort:        uint32(server.GamePort),
				SpectatorPort:   uint32(server.SpectatorPort),
				Dedicated:       server.Dedicated,
				Secure:          server.Secure,
				HasDetails:      true,
				HasPlatformInfo: true,
			}
		}
		return rows

	case []gameservers.DiscoveryServer:
		rows := make([]table.Row, len(servers))
		for index, server := range servers {
			rows[index] = table.Row{
				Address:               server.Addr,
				Name:                  server.Name,
				Map:                   server.Map,
				Game:                  server.GameDescription,
				Version:               server.Version,
				GameDir:               server.GameDir,
				GameType:              server.GameType,
				SteamID:               strconv.FormatUint(server.SteamID, 10),
				Players:               grpcTableCount(server.Players),
				MaxPlayers:            grpcTableCount(server.MaxPlayers),
				Bots:                  grpcTableCount(server.Bots),
				GamePort:              server.GamePort,
				SpectatorPort:         server.SpectatorPort,
				AuthPlayers:           server.AuthPlayers,
				SpectatorName:         server.SpectatorName,
				SDRPingLocation:       server.SDRPingLocation,
				Revision:              server.Revision,
				QueryPort:             server.QueryPort,
				SDRPopID:              server.SDRPopID,
				FriendsCount:          server.FriendsCount,
				FriendsCountConnected: server.FriendsCountConnected,
				PingEstimate:          server.PingEstimate,
				Flags:                 server.Flags,
				MapKnown:              server.MapKnown,
				VersionKnown:          server.VersionKnown,
				GameTypeKnown:         server.GameTypeKnown,
				HasDetails:            true,
				HasCMMetadata:         true,
			}
		}
		return rows

	default:
		return nil
	}
}

// grpcTableCount clamps protocol counts to the signed table representation.
func grpcTableCount(value uint32) int32 {
	const maxInt32 = int32(^uint32(0) >> 1)
	if value > uint32(maxInt32) {
		return maxInt32
	}

	return int32(value)
}
