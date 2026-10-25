# Weekly evidence record

Student number:
Week:w
Repository URL:
Immutable tag:
Commit hash:
Go version / OS:

## Commands actually executed

## Observed results (do not paste expected examples as results) 
# API and Protocol Contract

## 1. Overview

The status API exposes a resource representing the current operational status of a node. The endpoint is:

`/v1/status`

The API supports retrieving the current status and updating it through an HTTP request. Each node maintains its own status independently.

## 2. Supported Status Values

The only permitted status values are:

- `healthy`
- `degraded`

Any other status value is invalid and must be rejected without changing the current resource.

## 3. Request and Response Matrix

| Request | Condition | Expected HTTP status |
|---|---|---:|
| GET `/v1/status` | Retrieve current status | 200 OK |
| POST `/v1/status` | Valid JSON and permitted status | 200 OK |
| Unsupported method, e.g. DELETE | Method not allowed | 405 Method Not Allowed |
| POST with unsupported content type | Content type is not accepted | 415 Unsupported Media Type |
| POST with malformed JSON | Request body cannot be decoded | 400 Bad Request |
| POST with unknown JSON fields | Body contains disallowed fields | 400 Bad Request |
| POST with invalid status value | Domain validation fails | 400 Bad Request |
| POST with body exceeding the limit | Request exceeds 4096 bytes | 413 Content Too Large |

The server must implement the response codes consistently with this contract.

## 4. Request Examples

### GET: Retrieve the current status

Request:

`GET /v1/status`

Example response:

`HTTP/1.1 200 OK`

```json
{
  "status": "healthy"
}
```

The response represents the current status stored by the node.

### POST: Update the status

Request:

`POST /v1/status`

Content-Type: `application/json`

```json
{
  "status": "degraded"
}
```

Example successful response:

`HTTP/1.1 200 OK`

```json
{
  "status": "degraded"
}
```

A successful update changes the stored resource to the requested permitted value.

## 5. Error Response Format

API errors use a JSON object containing an error label:

```json
{
  "error": "<label>"
}
```

The label identifies the error condition. Examples may include `method_not_allowed`, `unsupported_media_type`, and `invalid_status`, depending on the implementation's defined error labels.

## 6. Request Body Limit

The maximum accepted request body size is 4096 bytes. A request body exceeding this limit must be rejected with HTTP 413.

The server must enforce the limit while reading the body rather than relying only on the declared Content-Length header.

## 7. State Preservation Invariant

A rejected request must leave the previously stored status unchanged. The server must complete request decoding and domain validation before committing the new value.

This invariant applies to malformed JSON, unknown fields, invalid status values, unsupported media types, oversized bodies, and other rejected requests.

## 8. Known JSON Decoding Limitation

Go's standard `encoding/json` decoder does not reject duplicate JSON object keys by default. For duplicate keys, later values generally replace or merge with earlier values according to the target type.

Consequently, a request containing duplicate `status` keys may be accepted rather than rejected as ambiguous. Strict duplicate-key rejection requires additional validation beyond the standard decoder's default behaviour.

## 9. Scope

The status endpoint reports and updates the node's application-level status. The `/healthz` endpoint separately indicates whether the process is responding to health-check requests. A successful health check does not, by itself, prove that every application function or dependency is healthy.

## Failure case and diagnosis

## Design decision and alternative

### Design Rationale — Week 2

The status resource endpoint provides a clear separation between retrieving a resource and updating it. Using GET to read the current status and POST to submit an update makes the API easier to understand, test, and maintain than a single endpoint that performs multiple unrelated actions.

The implementation distinguishes different request failures. HTTP 413 rejects bodies that exceed the size limit, while HTTP 415 rejects unsupported media types. HTTP 400 covers decoding failures and invalid domain values, including the `invalid_status` condition. These distinctions make client errors easier to diagnose.

POST returns HTTP 200 because the server successfully processes the update and returns the resulting status representation. The response confirms that the requested operation was completed.

The 4096-byte body limit protects the server from unnecessarily large requests, but it cannot replace a timeout. A small request can still arrive very slowly or take too long to process. Body-size limits and timeouts therefore address different resource risks and should be used together.

Finally, the handler validates the request before changing stored state. This ensures that rejected requests do not overwrite a previously valid status and preserves predictable behaviour for clients.

## Remaining limitations

## Sources and AI-use disclosure
