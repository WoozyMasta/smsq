// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package table

import (
	"bytes"
	"slices"
	"strings"
	"testing"
)

func TestResolveColumns(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    []string
		wantErr bool
	}{
		{name: "default", want: DefaultColumns()},
		{name: "preserves order", input: []string{ColumnName, ColumnAddress}, want: []string{ColumnName, ColumnAddress}},
		{name: "all alias", input: []string{ColumnAll}, want: selectableColumns},
		{name: "all with another column", input: []string{ColumnAll, ColumnMap}, wantErr: true},
		{name: "unknown", input: []string{"unknown"}, wantErr: true},
		{name: "duplicate", input: []string{ColumnMap, ColumnMap}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ResolveColumns(test.input, selectableColumns)
			if (err != nil) != test.wantErr {
				t.Fatalf("ResolveColumns() error = %v, wantErr %t", err, test.wantErr)
			}
			if err == nil && !equalStrings(got, test.want) {
				t.Errorf("ResolveColumns() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestNonSelectableColumnsRequireAllAlias(t *testing.T) {
	for _, hidden := range []string{
		ColumnGame,
		ColumnDedicated,
		ColumnAuthPlayers,
		ColumnRegion,
		ColumnSteamID,
		ColumnSpectatorPort,
	} {
		if _, err := ResolveColumns([]string{hidden}, selectableColumns); err == nil {
			t.Errorf("ResolveColumns() accepted hidden column %q", hidden)
		}
	}
	if !equalStrings(ColumnChoices()[len(ColumnChoices())-1:], []string{ColumnAll}) {
		t.Fatal("ColumnChoices() must expose all after shared fields")
	}
	all, err := ResolveColumns([]string{ColumnAll}, allColumns)
	if err != nil {
		t.Fatalf("ResolveColumns(all) error = %v", err)
	}
	if !slices.Contains(all, ColumnAuthPlayers) {
		t.Fatal("all columns do not include CM-specific fields")
	}
	for _, retained := range []string{
		ColumnGame,
		ColumnDedicated,
		ColumnSteamID,
		ColumnSpectatorPort,
	} {
		if !slices.Contains(all, retained) {
			t.Errorf("all columns do not include %q", retained)
		}
	}
	if slices.Contains(all, ColumnRegion) {
		t.Fatal("all columns include disabled region data")
	}
}

func TestResolveAllUsesColumnsAvailableForResult(t *testing.T) {
	tests := []struct {
		name      string
		available []string
		want      []string
	}{
		{name: "address-only", available: []string{ColumnAddress}, want: []string{ColumnAddress}},
		{name: "shared records", available: SharedColumns(), want: sharedColumns},
		{name: "CM discovery", available: DiscoveryColumns(), want: allColumns},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := ResolveColumns([]string{ColumnAll}, test.available)
			if err != nil {
				t.Fatalf("ResolveColumns(all) error = %v", err)
			}
			if !equalStrings(got, test.want) {
				t.Errorf("ResolveColumns(all) = %v, want %v", got, test.want)
			}
		})
	}
}

func TestHideEmptyColumns(t *testing.T) {
	columns := []string{
		ColumnAddress, ColumnPlayers, ColumnBots, ColumnSpectatorPort,
		ColumnGameType, ColumnSecure,
	}
	rows := []Row{{
		Address:         "192.0.2.1:27015",
		MaxPlayers:      16,
		HasDetails:      true,
		HasPlatformInfo: true,
	}}

	got := HideEmptyColumns(rows, columns)
	want := []string{ColumnAddress, ColumnPlayers}
	if !slices.Equal(got, want) {
		t.Fatalf("HideEmptyColumns() = %v, want %v", got, want)
	}
	if got := HideEmptyColumns(nil, columns); !slices.Equal(got, columns) {
		t.Fatalf("HideEmptyColumns(empty) = %v, want unchanged %v", got, columns)
	}
}

func TestColumnHeadersStayCompact(t *testing.T) {
	if got := columnHeader(ColumnPlayers); got != "PLAYERS" {
		t.Fatalf("columnHeader(%q) = %q, want PLAYERS", ColumnPlayers, got)
	}

	for _, column := range allColumns {
		if header := columnHeader(column); len(header) > 12 {
			t.Errorf("columnHeader(%q) = %q (%d chars), want at most 12", column, header, len(header))
		}
	}
}

func TestProjectSelectedFields(t *testing.T) {
	rows := []Row{
		{
			Address:    "192.0.2.1:27015",
			Players:    4,
			MaxPlayers: 32,
			Map:        "de_dust2",
			Name:       "Example",
			Game:       "Counter-Strike 2",
			Version:    "1.29",
			HasDetails: true,
		},
		{Address: "192.0.2.2:27015"},
	}
	projected := Project(rows, []string{ColumnAddress, ColumnPlayers, ColumnMap})
	if len(projected) != 2 {
		t.Fatalf("Project() returned %d records, want 2", len(projected))
	}
	if len(projected[0]) != 4 || projected[0]["max_players"] != int32(32) {
		t.Errorf("Project() full record = %#v, want selected fields and slot count", projected[0])
	}
	if len(projected[1]) != 1 || projected[1][ColumnAddress] != "192.0.2.2:27015" {
		t.Errorf("Project() address-only record = %#v, want address only", projected[1])
	}
}

func TestProjectAllColumnsIncludesEveryRegisteredField(t *testing.T) {
	projected := Project([]Row{{
		HasDetails:      true,
		HasPlatformInfo: true,
		HasCMMetadata:   true,
	}}, allColumns)
	if got, want := len(projected[0]), len(allColumns)+1; got != want {
		t.Fatalf("Project(all) returned %d fields, want %d", got, want)
	}

	for _, column := range allColumns {
		if column == ColumnPlayers {
			for _, field := range []string{"players", "max_players"} {
				if _, ok := projected[0][field]; !ok {
					t.Errorf("Project(all) omitted %q", field)
				}
			}
			continue
		}
		if _, ok := projected[0][column]; !ok {
			t.Errorf("Project(all) omitted %q", column)
		}
	}
}

func TestPrintFormatsIncludeLeadingNumber(t *testing.T) {
	rows := []Row{{
		Address:    "192.0.2.1:27015",
		Name:       "<server>",
		Map:        "map-a",
		Players:    2,
		MaxPlayers: 16,
		HasDetails: true,
	}}
	formats := []struct {
		name    string
		want    string
		wantRow string
	}{
		{name: FormatHTML, want: "<th align=\"right\">#</th>", wantRow: "<td align=\"right\">1</td>"},
		{name: FormatMarkdown, want: "| # | ADDRESS |", wantRow: "| 1 |"},
		{name: FormatCSV, want: "#,ADDRESS", wantRow: "1,"},
		{name: FormatTable, want: "╭", wantRow: "1"},
	}

	for _, format := range formats {
		t.Run(format.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := Print(
				&output, rows, format.name, []string{ColumnAddress, ColumnName}, `appid\730`,
			); err != nil {
				t.Fatalf("Print() error = %v", err)
			}
			if !strings.Contains(output.String(), format.want) {
				t.Errorf("Print() output = %q, want it to contain %q", output.String(), format.want)
			}
			if !strings.Contains(output.String(), format.wantRow) {
				t.Errorf("Print() output = %q, want row number %q", output.String(), format.wantRow)
			}
			if format.name == FormatHTML && !strings.Contains(output.String(), "&lt;server&gt;") {
				t.Errorf("RenderHTML() did not escape server text: %q", output.String())
			}
			if format.name == FormatHTML {
				for _, want := range []string{
					"<!doctype html>",
					"<title>Steam Master Server Query CLI report</title>",
					"<h1>Steam Master Server Query CLI report</h1>",
					"Request: <code>/appid/730</code>",
					"<style>",
					"</html>",
				} {
					if !strings.Contains(output.String(), want) {
						t.Errorf("HTML report %q does not contain %q", output.String(), want)
					}
				}
			}
			if format.name == FormatMarkdown &&
				!strings.HasPrefix(output.String(), "# Steam Master Server Query CLI report\n\nrequest: `/appid/730`\n\n") {
				t.Errorf("Markdown report has no title and request header: %q", output.String())
			}
		})
	}
}

func TestMarkdownWrapsNameAndGameTypeInCodeSpans(t *testing.T) {
	var output bytes.Buffer
	rows := []Row{{
		Address:    "192.0.2.1:27015",
		Name:       "War Torn |US|HELI|EVENT|\r",
		GameType:   "battleye,external|modded\r",
		HasDetails: true,
	}}
	if err := Print(
		&output,
		rows,
		FormatMarkdown,
		[]string{ColumnName, ColumnGameType},
		`appid\730`,
	); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	for _, want := range []string{
		"`War Torn \\|US\\|HELI\\|EVENT\\|`",
		"`battleye,external\\|modded`",
	} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("Markdown output %q does not contain %q", output.String(), want)
		}
	}
}

func equalStrings(left, right []string) bool {
	return slices.Equal(left, right)
}
