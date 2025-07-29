package keys

import (
	"fmt"
	"log"
	"strings"

	"github.com/woozymasta/steam/serverlist"
)

// ApplyFilters parses and validates multiple sets of filter expressions
// (Filter, NOR, NAND) and applies them to the given serverlist.Filter instance.
//
// Each list (filters, nors, nands) should contain items in the form "key=value".
// The function will:
//  1. Split each entry into key and value.
//  2. Check that the key is defined in the Registry.
//  3. Validate the value using Info.Validate (if defined).
//  4. Ensure the value matches Info.Pattern (if defined).
//  5. Append the "key\value" pair to the Filter using the corresponding Add,
//     AddNor, or AddNand method.
//
// Fatal errors will be triggered if:
//   - A key is unknown.
//   - A value fails validation or regex pattern check.
//   - A "key=value" string is malformed.
func ApplyFilters(filters, nors, nands []string, f *serverlist.Filter) {
	addKV(filters, f.Add)
	addKV(nors, f.AddNor)
	addKV(nands, f.AddNand)
}

// addKV processes a list of "key=value" strings and appends them to
// a Filter by calling the provided addFn.
//
// It is used internally by ApplyFilters to handle each group of filters
// (standard, NOR, and NAND).
func addKV(list []string, addFn func(serverlist.FilterKey, string)) {
	for _, kv := range list {
		k, v, err := splitKV(kv)
		if err != nil {
			log.Fatalf("invalid key=value pair %q: %v", kv, err)
		}

		info, ok := Registry[k]
		if !ok {
			log.Fatalf("unknown filter key %q. Use --help to list valid keys.", k)
		}

		if info.Validate != nil {
			if err := info.Validate(v); err != nil {
				log.Fatalf("invalid value %q for key %q: %v", v, k, err)
			}
		}
		if info.Pattern != nil && !info.Pattern.MatchString(v) {
			log.Fatalf("value %q for key %q doesn't match pattern %s", v, k, info.Pattern)
		}

		addFn(info.Key, v)
	}
}

// splitKV splits a "key=value" string into its components.
// It returns an error if the string is malformed or has no key.
//
// Example:
//
//	k, v, err := splitKV("appid=221100")
//	k == "appid", v == "221100"
func splitKV(s string) (string, string, error) {
	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 || parts[0] == "" {
		return "", "", fmt.Errorf("expected key=value")
	}

	return parts[0], parts[1], nil
}
