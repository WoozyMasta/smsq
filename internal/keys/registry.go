package keys

import (
	"regexp"

	"github.com/woozymasta/steam/serverlist"
)

// Registry maps CLI string keys to FilterKey constants with validation meta.
var Registry = map[string]Info{
	"dedicated": {
		Key:      serverlist.KeyDedicated,
		Desc:     "1 — Dedicated servers, 0 — Non-dedicated",
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"secure": {
		Key:      serverlist.KeySecure,
		Desc:     "1 — VAC/BattleEye, 0 — Insecure",
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"gamedir": {
		Key:     serverlist.KeyGameDir,
		Desc:    `Game dir, e.g. "tf", "cstrike", "dayz"`,
		Pattern: regexp.MustCompile(`^[a-z0-9_]+$`),
	},
	"map": {
		Key:     serverlist.KeyMap,
		Desc:    `Map name, e.g. "ctf_2fort"`,
		Pattern: regexp.MustCompile(`^[\w\-\*]+$`),
	},
	"linux": {
		Key:      serverlist.KeyLinux,
		Desc:     "1 — Linux host, 0 — Other OS",
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"password": {
		Key:      serverlist.KeyPassword,
		Desc:     "1 — Password protected, 0 — Open",
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"empty": {
		Key:      serverlist.KeyEmpty,
		Desc:     "1 — Not empty, 0 — Empty",
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"full": {
		Key:      serverlist.KeyFull,
		Desc:     "1 — Not full, 0 — Full",
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"proxy": {
		Key:      serverlist.KeyProxy,
		Desc:     "1 — Proxy servers, 0 — Normal servers",
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"appid": {
		Key:      serverlist.KeyAppID,
		Desc:     `Exact AppID, e.g. 221100`,
		Validate: uintVal,
	},
	"napp": {
		Key:      serverlist.KeyNotAppID,
		Desc:     `Exclude AppID`,
		Validate: uintVal,
	},
	"noplayers": {
		Key:      serverlist.KeyNoPlayers,
		Desc:     "1 — Zero players, 0 — One or more players",
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"white": {
		Key:      serverlist.KeyWhite,
		Desc:     "1 — Whitelist on, 0 — off",
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"gametype": {
		Key:  serverlist.KeyGameType,
		Desc: `sv_tags entry, e.g. "battleye"`,
	},
	"gamedata": {
		Key:  serverlist.KeyGameData,
		Desc: `Hidden gamedata tag, e.g. "tag1"`,
	},
	"name_match": {
		Key:  serverlist.KeyName,
		Desc: `Wildcard name, e.g. "*server*"`,
	},
	"version_match": {
		Key:  serverlist.KeyVersion,
		Desc: `Wildcard version, e.g. "1.3.*"`,
	},
	"collapse_addr_hash": {
		Key:      serverlist.KeySingleAddr,
		Desc:     "1 — One server per IP, 0 — many",
		Validate: boolVal,
		Allowed:  []string{"0", "1"},
	},
	"gameaddr": {
		Key:  serverlist.KeyGameAddr,
		Desc: `Exact IP[:port], e.g. "192.168.1.1:27015"`,
	},
}
