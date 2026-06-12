# Hulaak

Hulaak is an event-driven, authenticated webhook delivery & retry system built on Go. It ingests events from your applications and delivers them to the required destination with **at-least-once** delivery guarentee. It pairs Go's concurrency with NATS JetStream's persistence and backoff-based redelivery to ensure webhooks are retried until the destination receives them. The system is split into **two deployable services** that communicate exclusively over NATS JetStream, and is made ready and tested to run on a Kubernetes cluster. Check out the `infra/` directory for details regarding the Kubernetes setup.


## Architecture / Flow of Events

Hulaak is composed of two deployable services backed by **PostgreSQL** (state) and **NATS JetStream** (messaging). All cross-service communication happens through NATS — there are no direct service-to-service calls and only the Control plane talks to the database.

<img width="641" height="669" alt="hulaak_architecture drawio" src="https://github.com/user-attachments/assets/dabceeb9-c4f5-4d30-8adc-faf479786f66" />


1. **Control Plane** (`control/`): The stateful brain of the system. It runs three concurrent components inside a single process:
   - **HTTP API server** — ingests and authenticates incoming events, registers destination endpoints, and persists everything to Postgres. On each accepted event it writes the `events`, `delivery`, and `outbox` rows in a single transaction (the *transactional outbox* pattern).
   - **Runner** (`internal/workers/runner.go`) — a background goroutine that continuously claims fresh rows from the `outbox` table (`FOR UPDATE SKIP LOCKED`), publishes them to NATS, and deletes the claimed rows. 
   - **Listener** (`internal/workers/listener.go`) — a background goroutine that subscribes to delivery-result events coming back from Worker-Destination and updates the `delivery` table (attempts, last error, last attempt time, status), with duplicate-event detection.

2. **Worker-Destination** (`worker-destination/`): A stateless delivery worker. It pulls delivery payloads from NATS, performs the outbound HTTP `POST` to the customer's endpoint, and **publishes the result back to NATS** rather than touching the database directly. Because it holds no state, its replica count can be scaled up freely with load.

### Messaging model

A single JetStream stream named **`DELIVERIES`** (subjects `webhook.>`, file storage) carries two subjects, each consumed by its own durable pull consumer:

| Subject | Direction | Producer | Consumer (durable) |
|---|---|---|---|
| `webhook.delivery.payload` | Control → Worker-Destination | Runner | `worker-destination` |
| `webhook.delivery.state` | Worker-Destination → Control | Worker-Destination | `control` |

This gives a closed feedback loop: Control hands off a delivery, Worker-Destination attempts it and reports back, and Control records the outcome — keeping the delivery worker completely free of database access.





## Retry Mechanism

The retry mechanism relies on NATS JetStream's persistence and redelivery features. The `worker-destination` durable consumer is configured with explicit acknowledgements, `MaxDeliver = 10`, and a backoff schedule of `1m, 2m, 4m, 8m, 16m, 30m, 60m`.

### Success Case 


