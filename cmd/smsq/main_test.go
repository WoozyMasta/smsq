// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/woozymasta/flags"
	"github.com/woozymasta/smsq/internal/keys"
)

func TestRunRejectsLiteCLIProjection(t *testing.T) {
	var stderr bytes.Buffer
	err := run(
		[]string{"rest", "--api-key", "test", "--records", "lite"},
		&bytes.Buffer{},
		&stderr,
	)
	if err == nil {
		t.Fatal("run() accepted the module-only lite projection")
	}
}

func TestRESTIsDefaultCommand(t *testing.T) {
	commands := commandSet{}
	parser, err := newParser(&commands)
	if err != nil {
		t.Fatalf("newParser() error = %v", err)
	}

	_, err = parser.ParseArgs([]string{"--records", "invalid"})
	if err == nil ||
		!strings.Contains(err.Error(), "records") ||
		!strings.Contains(err.Error(), "Invalid value") {
		t.Fatalf("ParseArgs() error = %v, want REST option validation error", err)
	}
}

func TestGRPCRecordsDefaultToFull(t *testing.T) {
	commands := commandSet{}
	parser, err := newParser(&commands)
	if err != nil {
		t.Fatalf("newParser() error = %v", err)
	}

	grpcCommand := parser.Find("grpc")
	if grpcCommand == nil {
		t.Fatal("parser does not contain grpc command")
	}

	option := grpcCommand.FindOptionByLongName("records")
	if option == nil || option.Field().Tag.Get("default") != "full" {
		t.Fatalf("grpc --records default = %v, want full", option)
	}
}

func TestRunWithoutArgumentsShowsRootHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if err := run(nil, &stdout, &stderr); err != nil {
		t.Fatalf("run() error = %v", err)
	}

	help := stdout.String()
	if !strings.Contains(help, "smsq [OPTIONS] <command>") ||
		!strings.Contains(help, "Available commands:") ||
		strings.Contains(help, "smsq [OPTIONS] rest [rest-OPTIONS]") {
		t.Fatalf("run() did not print root help:\n%s", help)
	}
	if stderr.Len() != 0 {
		t.Fatalf("run() wrote unexpected stderr output: %s", stderr.String())
	}
}

func TestFilterFlagsEnforceExclusiveModes(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "raw and include",
			args: []string{"--app-id", "730", "--raw-filter", `\map\de_dust2`, "--filter", "map=de_dust2"},
		},
		{
			name: "raw and NOR",
			args: []string{"--app-id", "730", "--raw-filter", `\map\de_dust2`, "--nor", "password=1"},
		},
		{
			name: "raw and NAND",
			args: []string{"--app-id", "730", "--raw-filter", `\map\de_dust2`, "--nand", "password=1"},
		},
		{
			name: "NOR and NAND",
			args: []string{"--app-id", "730", "--nor", "password=1", "--nand", "secure=1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			commands := commandSet{}
			parser, err := newParser(&commands)
			if err != nil {
				t.Fatalf("newParser() error = %v", err)
			}

			args := append([]string{"rest", "--api-key", "test"}, test.args...)
			if _, err := parser.ParseArgs(args); err == nil {
				t.Fatal("ParseArgs() accepted mutually exclusive filter modes")
			}
		})
	}
}

func TestPlanInputConflictUsesDeclaredFlagNames(t *testing.T) {
	filters := filterFlags{
		Filter: []string{"map=de_dust2"},
		NOR:    []string{"password=1"},
		NAND:   []string{"secure=1"},
	}

	err := planInputConflict("plan.json", filters)
	if err == nil {
		t.Fatal("planInputConflict() returned nil")
	}

	want := "--plan-in cannot be combined with --filter, --nor, --nand"
	if err.Error() != want {
		t.Errorf("planInputConflict() = %q, want %q", err, want)
	}
}

func TestPlanInputConflictSkipsUnsetFlags(t *testing.T) {
	if err := planInputConflict("plan.json", filterFlags{}); err != nil {
		t.Fatalf("planInputConflict() = %v, want nil", err)
	}
}

func TestPlanInputRejectsRawFilter(t *testing.T) {
	filters := filterFlags{RawFilter: `\appid\730`}
	err := planInputConflict("plan.json", filters)
	if err == nil || !strings.Contains(err.Error(), "--raw-filter") {
		t.Fatalf("planInputConflict() = %v, want raw-filter conflict", err)
	}
}

