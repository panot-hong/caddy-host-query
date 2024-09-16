package hostquery

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "net/url"

    "github.com/caddyserver/caddy/v2"
    "github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
    "github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
    caddy.RegisterModule(HostQuery{})
    httpcaddyfile.RegisterHandlerDirective("caddy-host-query", bodyParseCaddyfile)
}

// HostQuery is a Caddy module that makes an API request to get the host.
type HostQuery struct {
    APIURL              string `json:"api_url,omitempty"`
    DEFAULT_HTTPS_SCHEME bool   `json:"default_https_scheme,omitempty"`
    DEFAULT_UPSTREAM    string `json:"default_upstream,omitempty"`
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

    fmt.Println("Default upstream:", m.DEFAULT_UPSTREAM)
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
    // If the new host is empty, set it to the default upstream
    if newHost == "" {
        newHost = m.DEFAULT_UPSTREAM
    }

    // Ensure the new host is correctly formatted
    parsedURL, err := url.Parse(newHost)
    if err != nil {
        return fmt.Errorf("invalid host URL: %v", err)
    }

    // Set the new host as a variable in the request context
    caddyhttp.SetVar(r.Context(), "shard.upstream", parsedURL.String())
    fmt.Printf("Default HTTPS Scheme: %v\n", m.DEFAULT_HTTPS_SCHEME)
    // Determine if the port is 443
    port := parsedURL.Port()
    if port == "" {
        port = "80" // Default port if not specified
        if parsedURL.Scheme == "https" || (parsedURL.Scheme == "" && m.DEFAULT_HTTPS_SCHEME) {
            port = "443"
            if parsedURL.Host == "" {
                caddyhttp.SetVar(r.Context(), "shard.upstream", parsedURL.String() + ":" + port)
            } else {
                caddyhttp.SetVar(r.Context(), "shard.upstream", parsedURL.Host + ":" + port)
            }
        }
    }

    // Set a variable indicating whether the port is 443
    isPort443 := port == "443"
    caddyhttp.SetVar(r.Context(), "shard.upstream.is_port_443", isPort443)

    // Read back the variables and log them
    upstream := caddyhttp.GetVar(r.Context(), "shard.upstream")
    fmt.Printf("Upstream: %s\n", upstream)
    fmt.Printf("Resolved Port: %s\n", port)
    fmt.Printf("Is Port 443: %v\n", isPort443)

    // Execute the next handler
    return next.ServeHTTP(w, r)
}

func (p HostQuery) Validate() error {
	if p.APIURL == "" {
		return fmt.Errorf("missing `api_url` in `caddy-host-query`")
	}

	return nil
}

// UnmarshalCaddyfile sets up the handler from Caddyfile tokens.
func (m *HostQuery) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
    m.DEFAULT_HTTPS_SCHEME = true // set default value
    d.Next()
    var defaultHttpsSchema string
    for d.NextBlock(0) {
		switch d.Val() {
		case "api_url":
			if !d.AllArgs(&m.APIURL) {
				return d.ArgErr()
			}
		case "default_upstream":
			if !d.AllArgs(&m.DEFAULT_UPSTREAM) {
				return d.ArgErr()
			}
        case "default_https_scheme":
            if !d.AllArgs(&defaultHttpsSchema) {
                return d.ArgErr()
            }
            m.DEFAULT_HTTPS_SCHEME = defaultHttpsSchema == "true"
		default:
			return fmt.Errorf("unknown option `%s` in `caddy-host-query`", d.Val())
		}
	}

    if m.APIURL == "" {
        return d.Err("missing API URL")
    }
    return nil
}

func bodyParseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
    handler := &HostQuery{}
	err := handler.UnmarshalCaddyfile(h.Dispenser)
	return handler, err
}

// Interface guards
var (
    _ caddyhttp.MiddlewareHandler = (*HostQuery)(nil)
    _ caddyfile.Unmarshaler       = (*HostQuery)(nil)
)