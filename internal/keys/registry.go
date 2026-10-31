// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: Copyright 2025-2026 WoozyMasta
// Source: https://github.com/WoozyMasta/smsq

package keys

import (
	"regexp"
)

// Registry maps shared CLI filter names to wire keys and validation metadata.
var Registry = map[string]Info{
	"dedicated": {
		Key:      "dedicated",
		Desc:     "Dedicated servers; 0 selects non-dedicated servers",
		Example:  "1",
		Flag:     true,
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"secure": {
		Key:      "secure",
		Desc:     "VAC-protected servers; 0 selects servers without VAC protection",
		Example:  "1",
		Flag:     true,
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"gamedir": {
		Key:     "gamedir",
		Desc:    "Match the game's directory identifier",
		Example: "cstrike",
		Pattern: regexp.MustCompile(`^[a-z0-9_]+$`),
	},
	"map": {
		Key:     "map",
		Desc:    "Match the server's map name",
		Example: "de_dust2",
		Pattern: regexp.MustCompile(`^[\w\-\*]+$`),
	},
	"linux": {
		Key:      "linux",
		Desc:     "Linux servers; 0 selects non-Linux servers",
		Example:  "1",
		Flag:     true,
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"password": {
		Key:      "password",
		Desc:     "Password-protected servers; 0 selects servers without a password",
		Example:  "1",
		Flag:     true,
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"empty": {
		Key:      "empty",
		Desc:     "Servers with players; 0 selects servers with no players",
		Example:  "1",
		Flag:     true,
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"full": {
		Key:      "full",
		Desc:     "Servers below capacity; 0 selects full servers",
		Example:  "1",
		Flag:     true,
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"notfull": {
		Key:     "notfull",
		Desc:    "Select servers below reported capacity; Steam ignores the operand",
		Example: "1",
		Flag:    true,
	},
	"proxy": {
		Key:      "proxy",
		Desc:     "Proxy servers; 0 selects non-proxy servers",
		Example:  "1",
		Flag:     true,
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"napp": {
		Key:      "napp",
		Desc:     "Exclude this Steam application ID",
		Example:  "10",
		Validate: uintVal,
	},
	"noplayers": {
		Key:      "noplayers",
		Desc:     "Servers with zero players; 0 selects servers with players",
		Example:  "1",
		Flag:     true,
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"white": {
		Key:      "white",
		Desc:     "Servers with whitelist enabled; 0 selects servers with it disabled",
		Example:  "1",
		Flag:     true,
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"gametype": {
		Key:     "gametype",
		Desc:    "Match a server tag (sv_tags)",
		Example: "battleye",
	},
	"gamedata": {
		Key:     "gamedata",
		Desc:    "Match a hidden game-data tag",
		Example: "tag1",
	},
	"name_match": {
		Key:     "name_match",
		Desc:    "Match the server name using wildcards",
		Example: "*server*",
	},
	"version_match": {
		Key:     "version_match",
		Desc:    "Match the game version using wildcards",
		Example: "1.3.*",
	},
	"gameaddr": {
		Key:     "gameaddr",
		Desc:    "Match this exact game address",
		Example: "192.0.2.1:27015",
	},
}
