# JSON API

## Overview

The API follows specification for native messaging from [Mozilla](https://developer.mozilla.org/en-US/Add-ons/WebExtensions/Native_messaging) and [Chrome](https://developer.chrome.com/apps/nativeMessaging). Each JSON-UTF8 encoded message is prefixed with a 32-bit integer specifying the length of the message. Communication is performed via stdin/stdout.

**WARNING**: This API **MUST NOT** be exposed over the network to remote hosts. **No authentication is performed** and the only safe way is to communicate via stdin/stdout as you do in your terminal.

The implementation is located in `utils/jsonapi`.

## Request Types

### `query`

#### Query:

```json
{
  "type": "query",
  "query": "secret"
}
```

#### Response:

```json
[
    "somewhere/mysecret/loginname",
    "somwhere/else/secretsauce"
]
```

### `queryHost`

Similar to `query` but cuts host names and sub domains from the left side until the response to the query is non-empty. Stops if only the [public suffix](https://publicsuffix.org/) is remaining.

#### Query:

```json
{
  "type": "queryHost",
  "host": "some.domain.example.com"
}
```

#### Response:

```json
[
    "somewhere/domain.example.com/loginname",
    "somwhere/other.domain.example.com"
]
```

### `getLogin`

#### Query:

```json
{
   "type": "getLogin",
   "entry": "somewhere/else/secretsauce"
}
```

#### Response:

```json
{
   "username": "hugo",
   "password": "thepassword"
}
```

## Error Response

If an uncaught error occurs, the stringified error message is send back as response:

```json
{
  "error": "Some error occurred with fancy message"
}
```
