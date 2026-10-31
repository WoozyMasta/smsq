// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

// Package table projects server results into selectable output columns.
package table

import (
	"fmt"
	"slices"
	"strings"
)

const (
	// ColumnAddress selects the server address and query port.
	ColumnAddress = "address"
	// ColumnPlayers selects the current player count and slot capacity.
	ColumnPlayers = "players"
	// ColumnMap selects the server's current map.
	ColumnMap = "map"
	// ColumnName selects the advertised server name.
	ColumnName = "name"
	// ColumnGame selects the server's game description.
	ColumnGame = "game"
	// ColumnVersion selects the advertised game version.
	ColumnVersion = "version"
	// ColumnGameDir selects the game's directory identifier.
	ColumnGameDir = "gamedir"
	// ColumnOS selects the reported server operating system.
	ColumnOS = "os"
	// ColumnGameType selects the server's advertised tags.
	ColumnGameType = "gametype"
	// ColumnSteamID selects the server's Steam identifier.
	ColumnSteamID = "steamid"
	// ColumnRegion retains the unsupported backend region field for future use.
	ColumnRegion = "region"
	// ColumnBots selects the reported bot count.
	ColumnBots = "bots"
	// ColumnGamePort selects the gameplay port.
	ColumnGamePort = "gameport"
	// ColumnSpectatorPort selects the spectator port.
	ColumnSpectatorPort = "specport"
	// ColumnDedicated selects whether the server is dedicated.
	ColumnDedicated = "dedicated"
	// ColumnSecure selects whether the server advertises secure authentication.
	ColumnSecure = "secure"
	// ColumnAuthPlayers selects the authenticated CM player count.
	ColumnAuthPlayers = "auth_players"
	// ColumnSpectatorName selects the advertised spectator server name.
	ColumnSpectatorName = "spectator_name"
	// ColumnSDRPingLocation selects the Steam Datagram Relay location.
	ColumnSDRPingLocation = "sdr_ping_location"
	// ColumnRevision retains the CM record revision for future use.
	ColumnRevision = "revision"
	// ColumnQueryPort retains the CM query port, already present in Address.
	ColumnQueryPort = "queryport"
	// ColumnSDRPopID retains the technical Steam Datagram Relay point ID.
	ColumnSDRPopID = "sdr_popid"
	// ColumnFriendsCount selects the CM-reported friend count.
	ColumnFriendsCount = "friends_count"
	// ColumnFriendsCountConnected selects the connected CM friend count.
	ColumnFriendsCountConnected = "friends_count_connected"
	// ColumnPingEstimate selects the CM-reported ping estimate.
	ColumnPingEstimate = "ping_estimate"
	// ColumnFlags retains opaque CM server flags for possible diagnostics.
	ColumnFlags = "flags"
	// ColumnMapKnown retains CM map-resolution diagnostics for future use.
	ColumnMapKnown = "map_known"
	// ColumnVersionKnown retains CM version-resolution diagnostics for future use.
	ColumnVersionKnown = "version_known"
	// ColumnGameTypeKnown retains CM tag-resolution diagnostics for future use.
	ColumnGameTypeKnown = "gametype_known"
	// ColumnAll expands to every field available for the selected result.
	ColumnAll = "all"
)

const (
	// FormatHTML selects go-pretty's HTML table renderer.
	FormatHTML = "html"
	// FormatMarkdown selects go-pretty's Markdown table renderer.
	FormatMarkdown = "md"
	// FormatCSV selects go-pretty's CSV renderer.
	FormatCSV = "csv"
	// FormatTable selects go-pretty's aligned terminal renderer.
	FormatTable = "table"
	// FormatYAML selects YAML serialization for the result document.
	FormatYAML = "yaml"
	// FormatJSON selects JSON serialization for the result document.
	FormatJSON = "json"
)

var (
	// defaultColumns are the useful fields shown when no selection is supplied.
	defaultColumns = []string{ColumnAddress, ColumnPlayers, ColumnMap, ColumnName}

	// selectableColumns are shared server fields available through structured selection.
	selectableColumns = []string{
		ColumnAddress,
		ColumnPlayers,
		ColumnMap,
		ColumnName,
		ColumnVersion,
		ColumnGameDir,
		ColumnOS,
		ColumnGameType,
		ColumnBots,
		ColumnGamePort,
		ColumnSecure,
	}

	// sharedColumns contains every field present in full REST and CM list records.
	sharedColumns = []string{
		ColumnAddress,
		ColumnPlayers,
		ColumnMap,
		ColumnName,
		ColumnGame,
		ColumnVersion,
		ColumnGameDir,
		ColumnOS,
		ColumnGameType,
		ColumnSteamID,
		ColumnBots,
		ColumnGamePort,
		ColumnSpectatorPort,
		ColumnDedicated,
		ColumnSecure,
	}

	// allColumns appends CM Discovery metadata to the shared full-record fields.
	allColumns = append(slices.Clone(sharedColumns),
		ColumnAuthPlayers,
		ColumnSpectatorName,
		ColumnSDRPingLocation,
		ColumnFriendsCount,
		ColumnFriendsCountConnected,
		ColumnPingEstimate,
	)

	/*
		Disabled protocol-only fields retained for possible future use:

		ColumnRegion,
		ColumnRevision,
		ColumnQueryPort,
		ColumnSDRPopID,
		ColumnFlags,
		ColumnMapKnown,
		ColumnVersionKnown,
		ColumnGameTypeKnown,
	*/
)

