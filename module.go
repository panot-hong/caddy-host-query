package hostquery

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"

    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
    caddy.RegisterModule(HostQuery{})
}

// HostQuery is a Caddy module that makes an API request to get the host.
type HostQuery struct {
    APIURL string `json:"api_url"`
}

// CaddyModule returns the Caddy module information.
func (HostQuery) CaddyModule() caddy.ModuleInfo {
    return caddy.ModuleInfo{
        ID:  "http.handlers.caddy-host-query",
        New: func() caddy.Module { return new(HostQuery) },
    }
}

// ServeHTTP implements the caddyhttp.MiddlewareHandler interface.
func (m HostQuery) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
    host := r.Host

    // Make the API request
    apiURL, err := url.Parse(m.APIURL)
    if err != nil {
        return fmt.Errorf("invalid API URL: %v", err)
    }
    query := apiURL.Query()
    query.Set("domain", host)
    apiURL.RawQuery = query.Encode()

    resp, err := http.Get(apiURL.String())
    if err != nil {
        return fmt.Errorf("failed to make API request: %v", err)
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("failed to read API response: %v", err)
    }

    var result map[string]string
    if err := json.Unmarshal(body, &result); err != nil {
        return fmt.Errorf("failed to parse API response: %v", err)
    }

    newHost, ok := result["host"]
    if !ok {
        return fmt.Errorf("API response does not contain 'host' field")
    }

    // Set the new host as a request header
    r.Header.Set("X-Resolved-Host", newHost)

    // Call the next handler in the chain
    return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile sets up the handler from Caddyfile tokens.
func (m *HostQuery) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
    for d.Next() {
        if !d.Args(&m.APIURL) {
            return d.ArgErr()
        }
    }
    return nil
}

// Interface guards
var (
    _ caddyhttp.MiddlewareHandler = (*HostQuery)(nil)
    _ caddyfile.Unmarshaler       = (*HostQuery)(nil)
)