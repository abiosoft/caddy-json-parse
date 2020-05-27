package jsonparse

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
	_ caddy.Provisioner           = (*JSONParse)(nil)
	_ caddyhttp.MiddlewareHandler = (*JSONParse)(nil)
	_ caddyfile.Unmarshaler       = (*JSONParse)(nil)
)

func init() {
	caddy.RegisterModule(JSONParse{})
	httpcaddyfile.RegisterHandlerDirective("json_parse", parseCaddyfile)
}

// JSONParse implements an HTTP handler that parses
// json body as placeholders.
type JSONParse struct {
	Strict bool `json:"strict,omitempty"`
	log    *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (JSONParse) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.json_parse",
		New: func() caddy.Module { return new(JSONParse) },
	}
}

// Provision implements caddy.Provisioner.
func (j *JSONParse) Provision(ctx caddy.Context) error {
	j.log = ctx.Logger(j)

	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (j JSONParse) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)

	replacerFunc, err := newReplacerFunc(r)
	if err != nil {
		if j.Strict {
			return caddyhttp.Error(http.StatusBadRequest, err)
		}
		j.log.Debug("", zap.Error(err))
	}

	if err == nil {
		repl.Map(replacerFunc)
	}

	return next.ServeHTTP(w, r)
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (j *JSONParse) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		args := d.RemainingArgs()
		switch len(args) {
		case 0:
		case 1:
			if args[0] != "strict" {
				return d.Errf("unexpected token '%s'", args[0])
			}
			j.Strict = true
		default:
			return d.ArgErr()
		}
	}
	return nil
}

// parseCaddyfile unmarshals tokens from h into a new Middleware.
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var m JSONParse
	err := m.UnmarshalCaddyfile(h.Dispenser)
	return m, err
}
