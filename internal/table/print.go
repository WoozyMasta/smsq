// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package table

import (
	"fmt"
	"html"
	"io"
	"strconv"
	"strings"
	"unicode"

	pretty "github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// Print renders the selected fields using a go-pretty table format. The row
// number is always prepended for tabular formats and is not a selectable field.
func Print(output io.Writer, rows []Row, format string, columns []string, filter string) error {
	writer, err := newWriter(rows, format, columns)
	if err != nil {
		return err
	}

	var rendered string
	switch format {
	case FormatHTML:
		rendered = renderHTMLReport(writer.RenderHTML(), filter)
	case FormatMarkdown:
		rendered = renderMarkdownReport(writer.RenderMarkdown(), filter)
	case FormatCSV:
		rendered = writer.RenderCSV()
	case FormatTable:
		rendered = writer.Render()
	default:
		return fmt.Errorf("unsupported table format %q", format)
	}

	if rendered == "" {
		return nil
	}
	if _, err := io.WriteString(output, rendered); err != nil {
		return fmt.Errorf("write rendered table: %w", err)
	}

	if rendered[len(rendered)-1] != '\n' {
		if _, err := io.WriteString(output, "\n"); err != nil {
			return fmt.Errorf("write rendered table newline: %w", err)
		}
	}

	return nil
}

// renderHTMLReport wraps the table in a standalone, readable HTML document.
func renderHTMLReport(renderedTable, filter string) string {
	const title = "Steam Master Server Query CLI report"

	var report strings.Builder
	report.WriteString(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>` + title + `</title>
  <style>
    body { margin: 2rem auto; padding: 0 1rem; max-width: 96rem;
      font: 15px/1.5 system-ui, sans-serif; color: #24292f; background: #fff; }
    h1 { margin-bottom: .5rem; font-size: 1.5rem; }
    .request { margin: 0 0 1.25rem; color: #57606a; }
    .table-wrap { overflow-x: auto; }
    table.go-pretty-table { width: 100%; border-collapse: collapse; }
    .go-pretty-table th, .go-pretty-table td {
      padding: .4rem .6rem; border-bottom: 1px solid #d0d7de; text-align: left; }
    .go-pretty-table th { background: #f6f8fa; }
    .go-pretty-table tbody tr:nth-child(even) { background: #f6f8fa; }
  </style>
</head>
<body>
  <main>
    <h1>` + title + `</h1>
    <p class="request">Request: <code>`)
	report.WriteString(html.EscapeString(displayFilter(filter)))
	report.WriteString(`</code></p>
    <div class="table-wrap">
`)
	report.WriteString(renderedTable)
	if !strings.HasSuffix(renderedTable, "\n") {
		report.WriteByte('\n')
	}
	report.WriteString(`    </div>
  </main>
</body>
</html>
`)

	return report.String()
}

// renderMarkdownReport adds a title and readable filter expression above the table.
func renderMarkdownReport(renderedTable, filter string) string {
	return "# Steam Master Server Query CLI report\n\nrequest: " +
		markdownCodeSpan(displayFilter(filter)) + "\n\n" + renderedTable
}

// displayFilter uses slash-separated notation for Steam's backslash-delimited filter.
func displayFilter(filter string) string {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return "(not available)"
	}

	return "/" + strings.TrimLeft(strings.ReplaceAll(filter, `\`, "/"), "/")
}

// markdownCodeSpan removes control characters and chooses a safe code delimiter.
func markdownCodeSpan(value string) string {
	value = strings.Map(func(character rune) rune {
		if unicode.IsControl(character) {
			return ' '
		}
		return character
	}, value)
	value = strings.TrimSpace(value)

	delimiter := "`"
	for strings.Contains(value, delimiter) {
		delimiter += "`"
	}
	if strings.HasPrefix(value, "`") || strings.HasSuffix(value, "`") {
		return delimiter + " " + value + " " + delimiter
	}

	return delimiter + value + delimiter
}

// newWriter builds a go-pretty table with columns in the requested order.
func newWriter(rows []Row, format string, columns []string) (pretty.Writer, error) {
	resolved, err := resolveRenderColumns(columns)
	if err != nil {
		return nil, fmt.Errorf("resolve output columns: %w", err)
	}

	writer := pretty.NewWriter()
	writer.SetStyle(pretty.StyleRounded)
	writer.Style().HTML.EscapeText = true
	writer.SuppressTrailingSpaces()

	header := pretty.Row{"#"}
	for _, name := range resolved {
		header = append(header, columnHeader(name))
	}

	writer.AppendHeader(header)

	for index, row := range rows {
		values := pretty.Row{index + 1}
		for _, name := range resolved {
			value := columnValue(row, name)
			if format == FormatMarkdown && (name == ColumnName || name == ColumnGameType) {
				value = markdownCodeSpan(fmt.Sprint(value))
			}
			values = append(values, value)
		}
		writer.AppendRow(values)
	}

	if format == FormatTable {
		writer.SetColumnConfigs(nameColumnConfigs(resolved))
	}

	return writer, nil
}

// columnHeader returns the stable display label for a selectable field.
func columnHeader(name string) string {
	switch name {
	case ColumnAddress:
		return "ADDRESS"
	case ColumnPlayers:
		return "PLAYERS"
	case ColumnMap:
		return "MAP"
	case ColumnName:
		return "NAME"
	case ColumnGame:
		return "GAME"
	case ColumnVersion:
		return "VERSION"
	case ColumnGameDir:
		return "GAME DIR"
	case ColumnOS:
		return "OS"
	case ColumnGameType:
		return "GAME TYPE"
	case ColumnSteamID:
		return "STEAM ID"
	case ColumnBots:
		return "BOTS"
	case ColumnGamePort:
		return "GAME PORT"
	case ColumnSpectatorPort:
		return "SPEC PORT"
	case ColumnDedicated:
		return "DEDICATED"
	case ColumnSecure:
		return "SECURE"
	case ColumnAuthPlayers:
		return "AUTH PLAYERS"
	case ColumnSpectatorName:
		return "SPEC NAME"
	case ColumnSDRPingLocation:
		return "SDR LOC"
	case ColumnFriendsCount:
		return "FRIENDS"
	case ColumnFriendsCountConnected:
		return "FRIENDS CONN"
	case ColumnPingEstimate:
		return "PING EST."

		/*
			Disabled technical fields retained for reference:

			case ColumnRegion:
				return "REGION"
			case ColumnRevision:
				return "REVISION"
			case ColumnQueryPort:
				return "QUERY PORT"
			case ColumnSDRPopID:
				return "SDR POP ID"
			case ColumnFlags:
				return "FLAGS"
			case ColumnMapKnown:
				return "MAP KNOWN"
			case ColumnVersionKnown:
				return "VERSION KNOWN"
			case ColumnGameTypeKnown:
				return "GAME TYPE KNOWN"
		*/

	default:
		return name
	}
}

// columnValue formats one row field for human-readable renderers.
func columnValue(row Row, name string) any {
	if !columnAvailable(row, name) {
		return ""
	}

	switch name {
	case ColumnAddress:
		return row.Address
	case ColumnPlayers:
		return playerLabel(row)
	case ColumnMap:
		return row.Map
	case ColumnName:
		return row.Name
	case ColumnGame:
		return row.Game
	case ColumnVersion:
		return row.Version
	case ColumnGameDir:
		return row.GameDir
	case ColumnOS:
		return row.OS
	case ColumnGameType:
		return row.GameType
	case ColumnSteamID:
		return row.SteamID
	case ColumnBots:
		return row.Bots
	case ColumnGamePort:
		return row.GamePort
	case ColumnSpectatorPort:
		return row.SpectatorPort
	case ColumnDedicated:
		return row.Dedicated
	case ColumnSecure:
		return row.Secure
	case ColumnAuthPlayers:
		return row.AuthPlayers
	case ColumnSpectatorName:
		return row.SpectatorName
	case ColumnSDRPingLocation:
		return row.SDRPingLocation
	case ColumnFriendsCount:
		return row.FriendsCount
	case ColumnFriendsCountConnected:
		return row.FriendsCountConnected
	case ColumnPingEstimate:
		return row.PingEstimate

		/*
			Disabled technical fields retained for reference:

			case ColumnRegion:
				return row.Region
			case ColumnRevision:
				return row.Revision
			case ColumnQueryPort:
				return row.QueryPort
			case ColumnSDRPopID:
				return row.SDRPopID
			case ColumnFlags:
				return row.Flags
			case ColumnMapKnown:
				return row.MapKnown
			case ColumnVersionKnown:
				return row.VersionKnown
			case ColumnGameTypeKnown:
				return row.GameTypeKnown
		*/

	default:
		return ""
	}
}

// nameColumnConfigs caps long names only in the terminal layout.
func nameColumnConfigs(columns []string) []pretty.ColumnConfig {
	for index, name := range columns {
		if name == ColumnName {
			return []pretty.ColumnConfig{{
				Number:           index + 2,
				WidthMax:         64,
				WidthMaxEnforcer: text.Trim,
			}}
		}
	}

	return nil
}

// playerLabel formats player and slot counts for tabular output.
func playerLabel(row Row) string {
	return strconv.FormatInt(int64(row.Players), 10) + "/" +
		strconv.FormatInt(int64(row.MaxPlayers), 10)
}
