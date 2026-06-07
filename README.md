<p align="center">
  <img src="logo.png" alt="shupo logo" width="240">
</p>

<h1 align="center">shupo</h1>

<p align="center">
  A video processing pipeline. Go at the front door, Rust at
  the heavy lifting, Redis Streams as the seam.
</p>

---

## What is this?

A video processing pipeline rebuilt from scratch.The rebuild splits the workload
across two runtimes:

- **Go** owns the front door: HTTP, rate limiting, S3 upload, Redis producer.
- **Rust** owns the heavy lifting: job consumption, FFmpeg, dead- letter
  handling, WebSocket notifier.
- **Redis Streams** is the seam. A second stream carries cancellations. The seam
  is what lets the two runtimes evolve independently.

## Quick start

```bash
# Enter the dev shell. nix provides Go, Rust, ffmpeg, redis-cli, etc.
nix develop

# Bring up the shared infrastructure (Redis, MinIO).
just dev-up

# Build every service with nix.
just build-nix

# Or build the Go service by hand, for faster iteration.
(cd services/ingest && go run ./...)
```

The full task list:

```bash
just --list
```

## Layout

```
shupo/
├── flake.nix              dev shell, build packages, NixOS modules
├── README.md             
├── .gitignore
│
├── services/
│   ├── ingest/            Go front door
│   ├── worker/            Rust worker (FFmpeg, DLQ, cancel)
│   └── notifier/          Rust WebSocket fan-out
│
├── crates/                shared Rust workspace members
│   └── contract/          the global types
│
├── shared/                cross-service artefacts
│   └── fixtures/          JSON contracts for the round-trip test
│
├── deploy/                production-shape operational
│   ├── Caddyfile
│   └── docker-compose.yml
│
├── nix/                   per-service build derivations + Hetzner module
```

## Architecture, in one diagram (text)

```
                 ┌────────────┐
   client ───▶   │  ingest    │  ──▶  S3 / MinIO  (input)
   (HTTP)        │  (Go)      │  ──▶  Redis: stream:jobs
                 └────────────┘                  │
                                                ▼
                                        ┌────────────┐
                                        │  worker    │  ──▶  FFmpeg
                                        │  (Rust)    │  ──▶  S3 / MinIO (output)
                                        └────────────┘  ──▶  Redis: status hash
                                                │
                  ┌────────────┐                 │
   client ◀───    │  notifier  │  ◀────  Redis pub/sub
   (WebSocket)    │  (Rust)    │
                  └────────────┘

   cancel path:  ingest → Redis: stream:cancel → worker (token)
```

## Going further

- **Stretch goals:** HLS transcoding, JWT auth, distributed tracing, swap Redis
  for NATS.

## License

MIT. See the workspace `Cargo.toml`.
