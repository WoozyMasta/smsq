# Steam Master Server Query CLI

Steam Master Server Query CLI (smsq) — a command-line tool to retrieve and
filter game server lists from the [Steam Master Server][].  
It supports advanced filtering, JSON and table output formats, and is
designed for server administrators, gamers, and automation scripts.  
A valid [Steam API Key][] is required to use this tool.

## Features

* Query Steam Master Server for game servers.
* Filter results by AppID, map, tags, or server properties (e.g. VAC, password).
* Output in JSON (for easy jq parsing) or formatted table view.
* Supports NOR and NAND filters for exclusion logic.
* Cross-platform: Linux, Windows, MacOS binaries.
* Lightweight and with minimal dependencies.

## Why smsq?

The Steam Master Server Query API requires constructing complex filter strings
(e.g., `appid\221100\map\namalsk`) and remembering all supported keys.
smsq simplifies this by providing all filter keys with built-in descriptions
and validation, so you can build powerful queries using simple `key=value`
pairs without memorizing obscure syntax.

It replaces manual curl calls and raw JSON parsing with a clean,
user-friendly CLI.

## Installation

You can download the latest version of the programme by following the links:

* [MacOS arm64][]
* [MacOS amd64][]
* [Linux i386][]
* [Linux amd64][]
* [Linux arm][]
* [Linux arm64][]
* [Windows i386][]
* [Windows amd64][]
* [Windows arm64][]

For Linux you can also use the command

```bash
curl -#SfLo /usr/bin/smsq \
  https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-linux-amd64
chmod +x /usr/bin/smsq
smsq -h && smsq -v
```

## Parameters

This help information is available when running the program
with the `--help` flag

```bash
smsq --help
```

```txt
Usage:
  smsq [OPTIONS]
  smsq -f appid=221100 -f map=namalsk -t -l 10

Steam Master-Server Query CLI.

Application Options:
  -k, --api-key=      Steam Web API key [$STEAM_API_KEY]
  -f, --filter=       Filter: key=value (repeatable)
  -x, --nor=          NOR filter: key=value (repeatable)
  -X, --nand=         NAND filter: key=value (repeatable)
  -l, --limit=        Max servers to fetch (Steam API limit 10 000) (default: 100)
  -p, --pretty        Pretty‑print JSON
  -t, --table         Print result as table
  -c, --column-width= Max width for NAME column when --table is used (default: 64)
  -v, --version       Show build info and exit
  -h, --help          Show help message

Filter keys:
  white                1 — Whitelist on, 0 — off
  gametype             sv_tags entry, e.g. "battleye"
  dedicated            1 — Dedicated servers, 0 — Non-dedicated
  password             1 — Password protected, 0 — Open
  gameaddr             Exact IP[:port], e.g. "192.168.1.1:27015"
  empty                1 — Not empty, 0 — Empty
  proxy                1 — Proxy servers, 0 — Normal servers
  collapse_addr_hash   1 — One server per IP, 0 — many
  gamedata             Hidden gamedata tag, e.g. "tag1"
  version_match        Wildcard version, e.g. "1.3.*"
  secure               1 — VAC/BattleEye, 0 — Insecure
  gamedir              Game dir, e.g. "tf", "cstrike", "dayz"
  map                  Map name, e.g. "ctf_2fort"
  linux                1 — Linux host, 0 — Other OS
  appid                Exact AppID, e.g. 221100
  name_match           Wildcard name, e.g. "*server*"
  full                 1 — Not full, 0 — Full
  napp                 Exclude AppID
  noplayers            1 — Zero players, 0 — One or more players
```

## Example usage

```bash
smsq -k <API_KEY> -f appid=221100 -l 5 -t
```

```txt
#   ADDR                   NAME                                      MAP             PRODUCT   VERSION       PLAYERS
1   212.124.22.212:27010   Test Development Server 1 chernarusplus   chernarusplus   dayz      1.28.160123   0/5
2   212.124.22.212:27012   Test Development Server 3 sakhal          sakhal          dayz      1.28.160123   0/5
3   94.181.44.83:27016     DAYZ 323                                  chernarusplus   dayz      1.28.160123   0/60
4   95.79.97.108:27016     lyskovo                                   chernarusplus   dayz      1.28.160123   0/64
5   193.238.44.61:2803     [FXT] FOXTROT DeathMatch [1PP] S          enoch           dayz      1.28.160123   0/10
```

More examples

```bash
# Query DayZ servers with a table view (limit 5), pass token in option
smsq -k <API_KEY> -f appid=221100 -t
# Query DayZ servers with a table view (limit 5), pass token in variable
STEAM_API_KEY=key smsq -f appid=221100 -f map=namalsk -t -l 5
# Export token
export STEAM_API_KEY=key
# Filter with jq to show only servers with players
smsq -f appid=221100 | jq '.[] | select(.players > 0) | .name'
# Show address, map, and player count
smsq -f appid=221100 | jq -r '.[] | "\(.addr) | \(.map) | \(.players)/\(.max_players)"'
```

## Support me 💖

If you enjoy my projects and want to support further development,
feel free to donate! Every contribution helps to keep the work going.
Thank you!

### Crypto Donations

<!-- cSpell:disable -->
* **BTC**: `1Jb6vZAMVLQ9wwkyZfx2XgL5cjPfJ8UU3c`
* **USDT (TRC20)**: `TN99xawQTZKraRyvPAwMT4UfoS57hdH8Kz`
* **TON**: `UQBB5D7cL5EW3rHM_44rur9RDMz_fvg222R4dFiCAzBO_ptH`
<!-- cSpell:enable -->

<!-- Links -->

[Steam Master Server]: https://developer.valvesoftware.com/wiki/Master_Server_Query_Protocol
[Steam API Key]: https://steamcommunity.com/dev/apikey
[MacOS arm64]: https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-darwin-arm64 "MacOS arm64 file"
[MacOS amd64]: https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-darwin-amd64 "MacOS amd64 file"
[Linux i386]: https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-linux-386 "Linux i386 file"
[Linux amd64]: https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-linux-amd64 "Linux amd64 file"
[Linux arm]: https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-linux-arm "Linux arm file"
[Linux arm64]: https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-linux-arm64 "Linux arm64 file"
[Windows i386]: https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-windows-386.exe "Windows i386 file"
[Windows amd64]: https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-windows-amd64.exe "Windows amd64 file"
[Windows arm64]: https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-windows-arm64.exe "Windows arm64 file"
