// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package main

import (
	"bytes"
	"net/netip"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/woozymasta/smsq/internal/table"
	"github.com/woozymasta/steamcm/gameservers"
	slist "github.com/woozymasta/steamserverlist"
)

func TestOutputChoicesMatchSupportedFormatsAndColumns(t *testing.T) {
	fields := reflect.TypeFor[outputFlags]()
	commands := commandSet{}
	parser, err := newParser(&commands)
	if err != nil {
		t.Fatalf("newParser() error = %v", err)
	}

	columnsOption := parser.FindOptionByLongName("columns")
	if columnsOption == nil {
		t.Fatal("parser has no --columns option")
	}
	if !slices.Equal(columnsOption.Choices, table.ColumnChoices()) {
		t.Errorf("--columns choices = %v, want %v", columnsOption.Choices, table.ColumnChoices())
	}

	formatOption := parser.FindOptionByLongName("format")
	if formatOption == nil {
		t.Fatal("parser has no --format option")
	}
	if !slices.Equal(formatOption.Choices, table.FormatChoices()) {
		t.Errorf("--format choices = %v, want %v", formatOption.Choices, table.FormatChoices())
	}

	limit, ok := reflect.TypeFor[queryFlags]().FieldByName("Limit")
	if !ok {
		t.Fatal("queryFlags has no Limit field")
	}
	if got := limit.Tag.Get("default"); got != "25" {
		t.Errorf("limit default = %q, want 25", got)
	}

	columnField, ok := fields.FieldByName("Columns")
	if !ok {
		t.Fatal("outputFlags has no Columns field")
	}
	if got := strings.Split(columnField.Tag.Get("defaults"), ";"); !slices.Equal(got, table.DefaultColumns()) {
		t.Errorf("column defaults = %v, want %v", got, table.DefaultColumns())
	}
}

