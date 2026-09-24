# Xylon demo customer admin app

Dummy **SaaS customer** backend. Holds one customer's current shared secret and:

- Mints short-lived Xylon Control HMAC tokens for browsers (`getToken`)
- Forwards Control admin (presence, history, inject, moderate, record, live
  session list, fleet) to
  [xylon-minimal-customer-interface](https://github.com/Lazarus404/xylon-minimal-customer-interface)
- Receives forwarded Xylon lifecycle / recording webhooks into SQLite history
  plus a lab ring buffer
- Polls interface fleet nodes into SQLite `node_snapshots`
- Exposes machine-readable `/v1/ops/*` from that cache plus advisory
  `scale.recommend` (orchestrators decide; admin never executes scale)
- Serves an embedded **strike-fw** / **strike-fw-ui** console (`GET /`) with
  Live | History | Fleet tabs

Never returns the shared secret. Never calls Xylon HTTP directly -- admin goes
through the platform interface so customer status and feature flags can deny.

Related: [GOLANG_APPS.md](../GOLANG_APPS.md), [ADMIN.md](../ADMIN.md).

## Run

Start the platform interface on `:8093` (and its secret store) before exercising
admin routes. Then:

```bash
cd xylon-demo-customer-admin-app
cd web && npm ci && npm run build && cd ..
go run ./cmd/admin -config config.example.json
```

Open `http://127.0.0.1:8092/` for the console. Vite lab: `cd web && npm run dev`
(proxies `/v1` to `:8092`; port `5174`).

Default listen: `127.0.0.1:8092` (loopback required; API is open on bind).

Lab secret in `config.example.json` is `lab-hmac-secret` (matches
`mix xylon.datachannel --control hmac` when Xylon uses the Config adapter).
With `XYLON_CONTROL_SECRETS_URL` / `Secrets.Local`, set `hmac_secret` to the
store's current secret for this `customer_id` (interface provision/rotate).
The test client Mint mode calls this service via SDK `getToken`.

Set interface `admin_webhook_url` to `http://127.0.0.1:8092/v1/webhooks/xylon`
and use the same `webhook_secret` as Xylon / compose / interface
(`lab-webhook-secret` in examples). Enable customer feature `fleet` (default
for seeded `acme`).

### Secret handoff

1. Point the customer interface `admin_config_path` at a JSON file used by this
   app (or paste `hmac_secret` from provision/rotate into `config.example.json`).
2. **Restart** this process after the interface rewrites `hmac_secret` (no
   hot-reload in v1).

Interface-written files may contain only `customer_id` and `hmac_secret`;
`listen_addr`, `interface_base_url`, TTLs, `cors_origins`, `sqlite_path`,
`fleet_poll_sec`, and `scale_out_participants` fall back to defaults on load.

## Config

| Key | Meaning |
| --- | --- |
| `customer_id` | Tenant id in minted tokens and interface admin paths |
| `hmac_secret` | Current signing secret for browser mint |
| `listen_addr` | Bind (must be loopback) |
| `interface_base_url` | Platform interface base (default `http://127.0.0.1:8093`) |
| `default_ttl_sec` / `max_ttl_sec` | Mint TTL |
| `cors_origins` | Browser origins (mint + Vite `:5174`) |
| `webhook_secret` | Shared Xylon Control webhook secret (ingest verify) |
| `sqlite_path` | History + fleet DB path (default `./data/admin-history.sqlite`) |
| `fleet_poll_sec` | Fleet poll interval (default `5`) |
| `scale_out_participants` | Advisory `scale_out` when any reachable node is `>=` this (default `100`) |

## HTTP API

| Method | Path | Notes |
| --- | --- | --- |
| GET | `/` / `/index.html` / `/assets/*` | Embedded strike-fw UI |
| GET | `/healthz` | `{"ok":true}` |
| POST | `/v1/token` | local mint |
| OPTIONS | `/v1/token` | CORS preflight |
| GET | `/v1/sessions` | forward live session list |
| GET | `/v1/sessions/:sid/presence` | forward to interface |
| GET | `/v1/sessions/:sid/history` | forward; query `label` |
| POST | `/v1/sessions/:sid/data` | forward inject body |
| POST | `/v1/sessions/:sid/moderate` | forward moderate body |
| POST | `/v1/sessions/:sid/record/start` | forward record start |
| POST | `/v1/sessions/:sid/record/stop` | forward record stop |
| GET | `/v1/sessions/:sid/record` | forward record status |
| GET | `/v1/history/sessions` | SQLite past sessions (`limit`) |
| GET | `/v1/history/sessions/:sid` | session + participants + recordings |
| GET | `/v1/fleet/summary` | derived from snapshots (`?refresh=1`) |
| GET | `/v1/fleet/nodes` | cached nodes (`?refresh=1` live pull) |
| POST | `/v1/fleet/accepting` | `{"node","accepting"}`; drain/undrain; refresh cache |
| GET | `/v1/ops/summary` | same body as fleet summary (`?refresh=1`) |
| GET | `/v1/ops/nodes` | same body as fleet nodes (`?refresh=1`) |
| POST | `/v1/ops/accepting` | same as fleet accepting |
| GET | `/v1/ops/recommendations/latest` | newest `scale.recommend`, or `404` empty |
| POST | `/v1/webhooks/xylon` | signed ingest (SQLite + lab ring) |
| GET | `/v1/webhooks/xylon/latest` | newest ring event |

Session admin, history, fleet, and ops paths use config `customer_id`. Interface
responses (status + body) are passed through on live proxies. Dial failures
return `502` `{"error":"interface_unreachable"}`.

Ops shares the fleet SQLite cache so autoscalers see the same counts as the
Fleet tab. Node JSON is counts/ids only (no `payload_json`). After each
successful refresh, if a reachable node has
`participant_count >= scale_out_participants`, an advisory
`{"type":"scale.recommend","action":"scale_out",...}` is stored in a separate
in-memory ring (identical consecutive advisories are not re-added).
Orchestrators pull `/v1/ops/*` and decide; this app never executes scale or
POSTs to an orchestrator.

### SQLite schema (history + fleet)

```text
events_raw            payload_hash UNIQUE, event_type, tenant_id, session_id, ...
sessions              PK (tenant_id, session_id); created_at, closed_at, node, region
participant_events    join/leave rows
recording_segments    PK (tenant_id, recording_id); status completed|failed
node_snapshots        PK node_id; region, session/participant counts, hub, polled_at
```

Ingest is idempotent on `payload_hash` (sha256 of raw body).

```bash
curl -s 'http://127.0.0.1:8092/v1/fleet/nodes?refresh=1'
curl -s 'http://127.0.0.1:8092/v1/fleet/summary'
curl -s 'http://127.0.0.1:8092/v1/ops/summary?refresh=1'
curl -s 'http://127.0.0.1:8092/v1/ops/nodes'
curl -s 'http://127.0.0.1:8092/v1/ops/recommendations/latest'
curl -s -X POST 'http://127.0.0.1:8092/v1/ops/accepting' \
  -H 'content-type: application/json' \
  -d '{"node":"a@host","accepting":false}'

curl -s -X POST http://127.0.0.1:8092/v1/token \
  -H 'content-type: application/json' \
  -d '{"session_id":"room-1","participant_id":"alice","capabilities":[],"ttl_sec":300}'

curl -s http://127.0.0.1:8092/v1/sessions
curl -s http://127.0.0.1:8092/v1/history/sessions
```

Mint response: `{"token":"<payload_b64>.<mac_b64>"}` matching `Xylon.Control.Auth.HMAC`.

CORS defaults: `http://127.0.0.1:8080`, `:5173`, `:5174`.

Live **Presence** fills a roster table (participant / role / meta) from Control
JSON; Kick pre-fills moderate JSON for that id. **Fleet** tab shows node load
from SQLite (reachable column); summary includes healthy/unhealthy. Refresh now
uses `?refresh=1`.

## Lab exit (history + fleet + ops + drain)

1. Start admin (built UI), interface with `admin_webhook_url` -> this process and
   `fleet: true`, Xylon with Control (+ libcluster for dual-node).
2. Open `http://127.0.0.1:8092/`, Live: record; History: session after close.
3. Fleet: Refresh now -- one node in single-node lab; dual-node same
   `XYLON_REGION_ID` shows both nodes and non-zero participant totals when rooms
   are live. JSON has counts/ids only (no media fields). Accepting column +
   Drain / Undrain buttons set `accepting` via Control.
4. Ops: `curl .../v1/ops/summary?refresh=1` matches Fleet tab /
   `/v1/fleet/summary` totals. With load above `scale_out_participants`,
   `/v1/ops/recommendations/latest` returns `scale.recommend` / `scale_out`.
5. Drain: `POST /v1/ops/accepting` with `accepting:false`; new rooms land on
   other accepting nodes; wait until drained node `session_count` /
   `participant_count` are 0, then stop that BEAM node outside this stack.
6. Restart admin; History and Fleet cache still list prior SQLite rows until the
   next poll refreshes nodes (recommendations ring is in-memory only).
