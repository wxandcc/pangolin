package pangolin

import (
	"errors"
	"fmt"
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	"net"
	"strconv"
	"strings"
)

var servers []string
var rawServers []string

const defaultPort = "53"

// init registers this plugin.
func init() { plugin.Register("pangolin", setup) }

// setup is the function that gets called when the config parser see the token "example". Setup is responsible
// for parsing any extra options the example plugin may have. The first token this function sees is "example".
func setup(c *caddy.Controller) error {
	c.Next() // Ignore "example" and give us the next token.
	rawServers = make([]string, 0)
	for c.NextArg() {
		rawServers = append(rawServers, c.Val())
	}

	if err := validateServer(rawServers); err != nil {
		return plugin.Error("pangolin", err)
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

func validateServer(servers []string) error {
	kv := make(map[string]struct{})
	for _, server := range servers {
		ipPort := strings.Split(server, ":")
		if len(ipPort) == 1 {
			address := net.ParseIP(ipPort[0])
			if address == nil {
				return errors.New(fmt.Sprintf("the dns server address %s is not correct,Ip is not IPv4 Address [%s] illegal", server, ipPort[0]))
			}
			k := address.String() + ":" + defaultPort
			kv[k] = struct{}{}
		} else if len(ipPort) == 2 {
			address := net.ParseIP(ipPort[0])
			if address == nil {
				return errors.New(fmt.Sprintf("the dns server address %s is not correct,Ip is not IPv4 Address [%s] illegal", server, ipPort[0]))
			}
			port, err := strconv.Atoi(ipPort[1])
			if err != nil {
				return errors.New(fmt.Sprintf("the dns server address %s is not correct,port [%s] illegal", server, ipPort[1]))
			}
			if port > 0 && port < 65535 {
				k := address.String() + ":" + strconv.Itoa(port)
				kv[k] = struct{}{}
			} else {
				return errors.New(fmt.Sprintf("the dns server address %s is not correct,port [%d] not between 0 and 65535,not illegal", server, port))
			}
		} else {
			return errors.New(fmt.Sprintf("the dns server address %s is not correct,too many [:]s", server))
		}
	}

	if len(kv) > 0 {
		initServers(kv)
	} else {
		return errors.New("at lease one dns server is required")
	}

	return nil

}

func initServers(m map[string]struct{}) {
	servers = make([]string, len(m))
	i := 0
	for k, _ := range m {
		servers[i] = k
		i++
	}
}
