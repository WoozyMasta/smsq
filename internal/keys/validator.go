// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package keys

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Validator checks value correctness for a filter key.
// It must return nil on success.
type Validator func(string) error

// boolVal validates 0|1.
func boolVal(v string) error {
	if v == "0" || v == "1" {
		return nil
	}

	return errors.New("must be 0 or 1, true/false, yes/no, or y/n")
}

// normalizeFlagValue maps accepted boolean spellings to Steam's 0|1 operands.
func normalizeFlagValue(value string) string {
	switch strings.ToLower(value) {
	case "1", "true", "yes", "y":
		return "1"

	case "0", "false", "no", "n":
		return "0"

	default:
		return value
	}
}

// uintVal validates an unsigned integer value.
func uintVal(v string) error {
	if _, err := strconv.ParseUint(v, 10, 64); err != nil {
		return fmt.Errorf("must be unsigned integer: %w", err)
	}

	return nil
}
