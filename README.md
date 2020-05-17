[![Go](https://github.com/abiosoft/caddy-json-parse/workflows/Go/badge.svg)](https://github.com/abiosoft/caddy-json-parse/actions)

# caddy-json-parse
Caddy v2 module for parsing json request body.

## Installation

```
xcaddy build v2.0.0 \
    --with github.com/abiosoft/caddy-json-parse
```


## Usage

`json_parse` parses the request body as json for reference as [placeholders](https://caddyserver.com/docs/caddyfile/concepts#placeholders).

### Caddyfile

Simply use the directive anywhere in a route. If set, `strict` responds with bad request if the request body is an invalid json.
```
json_parse [<strict>]
```

And reference variables via `{json.*}` placeholders. Where `*` can get as deep as possible. e.g. `{json.items.0.label}`


#### Example

Run a [command](https://github.com/abiosoft/caddy-exec) only if the github webhook is a push on master branch.
```
@webhook {
    expression {json.ref}.endsWith('/master')
}
route {
    json_parse # enable json parser
    exec @webhook git pull origin master
}
```

### JSON

`json_parse` can be part of any route as an handler

```jsonc
{
  ...
  "routes": [
    {
      "handle": [
        {
          "handler": "json_parse",

          // if set to true, returns bad request for invalid json
          "strict": false 
        },
        ...
      ]
    },
  ...
  ]
}
```

## License

Apache 2
