package jsonvars

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

// Interface guards
var (
	_ caddy.Provisioner           = (*JSONVars)(nil)
	_ caddyhttp.MiddlewareHandler = (*JSONVars)(nil)
	_ caddyfile.Unmarshaler       = (*JSONVars)(nil)
)

func init() {
	caddy.RegisterModule(JSONVars{})
	httpcaddyfile.RegisterHandlerDirective("json_vars", parseCaddyfile)
}

// JSONVars implements an HTTP handler that parses
// json body as placeholders.
type JSONVars struct {
	Strict string `json:"strict,omitempty"`

	strict bool
	log    *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (JSONVars) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.json_vars",
		New: func() caddy.Module { return new(JSONVars) },
	}
}

// Provision implements caddy.Provisioner.
func (m *JSONVars) Provision(ctx caddy.Context) error {
	m.log = ctx.Logger(m)

	if m.Strict == "strict" {
		m.strict = true
	}
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (m JSONVars) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)

	replacerFunc, err := newReplacerFunc(r)
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
func (m *JSONVars) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
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
	var m JSONVars
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}
