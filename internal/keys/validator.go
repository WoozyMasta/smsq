package keys

import (
	"fmt"
	"strconv"
)

// Validator checks value correctness for a filter key.
// It must return nil on success.
type Validator func(string) error

// boolVal validates 0|1.
func boolVal(v string) error {
	if v == "0" || v == "1" {
		return nil
	}
	return fmt.Errorf("must be 0 or 1")
}

// uintVal validates an unsigned integer value.
func uintVal(v string) error {
	if _, err := strconv.ParseUint(v, 10, 64); err != nil {
		return fmt.Errorf("must be unsigned integer: %w", err)
	}
	return nil
}
