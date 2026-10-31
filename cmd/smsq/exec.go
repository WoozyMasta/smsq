// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package main

// Execute runs the selected UDP command after flags finishes parsing.
func (command *udpCommand) Execute(_ []string) error {
	return runUDP(
		command.runtime.ctx,
		*command.runtime.filters,
		*command.runtime.query,
		*command.runtime.output,
		*command,
		command.runtime.stdout,
	)
}

// Execute runs the selected Web API command after flags finishes parsing.
func (command *restCommand) Execute(_ []string) error {
	return runREST(
		command.runtime.ctx,
		*command.runtime.filters,
		*command.runtime.query,
		*command.runtime.output,
		*command,
		command.runtime.stdout,
	)
}

// Execute runs the selected Steam CM command after flags finishes parsing.
func (command *grpcCommand) Execute(_ []string) error {
	return runGRPC(
		command.runtime.ctx,
		*command.runtime.filters,
		*command.runtime.query,
		*command.runtime.output,
		*command,
		command.runtime.stdin,
		command.runtime.stdout,
		command.runtime.stderr,
	)
}
