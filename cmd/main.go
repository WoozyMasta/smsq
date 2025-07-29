// Package main provides a command-line interface (CLI) for querying the
// Steam Master Server API using the serverlist package. The application
// supports filters, NOR/NAND conditions, JSON and table output formats.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"smsq/internal/keys"
	"smsq/internal/table"
	"smsq/internal/vars"

	"github.com/jessevdk/go-flags"
	"github.com/woozymasta/steam/serverlist"
)

// Options defines all CLI flags and their descriptions.
//
// Short flags:
//   - -x: for NOR filters (exclude).
//   - -X: for NAND filters (stronger exclude).
//
// Example usage:
//
//	smsq -f appid=221100 -f map=namalsk -t -l 10
//
// Environment variables:
//
//	STEAM_API_KEY can be used as a default for --api-key.
type Options struct {
	APIKey       string   `short:"k" long:"api-key" env:"STEAM_API_KEY" description:"Steam Web API key"`
	Filter       []string `short:"f" long:"filter"                      description:"Filter: key=value (repeatable)"`
	Nor          []string `short:"x" long:"nor"                         description:"NOR filter: key=value (repeatable)"`
	Nand         []string `short:"X" long:"nand"                        description:"NAND filter: key=value (repeatable)"`
	Limit        int      `short:"l" long:"limit" default:"100"         description:"Max servers to fetch (Steam API limit 10 000)"`
	Pretty       bool     `short:"p" long:"pretty"                      description:"Pretty‑print JSON"`
	Table        bool     `short:"t" long:"table"                       description:"Print result as table"`
	TableColumns int      `short:"c" long:"column-width" default:"64"   description:"Max width for NAME column when --table is used"`
	Version      bool     `short:"v" long:"version"                     description:"Show build info and exit"`
	Help         bool     `short:"h" long:"help"                        description:"Show help message"`
}

// main is the entry point of the application.
// It parses CLI flags, builds filters, queries the Steam API, and prints the result.
func main() {
	opts := &Options{}

	// Configure parser with custom usage.
	parser := flags.NewParser(opts, flags.PrintErrors)
	parser.Usage = "[OPTIONS]\n  " + filepath.Base(parser.Name) + " -f appid=221100 -f map=namalsk -t -l 10"
	parser.LongDescription = "Steam Master-Server Query CLI."
	parser.Name = filepath.Base(parser.Name)

	// Parse CLI flags.
	if _, err := parser.Parse(); err != nil {
		os.Exit(1) // error already printed by go‑flags
	}

	switch {
	case opts.Help:
		parser.WriteHelp(os.Stdout)
		printKeys()
		return
	case opts.Version:
		vars.PrintVersion()
		return
	}

	if opts.APIKey == "" {
		log.Fatal("Steam API key is required (flag --api-key or env STEAM_API_KEY)")
	}

	// Build Steam filters using keys package.
	filter := &serverlist.Filter{}
	keys.ApplyFilters(opts.Filter, opts.Nor, opts.Nand, filter)

	// Perform request to Steam API.
	q := serverlist.New(opts.APIKey)
	q.SetLimit(opts.Limit)

	servers, err := q.Get(filter)
	if err != nil {
		log.Fatalf("query failed: %v", err)
	}

	// Output table result.
	if opts.Table {
		table.Print(servers, opts.TableColumns)
		return
	}

	// JSON output.
	var data []byte
	if opts.Pretty {
		data, err = json.MarshalIndent(servers, "", "  ")
	} else {
		data, err = json.Marshal(servers)
	}
	if err != nil {
		log.Fatalf("encode JSON: %v", err)
	}

	fmt.Println(string(data))
}

// printKeys outputs all available filter keys with their descriptions.
// This is called when the user runs --help.
func printKeys() {
	fmt.Println("\nFilter keys:")
	for k, info := range keys.Registry {
		line := info.Desc
		fmt.Printf("  %-20s %s\n", k, line)
	}
}
