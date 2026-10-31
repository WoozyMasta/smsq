// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package main

import (
	"context"
	"io"
	"time"

	"github.com/woozymasta/msqp"
	"github.com/woozymasta/smsq/internal/table"
	masterudp "github.com/woozymasta/smsq/internal/udp"
)

func runUDP(
	parent context.Context,
	filters filterFlags,
	query queryFlags,
	output outputFlags,
	command udpCommand,
	stdout io.Writer,
) error {
	started := time.Now()
	var err error

	_, filter, _, err := buildFilter(filters)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(parent, query.Timeout)
	defer cancel()

	addresses, requests, complete, err := masterudp.Query(
		ctx,
		command.Masters,
		filter,
		msqp.Region(command.Region),
		uint(query.Limit),
		command.MaxPages,
		command.IOTimeout,
	)
	if err != nil {
		return err
	}

	rows := addressRows(addresses)
	document := resultDocument{
		Protocol:  "udp",
		Filter:    filter,
		Servers:   addresses,
		Count:     len(addresses),
		Requests:  requests,
		Complete:  new(complete),
		ElapsedMS: time.Since(started).Milliseconds(),
	}
	if output.Format == "table" {
		document.Servers = nil
	}

	return outputResult(output, document, rows, []string{table.ColumnAddress}, stdout)
}
