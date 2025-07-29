// Package table provides functions for printing server lists
// in a clean, aligned table format with support for Unicode,
// emojis, and variable column widths.
//
// The main entry point is the Print function, which takes a slice
// of serverlist.Server objects and prints a table with columns such as:
//
//	#   ADDR         NAME        MAP         PRODUCT   VERSION   PLAYERS
//
// Special Features
//
//   - Unicode-aware column alignment using the mattn/go-runewidth library.
//   - Emoji width detection (emojis often take 2 columns).
//   - Dynamic calculation of column widths based on content and headers.
//   - Automatic trimming of exotic whitespace (e.g., U+3000, U+3164) and
//     replacement of multiple spaces with a single space.
//   - NAME column cropping to a max width (with ellipsis …).
//
// Usage:
//
//	table.Print(servers, 64)
//
// This will print the server list as a table with the NAME column
// cropped to 64 visible characters (width).
package table