![retry_sucess](https://github.com/user-attachments/assets/8d1ccd78-3739-484f-a788-9c33f15a7b99)

For each payload JetStream delivers to Worker-Destination, the worker acknowledges it back to NATS if and only if the outbound delivery succeeded. It also publishes a `webhook.delivery.state` event so the Control plane's Listener marks the `delivery` row as `success`. Story ends.

### Failure Case 

![retry](https://github.com/user-attachments/assets/071c5a88-56e6-44af-b3a9-f21935e56a9a)


If the delivery fails, Worker-Destination publishes a failure result (consumed by Control to bump `num_attempts` and record `last_error`) and **does not acknowledge** the original payload message. JetStream then redelivers the message after the next backoff interval, and the worker tries again — so the wait between attempts grows toward the 60-minute ceiling.

To avoid perpetual redelivery, the consumer's `MaxDeliver` (10) caps the number of attempts. Once that limit is reached, JetStream stops redelivering the message; the `delivery` row retains its full attempt history (`num_attempts`, `last_error`, `last_attempt_at`) for inspection.

> **Note:** An explicit Dead Letter Queue (DLQ) and an automatic `failed`-status transition after the max attempts are not yet implemented in the current code — they are on the roadmap. Today, exhausted deliveries simply stop being redelivered while their history remains in the `delivery` table.


## Data Model

All tables live in the Control plane's Postgres database (see `control/database/migrations/`):

- **`client_user`** — registered API accounts (username, email, hashed password).
- **`events`** — every ingested event (type, source user, destination reference, JSON payload).
- **`endpoints`** — the target webhook URL registered for a given `(destination_reference, event_type)` pair.
- **`delivery`** — one row per delivery attempt lifecycle: `status` (`pending`/`success`/`failed`), `num_attempts`, `last_error`, `last_attempt_at`. This is the source of truth the Listener updates.
- **`outbox`** — the transactional outbox the Runner drains; rows move from `unprocessed` → `processing` and are deleted once published.


## Features

- **Separated concerns**: A stateful Control plane (ingestion + state) and a stateless delivery worker, decoupled entirely through NATS.

- **At-least-once delivery**: Guarantees the destination will receive each webhook at least once, given it is set up to listen correctly.

- **Transactional outbox**: Events and their outbox entries are written in a single Postgres transaction, so an accepted event is never lost between ingestion and publishing.

- **Status Tracking**: The `delivery` table exposes the details needed to track each event — number of attempts, last attempt time, last error, and current status.

- **Backoff-based retries**: JetStream redelivers failed messages on an increasing backoff schedule, which is convenient when a destination is briefly unavailable.

- **Scalability**: Worker-Destination is stateless, so its replica count can be increased freely to absorb large volumes of deliveries.






# Hulaak System API Documentation

## Authentication Model

- Uses JWT-based authentication
- Token is stored in HTTP cookie:
  - Cookie name: `access_token`
  - HttpOnly: true
  - Secure: true
  - SameSite: Lax

Token validation checks:
- Signature → HMAC using environment variable `JWT_KEY`
- Expiration → `exp` claim
- Subject → `sub` claim is used as `userID`

---

## Endpoint Reference

### 0. Health Check

#### `GET /healthz`

Description:
Liveness/readiness probe for the Control plane.

Authentication:
Not required.

---

### 1. Create Account

#### `POST /account`

Description:
Creates a new client user account.

Authentication:
Not required.

Request Body:
```json
{
  "username": "string",
  "email": "string",
  "password": "string"
}
```

| Field | Required | Type |
|---|---|---|
| username | Yes | string |
| email | Yes | string |
| password | Yes | string |

Response:

Status: `201 Created`

```json
{
  "id": "string",
  "email": "string",
  "username": "string"
}
```

Errors:
| Status | Meaning |
|---|---|
| 400 | Missing or invalid request body |
| 500 | Account creation failure |

---

## 2. Login User

### `POST /login`

Description:
Authenticates user credentials and issues JWT token.

Authentication:
Not required.

Request Body:
```json
{
  "username": "string",
  "password": "string"
}
```

Success Response:

Status: `200 OK`

```json
{
  "token": "jwt_token_string"
}
```

Side Effects:
- Sets HTTP cookie:
  - access_token = JWT token
  - HttpOnly = true
  - Secure = true
  - Path = "/"
  - MaxAge = 86400 seconds

Errors:
| Status | Meaning |
|---|---|
| 400 | Invalid request or authentication failure |

---

## 3. Receive Event

### `POST /events`

Description:
Stores an event in the system database for later delivery processing. This is where you send your events to kickstart the webhook delivery process.

Authentication:
Required (JWT cookie authentication).

Request Body:
```json
{
  "event_type": "string",
  "event_destination": "string",
  "payload": {}
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| event_type | string | Yes | Type/category of event |
| event_destination | string | Yes | Delivery destination reference |
| payload | JSON | Yes | Arbitrary event payload |

Notes:
- Payload is stored as raw JSON.
- Empty payload is rejected.

Context Processing:
- Event is associated with authenticated user ID.
- Request timeout: 5 seconds.

Response:
Status: `201 Created`

Returns stored event record (structure depends on internal service layer).

Errors:
| Status | Meaning |
|---|---|
| 400 | Missing fields / database insertion failure |
| 401 | Authentication failure |
| 500 | Internal service error |

---

## 4. Register Endpoint Destination

### `POST /endpoint`

Description:
Registers a webhook delivery endpoint for a specific event type. Do this before you start sending the events.

Authentication:
Required (JWT cookie authentication).

Request Body:
```json
{
  "destination_ref": "string",
  "event_type": "string",
  "endpoint": "string"
}
```

| Field | Required | Description |
|---|---|---|
| destination_ref | Yes | Logical destination identifier |
| event_type | Yes | Event category |
| endpoint | Yes | Target webhook URL |

Response:

Status: `201 Created`

```json
{}
```

Errors:
| Status | Meaning |
|---|---|
| 400 | Invalid request or missing fields |

---

## Middleware Security Behavior

Authentication Middleware protects:
- `/events`
- `/endpoint`

Failure conditions:
- Missing cookie → 400
- Invalid JWT → 400 / 401
- Expired JWT → 400

---

## Timeout Policy

All service calls are wrapped in 5 second context timeout.

---

## System Summary

| Feature | Description |
|---|---|
| Services | Control plane (stateful) + Worker-Destination (stateless) |
| API layering | Handler → Service → Repository → Postgres |
| Messaging | NATS JetStream, single `DELIVERIES` stream, two durable pull consumers |
| Event Model | Transactional outbox: `events` + `delivery` + `outbox` written atomically |
| Delivery feedback | Worker-Destination reports results over NATS; Control's Listener updates `delivery` |
| Auth Model | Cookie-based JWT authentication |
| Payload Type | Raw JSON payload storage (`JSONB`) |
| Processing Style | Asynchronous: outbox publisher (Runner) + delivery worker + result listener |
| Delivery guarantee | At-least-once, with `MaxDeliver = 10` and backoff redelivery |


## Deployment

Both services are containerized (multi-stage builds → distroless images) and deployed to Kubernetes via [Skaffold](skaffold.yaml). The manifests in [infra/k8s/](infra/k8s/) define:

- `nats` — a single-replica JetStream `StatefulSet` with a 5Gi persistent volume, plus a headless service and a `nats-client` service on port `4222`.
- `control-depl` — the Control plane `Deployment` + a `NodePort` service (`30008`), wired to `JWT_SECRET`, `DATABASE_URL`, and `NATS_URL`.
- `worker-destination-depl` — the Worker-Destination `Deployment` + a `NodePort` service (`30009`), wired to `DATABASE_URL` and `NATS_URL`.

Database migrations are managed with [`sql-migrate`](run.md) (`sql-migrate up`). Required environment variables: `DATABASE_URL`, `JWT_SECRET` (Control only), and `NATS_URL`.
