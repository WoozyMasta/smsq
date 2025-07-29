package keys

import (
	"regexp"

	"github.com/woozymasta/steam/serverlist"
)

// Info describes a single filter key and its validation rules.
type Info struct {
	// Key is the actual FilterKey constant used in serverlist.Filter.
	Key serverlist.FilterKey

	// Desc is a short human-readable description of the filter key.
	Desc string

	// Validate is an optional function to check if a value is valid.
	// If nil, no semantic validation is performed.
	Validate Validator

	// Allowed is an optional list of acceptable values for the key.
	// If empty, any value (matching Pattern) is allowed.
	Allowed []string

	// Pattern is an optional regular expression that the value must match.
	Pattern *regexp.Regexp
}
