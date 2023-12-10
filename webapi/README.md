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
- `AUTH_TOKEN`: The Bearer token to access the API server. Default: `""`

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
    - `prompt`: `string`
    - `context`: `string`
    - `conversation`: `CreateConversationResponse` (Optional)
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

### POST /v1/chat/completions

This endpoint is compatible with the OpenAI API. You can check the API reference [here](https://platform.openai.com/docs/api-reference/chat).

Due to differences between the OpenAI API and the Sydney API, only the following parameters are supported:

- `messages`: The same as OpenAI's, and can contain image url (only valid in the last message).
- `model`: `GPT-3.5-Turbo` series will be mapped to `Balance`, others will be mapped to `Creative`.
- `stream`: The same as OpenAI's.

There is an extra field or reusing conversation, if your SDK supports such customization:

- `conversation`: `CreateConversationResponse`

The `Cookie` header is also supported to provide custom cookies.

The response is full of dummy values, and only the `choices` field is valid. The stop reason is `length` if any error occurs, and `stop` otherwise.
