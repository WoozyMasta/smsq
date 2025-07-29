// Package keys provides metadata and validation logic for all supported
// serverlist.FilterKey constants used by the CLI.
//
// # Overview
//
// The Steam Master Server API supports a set of query filters (e.g.,
// "appid", "map", "dedicated"), each represented in this package by
// an entry in the Registry map. Each key has a description, optional
// validators, allowed value sets, and pattern constraints.
//
// Key Concepts
//
//   - Registry – a central map of CLI filter keys (string) to Info
//     structures, which link the CLI input to the serverlist.FilterKey
//     constants.
//
//   - Info – a metadata container describing one key, including
//     description text, allowed values, and validation rules.
//
//   - Validator – a function type that validates a given string value
//     for a key. Common validators like boolVal (for "0"/"1") and
//     uintVal (for unsigned integers) are provided.
//
// # Typical Use
//
// The CLI code calls ApplyFilters() with lists of key=value pairs.
// The keys package then validates these pairs against Registry, ensuring
// that the keys exist, and that the values match the defined constraints.
//
// Example
//
//	f := &serverlist.Filter{}
//	keys.ApplyFilters([]string{"appid=221100"}, nil, nil, f)
//
// This will validate the "appid" key, check that "221100" is an unsigned
// integer, and add it to the Filter.
//
// References
//
//   - Steam Master Server Query Protocol:
//     https://developer.valvesoftware.com/wiki/Master_Server_Query_Protocol
//
//   - serverlist.FilterKey constants are defined in the
//     github.com/woozymasta/steam/serverlist package.
package keys
