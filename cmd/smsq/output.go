// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/netip"
	"os"
	"slices"

	"github.com/woozymasta/smsq/internal/table"
	"go.yaml.in/yaml/v4"
)

type resultDocument struct {
	// Servers contains the requested result projection.
	Servers any `json:"servers" yaml:"servers"`
	// Status contains protocol-specific discovery progress, when available.
	Status any `json:"status,omitempty" yaml:"status,omitempty"`
	// Complete distinguishes exhausted discovery from a budget-limited result.
	Complete *bool `json:"complete,omitempty" yaml:"complete,omitempty"`
	// Protocol identifies the query transport used for this result.
	Protocol string `json:"protocol" yaml:"protocol"`
	// Filter is the effective wire filter when one applies.
	Filter string `json:"filter,omitempty" yaml:"filter,omitempty"`
	// PlanFile identifies the plan read or written by discovery.
	PlanFile string `json:"plan_file,omitempty" yaml:"plan_file,omitempty"`
	// Count is the number of retained server records.
	Count int `json:"count" yaml:"count"`
	// Requests reports query calls when the protocol exposes that count.
	Requests int `json:"requests,omitempty" yaml:"requests,omitempty"`
	// ElapsedMS is the total operation duration in milliseconds.
	ElapsedMS int64 `json:"elapsed_ms" yaml:"elapsed_ms"`
}

// outputResult writes the selected structured or tabular representation.
func outputResult(
	options outputFlags,
	document resultDocument,
	rows []table.Row,
	availableColumns []string,
	stdout io.Writer,
) (resultErr error) {
	columns, err := table.ResolveColumns(options.Columns, availableColumns)
	if err != nil {
		return fmt.Errorf("resolve output columns: %w", err)
	}

	// Address-only protocols cannot populate the shared metadata columns.
	if slices.Equal(availableColumns, []string{table.ColumnAddress}) {
		columns = []string{table.ColumnAddress}
	}
	if slices.Equal(options.Columns, []string{table.ColumnAll}) {
		columns = table.HideEmptyColumns(rows, columns)
	}

	var writer = stdout
	var file *os.File
	if options.Output != "" {
		created, err := os.OpenFile(
			options.Output,
			os.O_CREATE|os.O_TRUNC|os.O_WRONLY,
			0600,
		)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}

		file = created
		defer func() {
			if err := file.Close(); resultErr == nil && err != nil {
				resultErr = fmt.Errorf("close output file: %w", err)
			}
		}()
		writer = file
	}

	if options.Format == table.FormatHTML || options.Format == table.FormatMarkdown ||
		options.Format == table.FormatCSV || options.Format == table.FormatTable {
		return table.Print(writer, rows, options.Format, columns, document.Filter)
	}

	document.Servers = table.Project(rows, columns)

	var data []byte
	switch options.Format {
	case table.FormatJSON:
		if options.Pretty {
			data, err = json.MarshalIndent(document, "", "  ")
		} else {
			data, err = json.Marshal(document)
		}

	case table.FormatYAML:
		data, err = yaml.Marshal(document)

	default:
		return fmt.Errorf("unsupported output format %q", options.Format)
	}

	if err != nil {
		return fmt.Errorf("encode result as %s: %w", options.Format, err)
	}

	_, err = fmt.Fprintln(writer, string(data))
	return err
}

// addressRows projects endpoints into the shared table's address column.
func addressRows(addresses []netip.AddrPort) []table.Row {
	rows := make([]table.Row, len(addresses))
	for index, address := range addresses {
		rows[index].Address = address.String()
	}

	return rows
}