func TestRawFilterDiscoveryConflict(t *testing.T) {
	filters := filterFlags{RawFilter: `\appid\730`}
	tests := []struct {
		name             string
		limit, pageLimit uint32
		plan             planFiles
		discoveryOptions bool
		wantErr          bool
	}{
		{
			name:      "direct request",
			limit:     10000,
			pageLimit: 10000,
		},
		{
			name:      "over page cap",
			limit:     10001,
			pageLimit: 10000,
			wantErr:   true,
		},
		{
			name:      "plan output",
			limit:     100,
			pageLimit: 10000,
			plan:      planFiles{PlanOut: "plan.json"},
			wantErr:   true,
		},
		{
			name:             "discovery options",
			limit:            100,
			pageLimit:        10000,
			discoveryOptions: true,
			wantErr:          true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := rawFilterDiscoveryConflict(
				filters, test.limit, test.pageLimit, test.plan, test.discoveryOptions,
			)
			if (err != nil) != test.wantErr {
				t.Fatalf("rawFilterDiscoveryConflict() error = %v, wantErr %t", err, test.wantErr)
			}
		})
	}
}

func TestHelpListsRegisteredFilterKeys(t *testing.T) {
	commands := commandSet{}
	parser, err := newParser(&commands)
	if err != nil {
		t.Fatalf("newParser() error = %v", err)
	}

	var output bytes.Buffer
	parser.WriteHelp(&output)
	help := output.String()
	if !strings.Contains(help, "Filters:\n  Use the listed keys with --filter, --nor, and --nand.\n\n") {
		t.Fatalf("help does not contain the filter-key footer:\n%s", help)
	}
	if !strings.Contains(help, "  Flag filters:") ||
		!strings.Contains(help, "    Binary flags accept 0/1, true/false, yes/no, or y/n.") ||
		!strings.Contains(help, "  Value filters:") {
		t.Fatalf("help does not group filter keys by value type:\n%s", help)
	}
	if !strings.Contains(help, "\n\n  Value filters:\n") {
		t.Fatalf("help does not separate value filters with a blank line:\n%s", help)
	}
	if strings.ContainsAny(help, "—–") {
		t.Fatalf("help contains non-ASCII dash characters:\n%s", help)
	}
	for name := range keys.Registry {
		if !strings.Contains(help, name) {
			t.Errorf("help does not list registered filter key %q", name)
		}
	}
	if _, exists := keys.Registry["collapse_addr_hash"]; exists || strings.Contains(help, "collapse_addr_hash") {
		t.Errorf("help exposes the unverified collapse_addr_hash filter:\n%s", help)
	}
	if !strings.Contains(help, "    secure=1") ||
		!strings.Contains(help, "  Flag filters:\n    Binary flags") ||
		!strings.Contains(help, "    dedicated=1") {
		t.Errorf("help does not place flag filters in their section:\n%s", help)
	}
	if strings.Contains(help, "appid=") {
		t.Errorf("help exposes AppID as a structured filter:\n%s", help)
	}
	if strings.Index(help, "notfull=1") > strings.Index(help, "  Value filters:") {
		t.Errorf("help does not place notfull with flag filters:\n%s", help)
	}
	if !strings.Contains(help, "notfull=1") {
		t.Errorf("help does not show the notfull operand:\n%s", help)
	}
	if !strings.Contains(help, "dedicated=1") {
		t.Errorf("help does not show a usable flag-filter expression:\n%s", help)
	}
}

func TestMarkdownDocumentationListsRegisteredFilterKeys(t *testing.T) {
	commands := commandSet{}
	parser, err := newParser(&commands)
	if err != nil {
		t.Fatalf("newParser() error = %v", err)
	}

	var output bytes.Buffer
	if err := parser.WriteDoc(
		&output,
		flags.DocFormatMarkdown,
		flags.WithBuiltinTemplate(flags.DocTemplateMarkdownTable),
	); err != nil {
		t.Fatalf("WriteDoc() error = %v", err)
	}

	doc := output.String()
	if !strings.Contains(doc, "## Filter keys") ||
		!strings.Contains(doc, "### Flag filters") ||
		!strings.Contains(doc, "### Value filters") {
		t.Fatalf("Markdown documentation is missing the filter catalog:\n%s", doc)
	}

	for name, info := range keys.Registry {
		entry := "* `" + name + "=" + info.Example + "`: " + info.Desc
		if !strings.Contains(doc, entry) {
			t.Errorf("Markdown documentation does not contain filter entry %q", entry)
		}
	}
}
