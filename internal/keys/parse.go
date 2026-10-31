// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package keys

import (
	"fmt"
	"strings"
)

// Predicate is a validated filter key/value pair independent of a protocol client.
type Predicate struct {
	// Key is the Steam master-server filter key.
	Key string
	// Value is the filter operand supplied by the caller.
	Value string
}

// Parse validates CLI key=value expressions using the shared filter registry.
func Parse(values []string) ([]Predicate, error) {
	result := make([]Predicate, 0, len(values))
	for _, expression := range values {
		key, value, ok := strings.Cut(expression, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid filter %q: expected key=value", expression)
		}

		info, exists := Registry[key]
		if !exists {
			return nil, fmt.Errorf("unknown filter key %q", key)
		}
		if isFlagValue(info) {
			value = normalizeFlagValue(value)
		}

		if info.Validate != nil {
			if err := info.Validate(value); err != nil {
				return nil, fmt.Errorf("invalid value for filter %q: %w", key, err)
			}
		}

		if info.Pattern != nil && !info.Pattern.MatchString(value) {
			return nil, fmt.Errorf("value for filter %q does not match %s", key, info.Pattern)
		}

		result = append(result, Predicate{Key: info.Key, Value: value})
	}

	return result, nil
}

// isFlagValue reports whether the registry marks a key as a 0|1 filter.
func isFlagValue(info Info) bool {
	return len(info.Allowed) == 2 && info.Allowed[0] == "0" && info.Allowed[1] == "1"
}
