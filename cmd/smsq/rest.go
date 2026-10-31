// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package main

import (
	"context"
	"io"
	"net/netip"
	"strconv"
	"strings"

	"github.com/woozymasta/smsq/internal/rest"
	"github.com/woozymasta/smsq/internal/table"
	slist "github.com/woozymasta/steamserverlist"
)

// runREST adapts CLI options to the Web API package and formats its result.
func runREST(
	parent context.Context,
	filters filterFlags,
	query queryFlags,
	output outputFlags,
	command restCommand,
	stdout io.Writer,
) error {
	if err := planInputConflict(command.Plan.PlanIn, filters); err != nil {
		return err
	}

	if err := rawFilterDiscoveryConflict(
		filters,
		query.Limit,
		slist.DefaultLimit,
		command.Plan,
		command.Discovery.MaxRequests != 0 ||
			command.Discovery.MaxPlanQueries != 0 ||
			command.Discovery.MaxBufferedRecords != 0 ||
			command.Discovery.MaxDepth != 0 ||
			command.Discovery.MinValueCount != 0,
	); err != nil {
		return err
	}

	filter, filterString, _, err := buildFilter(filters)
	if err != nil {
		return err
	}

	result, err := rest.Query(parent, rest.Options{
		APIKey:             command.APIKey,
		Filter:             filter,
		FilterString:       filterString,
		Limit:              query.Limit,
		Records:            command.Records,
		PlanIn:             command.Plan.PlanIn,
		PlanOut:            command.Plan.PlanOut,
		MaxRequests:        command.Discovery.MaxRequests,
		MaxPlanQueries:     command.Discovery.MaxPlanQueries,
		MaxBufferedRecords: command.Discovery.MaxBufferedRecords,
		MaxDepth:           command.Discovery.MaxDepth,
		MinValueCount:      command.Discovery.MinValueCount,
		Timeout:            query.Timeout,
	})
	if err != nil {
		return err
	}

	rows := restRows(result.Servers)
	document := resultDocument{
		Protocol:  "rest",
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

	return outputResult(output, document, rows, restAvailableColumns(result.Servers), stdout)
}

// restAvailableColumns selects the fields present in the concrete Web API result.
func restAvailableColumns(value any) []string {
	switch value.(type) {
	case []netip.AddrPort:
		return []string{table.ColumnAddress}

	case slist.Servers, []slist.Server:
		return table.SharedColumns()

	default:
		return nil
	}
}

// restRows projects Web API records into the CLI's shared table columns.
func restRows(value any) []table.Row {
	switch servers := value.(type) {
	case []netip.AddrPort:
		return addressRows(servers)

	case slist.Servers:
		return restServerRows(servers)

	case []slist.Server:
		return restServerRows(servers)

	default:
		return nil
	}
}

// restServerRows projects full REST records from list and Discovery results.
func restServerRows(servers []slist.Server) []table.Row {
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
			GameType:        strings.Join(server.GameType, ","),
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
}