// Row contains the normalized values used by every output format.
type Row struct {
	// Address is the server's query address.
	Address string
	// Name is the advertised server name.
	Name string
	// Map is the current map name.
	Map string
	// Game is the server's game description.
	Game string
	// Version is the advertised game version.
	Version string
	// GameDir is the game's directory identifier.
	GameDir string
	// OS is the reported server operating-system marker.
	OS string
	// GameType is the server's advertised tag string.
	GameType string
	// SteamID is the server's Steam identifier as a decimal string.
	SteamID string
	// SpectatorName is the advertised spectator server name.
	SpectatorName string
	// SDRPingLocation is the Steam Datagram Relay location.
	SDRPingLocation string
	// Region is the backend's numeric region identifier.
	Region int32
	// Bots is the reported bot count.
	Bots int32
	// GamePort is the advertised gameplay port.
	GamePort uint32
	// SpectatorPort is the optional spectator port.
	SpectatorPort uint32
	// AuthPlayers is the CM-reported authenticated player count.
	AuthPlayers uint32
	// Revision is the CM server record revision.
	Revision uint32
	// QueryPort is the query port reported by CM.
	QueryPort uint32
	// SDRPopID identifies the server's SDR point of presence.
	SDRPopID uint32
	// FriendsCount is the CM-reported friend count.
	FriendsCount uint32
	// FriendsCountConnected is the connected CM friend count.
	FriendsCountConnected uint32
	// PingEstimate is the CM-reported ping estimate.
	PingEstimate uint32
	// Flags contains opaque CM server flags.
	Flags uint32
	// Players is the current player count.
	Players int32
	// MaxPlayers is the server's slot capacity.
	MaxPlayers int32
	// Dedicated reports whether the server is dedicated.
	Dedicated bool
	// Secure reports whether the server advertises secure authentication.
	Secure bool
	// MapKnown reports whether CM resolved the map field.
	MapKnown bool
	// VersionKnown reports whether CM resolved the version field.
	VersionKnown bool
	// GameTypeKnown reports whether CM resolved the game-type field.
	GameTypeKnown bool
	// HasDetails distinguishes full records from address-only results.
	HasDetails bool
	// HasPlatformInfo marks OS, region, dedicated, and secure values as available.
	HasPlatformInfo bool
	// HasCMMetadata marks fields only exposed by the CM discovery response.
	HasCMMetadata bool
}

// DefaultColumns returns the default server fields in display order.
func DefaultColumns() []string {
	return slices.Clone(defaultColumns)
}

// ColumnChoices returns shared selectable column names and the all alias.
func ColumnChoices() []string {
	choices := slices.Clone(selectableColumns)
	return append(choices, ColumnAll)
}

// SharedColumns returns every field present in full REST and CM list records.
func SharedColumns() []string {
	return slices.Clone(sharedColumns)
}

// DiscoveryColumns returns shared fields and fields exposed by CM Discovery records.
func DiscoveryColumns() []string {
	return slices.Clone(allColumns)
}

// FormatChoices returns output formats in CLI choice order.
func FormatChoices() []string {
	return []string{
		FormatHTML,
		FormatMarkdown,
		FormatCSV,
		FormatTable,
		FormatYAML,
		FormatJSON,
	}
}

// ResolveColumns expands defaults and the all alias using the fields supported
// by the selected result type. Explicit selections remain limited to shared fields.
func ResolveColumns(requested, available []string) ([]string, error) {
	if len(requested) == 0 {
		return DefaultColumns(), nil
	}
	if slices.Contains(requested, ColumnAll) {
		if len(requested) != 1 {
			return nil, fmt.Errorf("%q cannot be combined with individual columns", ColumnAll)
		}

		return slices.Clone(available), nil
	}

	resolved := make([]string, 0, len(requested))
	seen := make(map[string]struct{}, len(requested))
	for _, name := range requested {
		if !slices.Contains(selectableColumns, name) {
			return nil, fmt.Errorf("unknown output column %q", name)
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("output column %q is repeated", name)
		}

		seen[name] = struct{}{}
		resolved = append(resolved, name)
	}

	return resolved, nil
}

// HideEmptyColumns removes fields whose values are empty or zero in every row.
func HideEmptyColumns(rows []Row, columns []string) []string {
	if len(rows) == 0 {
		return slices.Clone(columns)
	}

	kept := make([]string, 0, len(columns))
	for _, name := range columns {
		if name == ColumnAddress || columnHasValue(rows, name) {
			kept = append(kept, name)
		}
	}

	return kept
}

// columnHasValue reports whether at least one row contains a non-zero value.
func columnHasValue(rows []Row, name string) bool {
	for _, row := range rows {
		if !columnAvailable(row, name) {
			continue
		}
		if name == ColumnPlayers && row.Players == 0 && row.MaxPlayers == 0 {
			continue
		}

		if !isZeroColumnValue(columnValue(row, name)) {
			return true
		}
	}

	return false
}