func TestOutputJSONProjectsSelectedColumns(t *testing.T) {
	var output bytes.Buffer
	document := resultDocument{Protocol: "rest", Count: 1}
	rows := []table.Row{{
		Address:    "192.0.2.1:27015",
		Name:       "Example",
		Map:        "map-a",
		Players:    3,
		MaxPlayers: 24,
		HasDetails: true,
	}}

	err := outputResult(outputFlags{
		Format:  "json",
		Columns: []string{table.ColumnAddress, table.ColumnPlayers},
	}, document, rows, restAvailableColumns(slist.Servers{}), &output)
	if err != nil {
		t.Fatalf("outputResult() error = %v", err)
	}

	got := output.String()
	for _, want := range []string{`"address":"192.0.2.1:27015"`, `"players":3`, `"max_players":24`} {
		if !strings.Contains(got, want) {
			t.Errorf("JSON output %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, `"name"`) || strings.Contains(got, `"number"`) {
		t.Errorf("JSON output includes an unselected field or row number: %s", got)
	}
}

func TestOutputAddressOnlyProtocolForcesAddressColumn(t *testing.T) {
	rows := []table.Row{{Address: "192.0.2.1:27015"}}
	for _, format := range []string{table.FormatTable, table.FormatJSON} {
		t.Run(format, func(t *testing.T) {
			var output bytes.Buffer
			err := outputResult(outputFlags{
				Format:  format,
				Columns: []string{table.ColumnMap, table.ColumnName},
			}, resultDocument{Protocol: "udp", Count: 1}, rows,
				[]string{table.ColumnAddress}, &output,
			)
			if err != nil {
				t.Fatalf("outputResult() error = %v", err)
			}

			got := output.String()
			if !strings.Contains(got, "192.0.2.1:27015") || strings.Contains(got, "MAP") ||
				strings.Contains(got, `"map"`) || strings.Contains(got, `"name"`) {
				t.Fatalf("address-only output retained unavailable columns: %s", got)
			}
		})
	}
}

func TestOutputRESTAllOmitsCMOnlyFields(t *testing.T) {
	var output bytes.Buffer
	document := resultDocument{Protocol: "rest", Count: 1}
	rows := []table.Row{{
		Address:         "192.0.2.1:27015",
		Name:            "Example server",
		Game:            "Counter-Strike 2",
		HasDetails:      true,
		HasPlatformInfo: true,
	}}

	err := outputResult(outputFlags{
		Format:  table.FormatJSON,
		Columns: []string{table.ColumnAll},
	}, document, rows, restAvailableColumns(slist.Servers{}), &output)
	if err != nil {
		t.Fatalf("outputResult(all) error = %v", err)
	}

	got := output.String()
	if !strings.Contains(got, `"game":"Counter-Strike 2"`) {
		t.Errorf("JSON output omitted shared fields: %s", got)
	}
	if strings.Contains(got, "auth_players") || strings.Contains(got, "spectator_name") {
		t.Errorf("REST output contains CM-only fields: %s", got)
	}
}

func TestOutputAllHidesEmptyColumns(t *testing.T) {
	var output bytes.Buffer
	rows := []table.Row{{
		Address:         "192.0.2.1:27015",
		MaxPlayers:      16,
		HasDetails:      true,
		HasPlatformInfo: true,
	}}
	err := outputResult(outputFlags{
		Format:  table.FormatJSON,
		Columns: []string{table.ColumnAll},
	}, resultDocument{Protocol: "rest", Count: 1}, rows, table.SharedColumns(), &output)
	if err != nil {
		t.Fatalf("outputResult(all) error = %v", err)
	}

	got := output.String()
	if strings.Contains(got, "specport") || strings.Contains(got, "bots") || strings.Contains(got, "gametype") {
		t.Errorf("JSON all output retained zero or empty columns: %s", got)
	}
	if strings.Contains(got, "appid") {
		t.Errorf("JSON all output contains AppID: %s", got)
	}
	if !strings.Contains(got, `"address":"192.0.2.1:27015"`) ||
		!strings.Contains(got, `"players":0`) || !strings.Contains(got, `"max_players":16`) {
		t.Errorf("JSON all output omitted populated fields: %s", got)
	}
}

func TestProtocolAvailableColumnsMatchResultShape(t *testing.T) {
	if got := restAvailableColumns([]netip.AddrPort{}); !slices.Equal(
		got,
		[]string{table.ColumnAddress},
	) {
		t.Errorf("REST address columns = %v, want address only", got)
	}
	if got := grpcAvailableColumns([]gameservers.Endpoint{}); !slices.Equal(
		got,
		[]string{table.ColumnAddress},
	) {
		t.Errorf("CM endpoint columns = %v, want address only", got)
	}
	if got := grpcAvailableColumns([]gameservers.DiscoveryServer{}); !slices.Equal(
		got,
		table.DiscoveryColumns(),
	) {
		t.Errorf("CM Discovery columns = %v, want all Discovery fields", got)
	}
}

func TestOutputYAMLUsesYAMLTagsAndSelectedColumns(t *testing.T) {
	var output bytes.Buffer
	document := resultDocument{
		Protocol:  "grpc",
		PlanFile:  "plan.json",
		Count:     1,
		Requests:  2,
		ElapsedMS: 12,
	}
	document.Status = gameservers.DiscoveryStatus{
		Requests:   2,
		Unresolved: 1,
		Complete:   false,
	}
	rows := []table.Row{{
		Address:    "192.0.2.1:27015",
		Name:       "Example",
		Map:        "map-a",
		Players:    3,
		MaxPlayers: 24,
		HasDetails: true,
	}}

	err := outputResult(outputFlags{
		Format:  "yaml",
		Columns: []string{table.ColumnAddress, table.ColumnPlayers},
	}, document, rows, table.SharedColumns(), &output)
	if err != nil {
		t.Fatalf("outputResult() error = %v", err)
	}

	got := output.String()
	for _, want := range []string{
		"protocol: grpc",
		"plan_file: plan.json",
		"count: 1",
		"requests: 2",
		"unresolved: 1",
		"complete: false",
		"elapsed_ms: 12",
		"address: 192.0.2.1:27015",
		"players: 3",
		"max_players: 24",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("YAML output %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "\n    name:") || strings.Contains(got, "number:") {
		t.Errorf("YAML output includes an unselected field or row number: %s", got)
	}
}

func TestGRPCRowsKeepFullListMetadata(t *testing.T) {
	rows := grpcRows([]gameservers.Server{{
		Addr:      netip.MustParseAddrPort("192.0.2.1:27015"),
		OS:        "l",
		Region:    1,
		Dedicated: true,
		Secure:    true,
	}})
	if len(rows) != 1 {
		t.Fatalf("grpcRows() returned %d rows, want 1", len(rows))
	}
	row := rows[0]
	if !row.HasPlatformInfo || row.OS != "l" || row.Region != 1 || !row.Dedicated || !row.Secure {
		t.Errorf("grpcRows() lost CM server metadata: %#v", row)
	}
}

func TestRESTRowsConvertServerlistCollection(t *testing.T) {
	servers := slist.Servers{{
		Addr:       netip.MustParseAddrPort("192.0.2.1:27015"),
		Name:       "Example server",
		Map:        "de_dust2",
		Players:    3,
		MaxPlayers: 24,
	}}

	rows := restRows(servers)
	if len(rows) != 1 {
		t.Fatalf("restRows() returned %d rows, want 1", len(rows))
	}
	if rows[0].Address != "192.0.2.1:27015" || rows[0].Name != "Example server" ||
		rows[0].Map != "de_dust2" || rows[0].Players != 3 || rows[0].MaxPlayers != 24 {
		t.Errorf("restRows() = %#v, want projected server fields", rows[0])
	}
}

func TestRESTRowsConvertDiscoverySlice(t *testing.T) {
	servers := []slist.Server{{
		Addr:       netip.MustParseAddrPort("192.0.2.1:27015"),
		Name:       "Discovery result",
		Map:        "de_dust2",
		Players:    2,
		MaxPlayers: 16,
	}}

	rows := restRows(servers)
	if len(rows) != 1 {
		t.Fatalf("restRows() returned %d rows, want 1", len(rows))
	}
	if rows[0].Address != "192.0.2.1:27015" || rows[0].Name != "Discovery result" ||
		rows[0].Map != "de_dust2" || rows[0].Players != 2 || rows[0].MaxPlayers != 16 {
		t.Errorf("restRows() = %#v, want projected Discovery fields", rows[0])
	}
	if got := restAvailableColumns(servers); !slices.Equal(got, table.SharedColumns()) {
		t.Errorf("restAvailableColumns() = %v, want shared fields", got)
	}

	var output bytes.Buffer
	err := outputResult(outputFlags{
		Format:  table.FormatMarkdown,
		Columns: []string{table.ColumnAll},
	}, resultDocument{Protocol: "rest", Count: 1}, rows, restAvailableColumns(servers), &output)
	if err != nil {
		t.Fatalf("outputResult() error = %v", err)
	}
	if !strings.Contains(output.String(), "Discovery result") {
		t.Fatalf("Markdown Discovery output contains no server row: %s", output.String())
	}
}
