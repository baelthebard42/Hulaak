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
Stores an event in the system database for later delivery processing.

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
Registers a webhook delivery endpoint for a specific event type.

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
