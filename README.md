# Caddy Custom Module Setup

This guide will help you set up your development environment to build and develop a custom Caddy module.

## Prerequisites

1. **Go**: Ensure you have Go installed on your machine. You can download it from [golang.org](https://golang.org/dl/).
2. **xcaddy** Install at https://github.com/caddyserver/xcaddy to build the custom build of Caddy. It is also required to have Go installed on the machine. See above for the installation guide or in the given link.

## Setup Instructions

`xcaddy` is a tool provided by the Caddy team to build Caddy with custom plugins. Install xcaddy by running:
    
```bash
go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
```

## Building Caddy with Custom Module
To build Caddy with your custom module, use the following command:
    
```bash
xcaddy build --with github.com/panot-hong/caddy-host-query
```

Or if you need to modify and build locally then use the following command:
    
```bash
xcaddy build --with local=./
```

## Cross-Compile for Linux (Optional)
In case you are on Windows and want to build Caddy for Linux, you can use the following command:
    
```bash
set GOOS=linux
set GOARCH=amd64
xcaddy build --with github.com/panot-hong/caddy-host-query
```

## Running Caddy
After building Caddy, you can run it using the following command:
    
```bash
./caddy run
```
or if you place `Caddyfile` in another directory, you can run it using the following command:
    
```bash
./caddy run --config ./test/Caddyfile
```

## Usage of this module in Caddyfile Configuration
The resolved host will be stored in the `X-Resolved-Host` header.

Here is an example `Caddyfile` configuration to use your custom module:
```caddy
{
    debug # enable debug mode
    order caddy-host-query before reverse_proxy
}

:80 {
    handle {
        caddy-host-query {
            api_url http://localhost:5214/get-actual-host
            default_upstream localhost:5215
        }
        @https {
            expression {http.vars.shard.upstream.is_port_443} == true
        }
        
        reverse_proxy @https {
            to {http.vars.shard.upstream}
            transport http {
                tls
            }
        }

        reverse_proxy {
            to {http.vars.shard.upstream}
        }
    }
}
```

### Configuration Options
- `api_url`: The URL to the API that will resolve the actual host. The API should return the actual host in the response body. This library expect the response to be in JSON format with the following structure:
    ```json
    {
        "host": "actual-host"
    }
    ```
    The actual host should be in format of `host:port` for HTTP for example `localhost:8080` and `https://host` for HTTPS for example `https://domain.com`.
- `default_upstream`: The default upstream to use if the host cannot be resolved and return empty string.

## Mocking the Host Query

ํYou can run mock server to test the host query by running the following command:
```bash
cd test
go run mock_server.go
```

Then run the caddy server and point to the Caddyfile in the test directory.
```bash
./caddy run --config ./test/Caddyfile
```


## Development Workflow
1. Make changes to your custom module.
2. Rebuild Caddy using xcaddy build --with github.com/panot-hong/caddy-host-query.
3. Test your changes by running the caddy binary.