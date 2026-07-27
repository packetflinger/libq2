# libq2

A Go module (`github.com/packetflinger/libq2`) implementing the Quake II network protocol, on-disk file formats, and related utilities — a toolkit for building bots, demo tools, master servers, and server-query utilities for id Software's Quake II.

## Packages

| Package | Purpose |
|---|---|
| `message` | Core of the library — encodes/decodes Q2 network packets (SVC_*/CLC_* messages, entity/player-state deltas, connectionless packets) |
| `bot` | Full UDP client/netchan implementation to connect a bot to a Q2 server (challenge/connect handshake, reliable ack, callbacks per message type) |
| `client` | Client-side math: view angles, movement, velocity |
| `player` | Userinfo string marshaling, death/obituary parsing |
| `demo` | Reads/writes `.dm2` demos and MVD (multi-view/GTV) demos |
| `bsp` | Parses `.bsp` map files (entities, planes, textures, vertices, PVS/visibility) |
| `pak` | Reads/writes `.pak` and `.pkz` (zip) asset archives |
| `master` | A Quake II master server (heartbeat protocol) + JSON HTTP API (`/GetServers`, `/HealthCheck`, `/ServerInfo`) |
| `state` | Server/client state, e.g. `Server.FetchInfo()` to query a live server, `UserCmd` |
| `flags` | Deathmatch flags bitfield parsing |
| `proto` | Protobuf schemas/generated code (challenge, packet, pak, server messages, etc.) used internally for structured data |
| `types`, `util` | Shared helpers (vectors, misc utilities) |
| `testdata` | Sample `.bsp`/`.pak`/`.dm2`/`.mvd2` fixtures for tests |

## Examples

The `examples/` directory contains runnable programs showing usage:

- `bots/simplebot.go` — connects a bot to a server
- `demos/decode`, `demos/mvd`, `demos/parsedemo` — demo file tools
- `master/master.go` — run a master server
- `pak/listpakfiles.go` — list PAK archive contents
- `serverinfo/info.go` — query a server's info string

## Usage

```bash
go get github.com/packetflinger/libq2
```

Then import the package(s) you need (e.g. `pak`, `state`, `bot`) in your Go code — the `examples/` directory doubles as usage documentation, e.g.:

```go
srv := state.Server{Address: "host", Port: 27910}
info, _ := srv.FetchInfo()
```

Run tests with:

```bash
go test ./...
```