// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

// Package keys parses and validates filter predicates shared by CLI transports.
package keys

import (
	"regexp"
)

// Info describes a single filter key and its validation rules.
type Info struct {
	// Validate is an optional function to check if a value is valid.
	// If nil, no semantic validation is performed.
	Validate Validator

	// Pattern is an optional regular expression that the value must match.
	Pattern *regexp.Regexp

	// Key is the actual FilterKey constant used in serverlist.Filter.
	Key string

	// Desc is a short human-readable description of the filter key.
	Desc string

	// Example is a representative operand shown in filter help.
	Example string

	// Allowed is an optional list of acceptable values for the key.
	// If empty, any value (matching Pattern) is allowed.
	Allowed []string

	// Flag marks an operation-style filter rather than a value matcher.
	Flag bool
}
