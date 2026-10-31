// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"time"
)

// commandRuntime carries process dependencies to the selected command.
type commandRuntime struct {
	ctx     context.Context // ctx bounds the complete operation, including authentication and discovery.
	filters *filterFlags    // filters points to parsed root-level filter options.
	query   *queryFlags     // query points to parsed root-level limit and timeout options.
	output  *outputFlags    // output points to parsed root-level formatting options.
	stdin   *os.File        // stdin supplies interactive Steam Guard input.
	stdout  io.Writer       // stdout receives result data and parser help.
	stderr  io.Writer       // stderr receives interactive authentication prompts.
}

type commandSet struct { // betteralign:ignore
	UDP  udpCommand  `command:"udp"  description:"Query server addresses through legacy master-server UDP"`
	REST restCommand `command:"rest" description:"Query server lists through the Steam Web API" default-command:"true"`
	GRPC grpcCommand `command:"grpc" description:"Query server lists through an authenticated Steam CM session"`

	Filters filterFlags `group:"Filter options"`
	Query   queryFlags  `group:"Query options"`
	Output  outputFlags `group:"Output options"`
}

type udpCommand struct {
	runtime *commandRuntime `no-flag:"true"`

	Masters   []string      `long:"master"     description:"Steam master server address (host:port)" required:"true" validate-non-empty:"true"`
	IOTimeout time.Duration `long:"io-timeout" description:"Timeout for each UDP request" default:"1s" validate-min:"1"`
	MaxPages  uint          `long:"max-pages"  description:"Maximum pages to read from each master; 0 means no page cap"`
	Region    uint8         `long:"region"     description:"Legacy geographic region code (0-255; 255 requests all regions)" default:"0"`
}

type restCommand struct { // betteralign:ignore
	runtime *commandRuntime `no-flag:"true"`

	APIKey  string `long:"api-key" description:"Steam Web API key; alternatively set STEAM_API_KEY" short:"k" secret:"true" required:"true" env:"STEAM_API_KEY"`
	Records string `long:"records" description:"Return full server records or addresses only" default:"full" choices:"full;addresses"`

	Plan      planFiles          `group:"Discovery plan files"`
	Discovery restDiscoveryFlags `group:"Discovery limits"`
}

type grpcCommand struct {
	runtime *commandRuntime `no-flag:"true"`

	Plan       planFiles           `group:"Discovery plan files"`
	Auth       grpcAuthFlags       `group:"Authentication"`
	Token      grpcTokenFlags      `group:"Token storage"`
	Discovery  grpcDiscoveryFlags  `group:"Discovery"`
	Connection grpcConnectionFlags `group:"Steam Connection Manager connection"`
}

type filterFlags struct {
	RawFilter string   `short:"r" long:"raw-filter" description:"Raw Steam filter expression; an embedded AppID takes precedence over --app-id" xor:"rf;rn;rx"`
	Filter    []string `short:"f" long:"filter"     description:"Add an include filter as key=value"                                            xor:"rf"`
	NOR       []string `short:"x" long:"nor"        description:"Exclude servers matching any supplied condition; incompatible with --nand"     xor:"rn;ex"`
	NAND      []string `short:"X" long:"nand"       description:"Exclude servers matching all supplied conditions; incompatible with --nor"     xor:"rx;ex"`
	AppID     uint32   `short:"i" long:"app-id"     description:"Steam application ID; required unless included in --raw-filter" validate-min:"1"`
}

type queryFlags struct {
	Limit   uint32        `short:"l" default:"25" long:"limit"   validate-min:"1" description:"Maximum number of unique servers to return"`
	Timeout time.Duration `short:"t" default:"2m" long:"timeout" validate-min:"1" description:"Maximum duration of the complete operation"`
}

type outputFlags struct {
	Format  string   `short:"F" long:"format"  description:"Output format" default:"table" choices:"table;md;html;csv;yaml;json"`
	Output  string   `short:"o" long:"output"  description:"Write output to a file instead of stdout" validate-writable:"true"`
	Columns []string `short:"c" long:"columns" description:"Select output fields; all includes fields available for this protocol and hides empty or zero fields" defaults:"address;players;map;name"`
	Pretty  bool     `short:"p" long:"pretty"  description:"Indent JSON output; YAML uses block format by default"`
}

type planFiles struct {
	PlanIn  string `long:"plan-in"  description:"Collect results using a saved Discovery plan" validate-readable:"true" validate-existing-file:"true"`
	PlanOut string `long:"plan-out" description:"Save the generated Discovery plan; only applies when Discovery runs" validate-writable:"true"`
}