// isZeroColumnValue recognizes empty strings and zero-valued output scalars.
func isZeroColumnValue(value any) bool {
	switch value := value.(type) {
	case string:
		return strings.TrimSpace(value) == ""
	case bool:
		return !value
	case int32:
		return value == 0
	case uint32:
		return value == 0
	default:
		return value == nil
	}
}

// resolveRenderColumns validates the already-expanded fields passed to Print.
func resolveRenderColumns(columns []string) ([]string, error) {
	if len(columns) == 0 {
		return DefaultColumns(), nil
	}
	if slices.Contains(columns, ColumnAll) {
		if len(columns) != 1 {
			return nil, fmt.Errorf("%q cannot be combined with individual columns", ColumnAll)
		}

		return DiscoveryColumns(), nil
	}

	resolved := make([]string, 0, len(columns))
	seen := make(map[string]struct{}, len(columns))
	for _, name := range columns {
		if !slices.Contains(allColumns, name) {
			return nil, fmt.Errorf("unknown output column %q", name)
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("output column %q is repeated", name)
		}

		seen[name] = struct{}{}
		resolved = append(resolved, name)
	}

	return resolved, nil
}

// Project converts rows to JSON/YAML-friendly objects containing only selected fields.
// The players column maps to both numeric player-count fields.
func Project(rows []Row, columns []string) []map[string]any {
	projected := make([]map[string]any, len(rows))
	for rowIndex, row := range rows {
		fields := make(map[string]any, len(columns)+1)
		for _, name := range columns {
			if !columnAvailable(row, name) {
				continue
			}

			switch name {
			case ColumnAddress:
				fields[ColumnAddress] = row.Address
			case ColumnPlayers:
				if row.HasDetails {
					fields["players"] = row.Players
					fields["max_players"] = row.MaxPlayers
				}
			case ColumnMap:
				if row.HasDetails {
					fields[ColumnMap] = row.Map
				}
			case ColumnName:
				if row.HasDetails {
					fields[ColumnName] = row.Name
				}
			case ColumnGame:
				if row.HasDetails {
					fields[ColumnGame] = row.Game
				}
			case ColumnVersion:
				if row.HasDetails {
					fields[ColumnVersion] = row.Version
				}
			case ColumnGameDir:
				if row.HasDetails {
					fields[ColumnGameDir] = row.GameDir
				}
			case ColumnOS:
				if row.HasPlatformInfo {
					fields[ColumnOS] = row.OS
				}
			case ColumnGameType:
				if row.HasDetails {
					fields[ColumnGameType] = row.GameType
				}
			case ColumnSteamID:
				if row.HasDetails {
					fields[ColumnSteamID] = row.SteamID
				}
			case ColumnBots:
				if row.HasDetails {
					fields[ColumnBots] = row.Bots
				}
			case ColumnGamePort:
				if row.HasDetails {
					fields[ColumnGamePort] = row.GamePort
				}
			case ColumnSpectatorPort:
				if row.HasDetails {
					fields[ColumnSpectatorPort] = row.SpectatorPort
				}
			case ColumnDedicated:
				if row.HasPlatformInfo {
					fields[ColumnDedicated] = row.Dedicated
				}
			case ColumnSecure:
				if row.HasPlatformInfo {
					fields[ColumnSecure] = row.Secure
				}
			case ColumnAuthPlayers:
				if row.HasDetails {
					fields[ColumnAuthPlayers] = row.AuthPlayers
				}
			case ColumnSpectatorName:
				if row.HasDetails {
					fields[ColumnSpectatorName] = row.SpectatorName
				}
			case ColumnSDRPingLocation:
				if row.HasDetails {
					fields[ColumnSDRPingLocation] = row.SDRPingLocation
				}
			case ColumnFriendsCount:
				if row.HasDetails {
					fields[ColumnFriendsCount] = row.FriendsCount
				}
			case ColumnFriendsCountConnected:
				if row.HasDetails {
					fields[ColumnFriendsCountConnected] = row.FriendsCountConnected
				}
			case ColumnPingEstimate:
				if row.HasDetails {
					fields[ColumnPingEstimate] = row.PingEstimate
				}
			}
		}

		projected[rowIndex] = fields
	}

	return projected
}

// columnAvailable prevents fields absent from a protocol response from being
// represented as valid zero values in structured output.
func columnAvailable(row Row, name string) bool {
	if !row.HasDetails {
		return name == ColumnAddress
	}
	if name == ColumnOS || name == ColumnDedicated || name == ColumnSecure {
		return row.HasPlatformInfo
	}

	switch name {
	case
		ColumnAuthPlayers,
		ColumnSpectatorName,
		ColumnSDRPingLocation,
		ColumnRevision,
		ColumnQueryPort,
		ColumnSDRPopID,
		ColumnFriendsCount,
		ColumnFriendsCountConnected,
		ColumnPingEstimate:
		return row.HasCMMetadata
	}

	return true
}
