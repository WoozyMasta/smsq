// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

// Package main implements the smsq command-line client for Steam server lists.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/woozymasta/flags"
	"github.com/woozymasta/smsq/internal/table"
)

var (
	version       = "dev"
	commit        = "unknown"
	buildTime     = time.Unix(0, 0)
	repositoryURL = "https://github.com/WoozyMasta/smsq"
	_buildTime    string
)

// main runs the CLI and reports failures to stderr.
func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		os.Exit(1)
	}
}

// run parses arguments; the selected command executes through flags.Commander.
func run(args []string, stdout, stderr io.Writer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	commands := commandSet{}
	runtime := &commandRuntime{
		ctx:     ctx,
		filters: &commands.Filters,
		query:   &commands.Query,
		output:  &commands.Output,
		stdin:   os.Stdin,
		stdout:  stdout,
		stderr:  stderr,
	}
	commands.UDP.runtime = runtime
	commands.REST.runtime = runtime
	commands.GRPC.runtime = runtime

	parser, err := newParser(&commands)
	if err != nil {
		if _, writeErr := fmt.Fprintln(stderr, "smsq:", err); writeErr != nil {
			return errors.Join(err, fmt.Errorf("write parser error to stderr: %w", writeErr))
		}

		return err
	}

	if len(args) == 0 {
		parser.WriteHelp(stdout)
		return nil
	}

	if _, err := parser.ParseArgs(args); err != nil {
		if parserErr, ok := errors.AsType[*flags.Error](err); ok {
			if parserErr.Type == flags.ErrHelp || parserErr.Type == flags.ErrVersion {
				return nil
			}
		}
		return err
	}

	return nil
}

// newParser configures shared flags, protocol commands, and generated examples.
func newParser(commands *commandSet) (*flags.Parser, error) {
	parser := flags.NewParser(commands,
		flags.Default|
			flags.DetectShellFlagStyle|
			flags.DetectShellEnvStyle|
			flags.PrintHelpOnInputErrors|
			flags.HelpCommand|
			flags.VersionCommand|
			flags.CompletionCommand|
			flags.DocsCommand|
			flags.VersionFlag|
			flags.AutoShowChoiceListInHelp|
			flags.StrictPositionalArgs|
			flags.ShowRepeatableInHelp,
	)

	columnsOption := parser.FindOptionByLongName("columns")
	if columnsOption == nil {
		return nil, errors.New("find --columns option")
	}

	columnsOption.SetChoices(table.ColumnChoices()...)
	formatOption := parser.FindOptionByLongName("format")
	if formatOption == nil {
		return nil, errors.New("find --format option")
	}

	formatOption.SetChoices(table.FormatChoices()...)

	parser.Name = "smsq"
	parser.LongDescription = "Find Steam game servers through the Web API, legacy master-server UDP, " +
		"or an authenticated Steam Connection Manager session.\n" +
		"The Web API command is the default and is recommended for most queries."
	parser.SetCommandLongDescriptions(map[string]string{
		"udp": "Query one or more legacy Steam master servers.\n" +
			"Results are deduplicated across masters; collection stops when the servers are exhausted, " +
			"--limit is reached, or --max-pages stops a master.",
		"rest": "List Steam game servers through IGameServersService/GetServerList.\n" +
			"One request returns at most 10,000 records. " +
			"Larger limits, Discovery options, or a saved plan use the Discovery workflow.",
		"grpc": "Query Steam's game-server browser through an authenticated Steam Connection Manager session " +
			"using protobuf over WebSocket.\nMasterServerQuery returns server addresses and available metadata. " +
			"Despite the command name, this does not use standard gRPC.",
	})
	parser.SetVersionInfo(flags.VersionInfo{
		Version:      version,
		Revision:     commit,
		RevisionTime: buildTime,
		URL:          repositoryURL,
	})
	parser.SetVersionFields(flags.VersionFieldsCore)
	parser.SetHelpFooter(filterKeyHelpFooter())
	if err := parser.SetDocFooter(flags.DocFormatMarkdown, filterKeyMarkdownFooter()); err != nil {
		return nil, fmt.Errorf("configure Markdown documentation footer: %w", err)
	}

	if err := parser.SetCommandExamples(map[string][]*flags.CommandExample{
		"udp": {
			flags.Example().Describe("Query a legacy master server").
				Option(&commands.Filters.AppID, "10").
				Option(&commands.UDP.Masters, "master.example.net:27011").
				Option(&commands.Query.Limit, "1000"),
			flags.Example().Describe("Combine include and exclude filters").
				Option(&commands.Filters.AppID, "10").
				Option(&commands.UDP.Masters, "master.example.net:27011").
				Option(&commands.Filters.Filter, "map=de_dust2").
				Option(&commands.Filters.NOR, "password=1"),
			flags.Example().Describe("Pass a raw filter expression without CLI validation").
				Option(&commands.Filters.AppID, "10").
				Option(&commands.UDP.Masters, "master.example.net:27011").
				Option(&commands.Filters.RawFilter, `\map\de_dust2`),
		},

		"rest": {
			flags.Example().Describe("Query server addresses with the Web API").
				Option(&commands.Filters.AppID, "730").
				Option(&commands.REST.APIKey, "<KEY>").
				Option(&commands.Query.Limit, "500").
				Option(&commands.REST.Records, "addresses"),
			flags.Example().Describe("Discover beyond the 10,000-row API limit and save its plan").
				Option(&commands.Filters.AppID, "730").
				Option(&commands.REST.APIKey, "<KEY>").
				Option(&commands.Query.Limit, "120000").
				Option(&commands.REST.Plan.PlanOut, "plan.json").
				Option(&commands.REST.Records, "addresses"),
			flags.Example().Describe("Collect using a saved plan").
				Option(&commands.Filters.AppID, "730").
				Option(&commands.REST.APIKey, "<KEY>").
				Option(&commands.REST.Plan.PlanIn, "plan.json").
				Option(&commands.Query.Limit, "120000"),
		},

		"grpc": {
			flags.Example().Describe("Authenticate with mobile QR approval and cache the refresh token").
				Option(&commands.Filters.AppID, "730").
				Option(&commands.GRPC.Auth.Mode, "qr").
				Option(&commands.GRPC.Token.File, "./.steam-token.json").
				Option(&commands.Query.Limit, "10000"),
			flags.Example().Describe("Collect addresses using a saved Steam Connection Manager discovery plan").
				Option(&commands.Filters.AppID, "730").
				Option(&commands.GRPC.Plan.PlanIn, "plan.json").
				Option(&commands.GRPC.Discovery.Records, "addresses").
				Option(&commands.Query.Limit, "120000"),
		},
	}); err != nil {
		return nil, fmt.Errorf("configure command examples: %w", err)
	}

	return parser, nil
}

// init applies the optional build timestamp provided by release tooling.
func init() {
	if _buildTime == "" {
		return
	}

	if value, err := time.Parse(time.RFC3339, _buildTime); err == nil {
		buildTime = value.UTC()
	}
}