type restDiscoveryFlags struct {
	MaxRequests        int `validate-min:"0" long:"max-requests"         description:"Maximum Web API requests used by Discovery; 0 derives the budget from --limit"`
	MaxPlanQueries     int `validate-min:"0" long:"max-plan-queries"     description:"Maximum query leaves in the plan; 0 derives the value from --limit"`
	MaxBufferedRecords int `validate-min:"0" long:"max-buffered-records" description:"Maximum sample records retained for split validation; 0 uses the default"`
	MaxDepth           int `validate-min:"0" long:"max-depth"            description:"Maximum recursive split depth; 0 uses the default"`
	MinValueCount      int `validate-min:"0" long:"min-value-count"      description:"Ignore split values with fewer sample records; 0 uses the default"`
}

type grpcAuthFlags struct {
	Mode       string `long:"auth-mode"   description:"Steam authentication: QR approval, Steam Guard code, or mobile approval" short:"m" default:"qr" choices:"qr;code;approve"`
	Account    string `long:"account"     description:"Steam account name for password-based login"                             short:"a" env:"STEAM_ACCOUNT"`
	DeviceName string `long:"device-name" description:"Device name shown during Steam login"                                              default:"smsq" validate-min-len:"1" validate-max-len:"128"`
}

type grpcTokenFlags struct {
	File string `long:"token-file" env:"STEAM_TOKEN_FILE" validate-writable:"true" description:"Read and update a plaintext refresh-token file; omit to keep the token in memory"`
}

type grpcDiscoveryFlags struct {
	Records     string `long:"records"      description:"Return full server records or addresses only" default:"full" choices:"full;addresses"`
	MaxRequests int    `long:"max-requests" description:"Maximum CM Discovery requests; 0 derives the budget from --limit" validate-min:"0"`
	MaxDepth    int    `long:"max-depth"    description:"Maximum recursive Discovery split depth; 0 uses the default"      validate-min:"0"`
}

type grpcConnectionFlags struct {
	Endpoint        string        `long:"cm-endpoint"      description:"Use this Steam CM host:port instead of looking up endpoints from Steam"`
	RequestInterval time.Duration `long:"request-interval" description:"Minimum delay between Discovery request starts after the initial burst; 0 disables pacing" validate-min:"0"`
	InitialBurst    int           `long:"initial-burst"    description:"Number of initial Discovery requests sent without pacing; 0 disables the burst"            validate-min:"0"`
}

// planInputConflict reports explicitly supplied query flags that conflict with a saved plan,
// using the long names declared on their flag structs.
func planInputConflict(plan string, filters filterFlags) error {
	if plan == "" {
		return nil
	}

	conflicts := []string{longFlagName(planFiles{}, "PlanIn")}
	if filters.RawFilter != "" {
		conflicts = append(conflicts, longFlagName(filterFlags{}, "RawFilter"))
	}

	if len(filters.Filter) > 0 {
		conflicts = append(conflicts, longFlagName(filterFlags{}, "Filter"))
	}
	if len(filters.NOR) > 0 {
		conflicts = append(conflicts, longFlagName(filterFlags{}, "NOR"))
	}
	if len(filters.NAND) > 0 {
		conflicts = append(conflicts, longFlagName(filterFlags{}, "NAND"))
	}
	if len(conflicts) == 1 {
		return nil
	}

	return fmt.Errorf("%s cannot be combined with %s", conflicts[0], strings.Join(conflicts[1:], ", "))
}

// rawFilterDiscoveryConflict rejects raw expressions when a command will plan queries.
func rawFilterDiscoveryConflict(
	filters filterFlags,
	limit, pageLimit uint32,
	plan planFiles,
	discoveryOptions bool,
) error {
	if filters.RawFilter == "" ||
		(limit <= pageLimit && plan.PlanIn == "" && plan.PlanOut == "" && !discoveryOptions) {
		return nil
	}

	return errors.New("--raw-filter cannot be used with discovery or discovery plan options")
}

// longFlagName reads the CLI spelling from the option struct's long tag.
func longFlagName(options any, fieldName string) string {
	field, ok := reflect.TypeOf(options).FieldByName(fieldName)
	if ok {
		if name := field.Tag.Get("long"); name != "" {
			return "--" + name
		}
	}

	return "--" + fieldName
}
