package table

import (
	"fmt"

	"github.com/woozymasta/steam/serverlist"
)

// Print prints the server list as an aligned table.
// It automatically:
//   - Sanitizes server fields (trims spaces, fills empty fields with defaults).
//   - Computes column widths dynamically.
//   - Crops the NAME column if it exceeds maxLen (in display width).
//
// The table includes the following columns:
//
//	#   ADDR   NAME   MAP   PRODUCT   VERSION   PLAYERS
func Print(srv serverlist.Servers, maxLen int) {
	// cleanup
	for i := range srv {
		srv[i] = sanitize(srv[i])
	}

	colWidths := calcWidths(srv, maxLen)

	// width for number coll
	numWidth := len(fmt.Sprintf("%d", len(srv)))
	if numWidth < 3 {
		numWidth = 3
	}

	// header
	fmt.Printf("%-*s %-*s %-*s %-*s %-*s %-*s %-*s\n",
		numWidth, "#",
		colWidths["ADDR"], "ADDR",
		colWidths["NAME"], "NAME",
		colWidths["MAP"], "MAP",
		colWidths["PRODUCT"], "PRODUCT",
		colWidths["VERSION"], "VERSION",
		colWidths["PLAYERS"], "PLAYERS",
	)

	// rows
	for i, s := range srv {
		fmt.Printf("%-*d %-*s %-*s %-*s %-*s %-*s %-*s\n",
			numWidth, i+1,
			colWidths["ADDR"], s.Addr,
			colWidths["NAME"], crop(s.Name, maxLen),
			colWidths["MAP"], s.Map,
			colWidths["PRODUCT"], s.Product,
			colWidths["VERSION"], s.Version,
			colWidths["PLAYERS"], fmt.Sprintf("%d/%d", s.Players, s.MaxPlayers),
		)
	}
}
