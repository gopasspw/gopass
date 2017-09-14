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




