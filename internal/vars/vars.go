// Package vars is an internal technical variable store used at build time,
// populated with values ​​based on the state of the git repository.
package vars

import (
	"fmt"
	"os"
)

var (
	// Version of application (git tag)
	Version string

	// Commit sha from git
	Commit string

	// BuildTime time start build app
	BuildTime string

	// URL to repository
	URL string
)

// PrintVersion prints build‑time metadata embedded via -ldflags.
func PrintVersion() {
	fmt.Printf(`file:     %s
version:  %s
commit:   %s
built:    %s
project:  %s
`,
		os.Args[0], Version, Commit, BuildTime, URL,
	)
}
