package pangolin

import (
	"errors"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
)

var servers []string

// init registers this plugin.
func init() { plugin.Register("pangolin", setup) }

// setup is the function that gets called when the config parser see the token "example". Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is "example".
func setup(c *caddy.Controller) error {
	c.Next() // Ignore "example" and give us the next token.
	servers = make([]string, 0)
	for c.NextArg() {
		servers = append(servers, c.Val())
	}

	if len(servers) <= 0 {
		return plugin.Error("pangolin", errors.New("pangolin requires at least one dns server"))
	} else {
		log.Infof("the server list is %+v", servers)
	}

	// Add the Plugin to CoreDNS, so Servers can use it in their plugin chain.
	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		return Pangolin{Next: next}
	})

	// All OK, return a nil error.
	return nil
}
