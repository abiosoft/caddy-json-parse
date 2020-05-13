package jsonvars

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

// Interface guards
var (
	_ caddy.Provisioner           = (*JSONMiddleware)(nil)
	_ caddyhttp.MiddlewareHandler = (*JSONMiddleware)(nil)
	_ caddyfile.Unmarshaler       = (*JSONMiddleware)(nil)
)

func init() {
	caddy.RegisterModule(JSONMiddleware{})
	httpcaddyfile.RegisterHandlerDirective("json_vars", parseCaddyfile)
}

// JSONMiddleware implements an HTTP handler that writes the
// visitor's IP address to a file or stream.
type JSONMiddleware struct {
	Strict string `json:"strict,omitempty"`

	strict bool
	log    *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (JSONMiddleware) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.json_vars",
		New: func() caddy.Module { return new(JSONMiddleware) },
	}
}

// Provision implements caddy.Provisioner.
func (m *JSONMiddleware) Provision(ctx caddy.Context) error {
	m.log = ctx.Logger(m)

	if m.Strict == "strict" {
		m.strict = true
	}
	return nil
}

func (m JSONMiddleware) replacerFunc(r *http.Request) (caddy.ReplacerFunc, error) {
	var v interface{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &v)
	if err != nil {
		return nil, err
	}

	// prevent repetitive parsing. cache values
	keys := map[string]interface{}{}

	return func(key string) (interface{}, bool) {
		prefix := "json."
		if !strings.HasPrefix(key, prefix) {
			return nil, false
		}
		key = strings.TrimPrefix(key, prefix)

		// use cache if previously fetched
		if val, ok := keys[key]; ok {
			return val, true
		}

		return fetchValue(v, key), true

	}, nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m JSONMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)

	replacerFunc, err := m.replacerFunc(r)
	if err != nil {
		if m.strict {
			return caddyhttp.Error(http.StatusBadRequest, err)
		}
		m.log.Debug("", zap.Error(err))
	}

	if err == nil {
		repl.Map(replacerFunc)
	}

	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (m *JSONMiddleware) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if d.Args(&m.Strict) {
			if m.Strict != "strict" {
				return d.Errf("unexpected token '%s'", m.Strict)
			}
		}
	}
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m JSONMiddleware
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}
