package table

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/mattn/go-runewidth"
	"github.com/woozymasta/steam/serverlist"
)

var multiSpace = regexp.MustCompile(`\s+`)

// sanitize cleans up server fields by trimming Unicode spaces,
// normalizing multiple spaces, and filling empty values with
// placeholders like "<unnamed>" or "<no map>".
func sanitize(s serverlist.Server) serverlist.Server {
	s.Addr = trimUnicodeSpace(s.Addr)
	s.Name = normalizeSpaces(trimUnicodeSpace(s.Name))
	s.Map = trimUnicodeSpace(s.Map)
	s.Product = trimUnicodeSpace(s.Product)
	s.Version = trimUnicodeSpace(s.Version)

	if s.Addr == "" {
		s.Addr = "localhost"
	}
	if s.Name == "" {
		s.Name = "<unnamed>"
	}
	if s.Map == "" {
		s.Map = "<no map>"
	}
	if s.Product == "" {
		s.Product = "<unknown>"
	}
	return s
}

// stringWidth returns the display width of the string,
// taking into account emoji characters (counted as width 2).
func stringWidth(s string) int {
	width := 0
	for _, r := range s {
		// Emoji обычно 2 колонки
		if isEmoji(r) {
			width += 2
		} else {
			width += runewidth.RuneWidth(r)
		}
	}
	return width
}

// isEmoji checks if a rune belongs to the known emoji ranges.
func isEmoji(r rune) bool {
	// Emoji in ranges
	// 1F300–1F5FF, 1F600–1F64F, 1F680–1F6FF, 1F900–1F9FF, 1FA70–1FAFF, 2600–26FF, 2700–27BF
	return (r >= 0x1F300 && r <= 0x1F5FF) ||
		(r >= 0x1F600 && r <= 0x1F64F) ||
		(r >= 0x1F680 && r <= 0x1F6FF) ||
		(r >= 0x1F900 && r <= 0x1F9FF) ||
		(r >= 0x1FA70 && r <= 0x1FAFF) ||
		(r >= 0x2600 && r <= 0x26FF) ||
		(r >= 0x2700 && r <= 0x27BF)
}

// trimUnicodeSpace removes all leading and trailing Unicode spaces,
// including exotic ones such as U+3164 or ZERO WIDTH SPACE.
func trimUnicodeSpace(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return unicode.IsSpace(r) || r == '\u3164' || r == '\u200B' // ZERO WIDTH SPACE etc
	})
}

// normalizeSpaces replaces consecutive whitespace with a single space.
func normalizeSpaces(s string) string {
	return multiSpace.ReplaceAllString(s, " ")
}

// crop trims a string so that its display width does not exceed n,
// appending an ellipsis ("…") if truncation is necessary.
func crop(in string, n int) string {
	if stringWidth(in) <= n {
		return in
	}
	out := ""
	for _, r := range in {
		if stringWidth(out)+runewidth.RuneWidth(r) >= n-1 {
			break
		}
		out += string(r)
	}
	return out + "…"
}

// calcWidths calculates the required width for each table column
// based on the maximum content width and header length.
func calcWidths(srv serverlist.Servers, maxLen int) map[string]int {
	widths := map[string]int{
		"ADDR":    len("ADDR"),
		"NAME":    len("NAME"),
		"MAP":     len("MAP"),
		"PRODUCT": len("PRODUCT"),
		"VERSION": len("VERSION"),
		"PLAYERS": len("PLAYERS"),
	}
	for _, s := range srv {
		widths["ADDR"] = max(widths["ADDR"], stringWidth(s.Addr))
		widths["NAME"] = max(widths["NAME"], min(stringWidth(s.Name), maxLen))
		widths["MAP"] = max(widths["MAP"], stringWidth(s.Map))
		widths["PRODUCT"] = max(widths["PRODUCT"], stringWidth(s.Product))
		widths["VERSION"] = max(widths["VERSION"], stringWidth(s.Version))
		p := fmt.Sprintf("%d/%d", s.Players, s.MaxPlayers)
		widths["PLAYERS"] = max(widths["PLAYERS"], stringWidth(p))
	}
	for k, v := range widths {
		widths[k] = v + 2
	}
	return widths
}
