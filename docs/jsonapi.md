# JSON API

## Overview

The API follows specification for native messaging from [Mozilla](https://developer.mozilla.org/en-US/Add-ons/WebExtensions/Native_messaging) and [Chrome](https://developer.chrome.com/apps/nativeMessaging). 
Each json-utf8 encoded message is prefixed with a 32 bit integer specifying the length of the message. 
Communication is performed via stdin/stdout. Currently, only a single request is repsonded `gopass jsonapi` call.

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

Similar to `query` but cuts hostnames and subdomains from the left side until the response to the query is non-empty. Stops if only one dot (domain + tld) is remaining.

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
  "error": "Some error occured with fancy message"
}
```


