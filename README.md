# Hulaak

Hulaak is an event-driven, authenticated webhook delivery & retry system built on Go. It ingests the events from your applications and sends it to the required destination with **at-least-once** delivery. It makes use of Go's seamless networking with NATS JetStream's persistence, exponential backoff time to ensure the webhooks are retried till the destination receives it. This entire system, comprising of three separate services, is made ready and is well tested to run on a Kubernetes cluster. Check out the infra directory for details regarding the Kubernetes setup.

## Features

- **Separated concerns**: Separate services to handle event ingestion and deliveries (check out the system architecture below!)

- **At-least-once delivery**: Guarentee that the destination will receive the webhooks at least once, given it is setup to listen to them correctly.

- **Status Tracking**: The system exposes necessary details required for the client to track the status of their events including number of attempts, last retry time, last error and so on.

- **Exponential backoff retries** : The time to wait before retrying increases exponentially on each retry, making it convenient if your destination system shuts for a brief period of time.

- **Scalability**: The workers for event processing and deliveries can be increased easily if needed in context of large number of incoming events.




## Architecture / Flow of Events

Hulaak is made up of three separate services that have a well defined role to play for the overall flow.

1. **Control Plane**: This service is responsible for ingesting and authenticating the incoming events, persisting them to necessary tables, sending back responses to client and maintaining the outbox table, where the fresh events are kept for workers to fetch from.

2. **Worker-NATS**: This is an independent worker that continuously fetches deliveries from the outbox table, processes it and sends the delivery details to Worker-Destination through NATS JetStream. Number of instances of this worker can be increased as per the load.

3. **Worker-Destination**: This is another independent worker that receives delivery details from Worker-NATS and sends it to the required destination, and performs retry if needed. This is also responsible for updating the status of each  event delivery. Again, the number of instances of this worker can too be increased as per load.


![hulaak (1)](https://github.com/user-attachments/assets/bc948fec-5911-492a-b4bf-d8d4e002c9a3)


## Retry Mechanism

The retry mechanism relies on NATS JetStream's persistence and resending features. 

### Success Case 


![retry_sucess](https://github.com/user-attachments/assets/8d1ccd78-3739-484f-a788-9c33f15a7b99)

Specifically, for each event the JetStream sends to Worker-Destination, the Worker-Destination acknowledges it back to NATS if and only if the delivery to destination is successful. In that case, the status of the delivery is set to 'successful', story ends.

### Failure Case 

![retry](https://github.com/user-attachments/assets/071c5a88-56e6-44af-b3a9-f21935e56a9a)


If the delivery for an event is unsuccessful, the Worker-Destination sends back no acknowledgement for the event. In that case, NATS JetStream is configured to resend the event to Worker-Destination on an exponential backoff basis after which the worker retries the delivery. Meaning for each successive retry, the time to try again increases exponentially from previous.

Note that to ensure there is no perpetual resending of events, a mechanism to stop the retry after a MAXIMUM_RETRIES is implemented. After MAXIMUM_RETRIES, the delivery row is marked as 'failed' and is move to a Dead Letter Queue (DLQ) for manual inspection.




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
| Architecture | Handler → Service → Database |
| Event Model | Stored event ingestion for webhook delivery |
| Auth Model | Cookie-based JWT authentication |
| Payload Type | Raw JSON payload storage |
| Processing Style | Likely asynchronous downstream worker delivery |
