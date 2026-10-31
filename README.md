# SMSQ - Steam Master Server Query CLI

<!-- markdownlint-disable-next-line MD033 -->
<img src="logo.svg" alt="SMSQ Logo" align="right" width="128">

Command-line client for finding Steam game servers through -
legacy master server UDP, the Steam Web API, and Steam Connection Manager (CM).
It provides shared filters and output options
while keeping transport-specific settings separate.

<!-- markdownlint-disable-next-line MD033 -->
## Supported protocols and components <br clear="right"/>

### Supported protocols

* Legacy Steam Master Server Query Protocol (MSQP) over UDP.
  Returns server addresses and supports reading subsequent master-server pages.
  The UDP command requires reachable master-server endpoints,
  whose availability is not guaranteed.
* Steam `IGameServersService/GetServerList` Web API.
  Returns full server records or addresses
  and can use Discovery to work around the per-request result limit.
* Steam Connection Manager (CM) protobuf over WebSocket.
  Uses an authenticated Steam session and returns
  full available server metadata by default through `MasterServerQuery`.
  Use `--records addresses` for endpoint-only results.
  Discovery plans are reusable.

### Project components

* [`smsq`][CLI]: CLI for querying all three transports
  with a shared filter syntax and output configuration;
* [MSQP][MSQP]: Go module for the legacy UDP master-server protocol;
* [Steam Server List][Steam Server List]:
  Go module for the Web API and server-list Discovery;
* [Steam CM][Steam CM]:
  Go module for authenticated CM sessions and game-server queries.

## Installation

### CLI

Download the latest release for your platform:

Arch/OS   | macOS           | Linux           | Windows
--------- | --------------- | --------------- | -----------------
**AMD64** | [macOS amd64][] | [Linux amd64][] | [Windows amd64][]
**ARM64** | [macOS arm64][] | [Linux arm64][] | [Windows arm64][]

Or install with Go toolchain:

```shell
go install github.com/woozymasta/smsq/cmd/smsq
```

<!-- markdownlint-disable MD033 -->

<details>
  <summary>Download the latest release with curl in Bash:</summary>

```shell
case "$(uname -s)" in
  Linux*) OS=linux;;
  Darwin*) OS=darwin;;
  MINGW*|MSYS*|CYGWIN*) OS=windows; EXT=.exe;;
  *) exit 1;;
esac

case "$(uname -m)" in
  x86_64|amd64) ARCH=amd64;;
  aarch64|arm64) ARCH=arm64;;
  *) exit 1;;
esac

BIN="./smsq$EXT"
REPO='https://github.com/WoozyMasta/smsq'

curl -#SfLo "$BIN" "$REPO/releases/latest/download/smsq-$OS-$ARCH$EXT"

chmod +x "$BIN"
"$BIN" -h && "$BIN" -v

# Optionally move the binary to some PATH dir:
mv "$BIN" /usr/local/bin/
```

</details>

<!-- markdownlint-enable MD033 -->

> [!NOTE]
> SMSQ is an independent client and is not affiliated with or endorsed by Valve.

## CLI usage

The root command provides shared filtering,
result limits, timeouts, and output options.

> [!TIP]
> Without a subcommand, SMSQ runs `rest`,
> the recommended option for most queries.
> Use `udp` or `grpc` when you specifically need those transports.

Filter support and behavior can vary by protocol and game.
Select the Steam AppID with `--app-id`,
for example `--app-id 730` for Counter-Strike 2.
See the [CLI reference][CLI] for every option,
validation rule, and generated command example.

### Commands

* `udp`: query one or more legacy master servers;
* `rest` (default): query `IGameServersService/GetServerList`
  and optionally discover results beyond one response page;
* `grpc`: query Steam CM over an authenticated connection.
  This command uses protobuf over WebSocket; it is not standard gRPC.

### Examples

```shell
# Query a legacy master server.
smsq udp --master master.example.net:27011 \
  --app-id 10 --limit 1000

# Query server addresses through the default Web API command.
STEAM_API_KEY=some-key
smsq --app-id 730 --records addresses --limit 500

# Discover a large Web API result set and save a reusable plan.
smsq rest --api-key some-key \
  --app-id 730 --limit 120000 \
  --records addresses --plan-out rest-plan.json

# Authenticate to Steam CM and cache the refresh token locally.
smsq grpc --app-id 730 --auth-mode qr \
  --token-file ./.steam-token.json --records addresses
```

## Authentication and Discovery

The Web API command requires a Steam Web API key.
Pass it with `--api-key` or provide it through `STEAM_API_KEY`.

CM queries require a Steam account session.
The CLI supports QR approval, Steam Guard code, and mobile approval.
`--token-file` stores the refresh token as plaintext
so later runs can reuse the session instead of repeating authentication.

> [!CAUTION]
> Treat the token file as a credential. Do not commit or share it.
> Without `--token-file`, the token stays in memory for the current run.

REST and CM Discovery split large result sets into filtered requests.
A single REST response is limited to 10,000 records,
so larger requests use Discovery.
Check `complete` in JSON or YAML output:
`false` means collection stopped before all planned work was exhausted;
`true` describes plan execution,
not a guarantee that Steam's live listing is complete or static.

Save the resulting plan with `--plan-out`
and reuse it on later runs with `--plan-in`;
this avoids rebuilding the same query plan for recurring collections.
Discovery can issue multiple Web API requests,
so avoid polling it too frequently.

> [!NOTE]
> Raw filters bypass structured validation
> and cannot be combined with structured filters or Discovery.

Supply the AppID with `--app-id` or as the leading `\appid\<id>` pair
in the raw filter; an AppID in the raw filter takes precedence over `--app-id`.

## Protocol Documentation

For details on the protocols and APIs used, see:

* [Steam Master Server Query Protocol][Master Server Query Protocol]
* [Steamworks `IGameServersService`][IGameServersService]
* [Steam CM module documentation][Steam CM]

## Tested Games

Live server-list checks have covered Counter-Strike titles and other games.
CM Discovery has also been exercised with large result sets.
Availability and results depend on the game and Steam's current server listing.

## 👉 [Support Me][]

Your support is greatly appreciated!

<!-- Links -->
[Master Server Query Protocol]: https://developer.valvesoftware.com/wiki/Master_Server_Query_Protocol "Valve Master Server Query Protocol"
[IGameServersService]: https://partner.steamgames.com/doc/webapi/IGameServersService "Steamworks IGameServersService"

[CLI]: ./CLI.md "Generated command-line reference"
[MSQP]: https://github.com/WoozyMasta/msqp "Legacy UDP master-server module"
[Steam Server List]: https://github.com/WoozyMasta/steamserverlist "Steam Web API server-list module"
[Steam CM]: https://github.com/WoozyMasta/steamcm "Steam Connection Manager module"

[macOS arm64]: https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-darwin-arm64 "macOS arm64 binary"
[macOS amd64]: https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-darwin-amd64 "macOS amd64 binary"
[Linux amd64]: https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-linux-amd64 "Linux amd64 binary"
[Linux arm64]: https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-linux-arm64 "Linux arm64 binary"
[Windows amd64]: https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-windows-amd64.exe "Windows amd64 binary"
[Windows arm64]: https://github.com/WoozyMasta/smsq/releases/latest/download/smsq-windows-arm64.exe "Windows arm64 binary"

[Support Me]: https://gist.github.com/WoozyMasta/7b0cabb538236b7307002c1fbc2d94ea
