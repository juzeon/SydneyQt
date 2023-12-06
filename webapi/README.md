# Sydney Web API

## Quick Start

```bash
go run ./webapi
```

Then the server will be running at <http://localhost:8080>.

## Environment Variables

- `PORT`: The port to listen on. Default: `8080`
- `ALLOWED_ORIGINS`: The allowed origins for CORS. Default: `*`
- `NO_LOG`: Whether to disable logging. Default: `false`
- `DEFAULT_COOKIES`: Default cookies to use, can be obtained by `document.cookie`. Default: `""`
- `HTTPS_PROXY` or `HTTP_PROXY`: The proxy to use for requests to Microsoft. Default: `""`

## Endpoints

### POST /conversation/new

Create a new conversation.

- **Request**:
  - Content-Type: `application/json`
  - Body:
    - `cookies`: `string` (Optional)
- **Response**:
  - Content-Type: `application/json`
  - Body: `CreateConversationResponse`

### POST /image/upload

Upload an image and return its URL.

- **Request**:
  - Content-Type: `multipart/form-data`
  - Body:
    - `image`: `File`
    - `cookies`: `string` (Optional)

- **Response**:
  - Content-Type: `text/plain`
  - Body: `string`

### POST /chat/stream

Start a chat stream.

- **Request**:
  - Content-Type: `application/json`
  - Body:
    - `conversation`
    - `prompt`: `string`
    - `context`: `string`
    - `cookies`: `string` (Optional)
    - `imageUrl`: `string` (Optional)
    - `noSearch`: `boolean` (Optional)
    - `conversationStyle`: `string` (Optional)
    - `locale`: `string` (Optional)

- **Response**:
  - Content-Type: `text/event-stream`
  - Body: Server-sent events
    - `event`: `string`
    - `data`: `string`
