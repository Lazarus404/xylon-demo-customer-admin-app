# Xylon demo customer admin app

Dummy **SaaS customer** backend. Holds one customer's current shared secret and:

- Mints short-lived Xylon Control HMAC tokens for browsers (`getToken`)
- Forwards Control admin (presence, history, inject, moderate) to
  [xylon-minimal-customer-interface](https://github.com/Lazarus404/xylon-minimal-customer-interface)

Never returns the shared secret. Never calls Xylon HTTP directly -- admin goes
through the platform interface so customer status and feature flags can deny.

Related: [GOLANG_APPS.md](../GOLANG_APPS.md).

## Run

Start the platform interface on `:8093` (and its secret store) before exercising
admin routes. Then:

```bash
cd xylon-demo-customer-admin-app
go run ./cmd/admin -config config.example.json
```

Default listen: `127.0.0.1:8092` (loopback required; API is open on bind).

Lab secret in `config.example.json` is `lab-hmac-secret` (matches
`mix xylon.datachannel --control hmac` when Xylon uses the Config adapter).
With `XYLON_CONTROL_SECRETS_URL` / `Secrets.Local`, set `hmac_secret` to the
store's current secret for this `customer_id` (interface provision/rotate).
The test client Mint mode calls this service via SDK `getToken`.

### Secret handoff

1. Point the customer interface `admin_config_path` at a JSON file used by this
   app (or paste `hmac_secret` from provision/rotate into `config.example.json`).
2. **Restart** this process after the interface rewrites `hmac_secret` (no
   hot-reload in v1).

Interface-written files may contain only `customer_id` and `hmac_secret`;
`listen_addr`, `interface_base_url`, TTLs, and `cors_origins` fall back to
defaults on load.

## Config

| Key | Meaning |
| --- | --- |
| `customer_id` | Tenant id in minted tokens and interface admin paths |
| `hmac_secret` | Current signing secret for browser mint |
| `listen_addr` | Bind (must be loopback) |
| `interface_base_url` | Platform interface base (default `http://127.0.0.1:8093`) |
| `default_ttl_sec` / `max_ttl_sec` | Mint TTL |
| `cors_origins` | Browser `getToken` origins |

## HTTP API

| Method | Path | Notes |
| --- | --- | --- |
| GET | `/healthz` | `{"ok":true}` |
| POST | `/v1/token` | local mint |
| OPTIONS | `/v1/token` | CORS preflight |
| GET | `/v1/sessions/:sid/presence` | forward to interface |
| GET | `/v1/sessions/:sid/history` | forward; query `label` |
| POST | `/v1/sessions/:sid/data` | forward inject body |
| POST | `/v1/sessions/:sid/moderate` | forward moderate body |

Session admin paths use config `customer_id` (clients cannot pick another tenant).
Interface responses (status + body) are passed through. Dial failures return
`502` `{"error":"interface_unreachable"}`.

```bash
curl -s -X POST http://127.0.0.1:8092/v1/token \
  -H 'content-type: application/json' \
  -d '{"session_id":"room-1","participant_id":"alice","capabilities":[],"ttl_sec":300}'

curl -s http://127.0.0.1:8092/v1/sessions/room-1/presence

curl -s -X POST http://127.0.0.1:8092/v1/sessions/room-1/moderate \
  -H 'content-type: application/json' \
  -d '{"action":"kick","participant_id":"p2"}'
```

Mint response: `{"token":"<payload_b64>.<mac_b64>"}` matching `Xylon.Control.Auth.HMAC`.

CORS defaults: `http://127.0.0.1:8080`, `http://127.0.0.1:5173`.
