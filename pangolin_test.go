package pangolin

import (
	"context"
	"github.com/coredns/caddy"
	"testing"

	"github.com/coredns/coredns/plugin/pkg/dnstest"
	"github.com/coredns/coredns/plugin/test"

	"github.com/miekg/dns"
)

func TestPangolin(t *testing.T) {
	// Create a new Pangolin Plugin. Use the test.ErrorHandler as the next plugin.
	x := Pangolin{Next: test.ErrorHandler()}
	c := caddy.NewTestController("dns", `pangolin 114.114.114.114:53`)
	setup(c)
	ctx := context.TODO()
	r := new(dns.Msg)
	r.SetQuestion("baidu.com", dns.TypeA)
	// Create a new Recorder that captures the result, this isn't actually used in this test
	// as it just serves as something that implements the dns.ResponseWriter interface.
	rec := dnstest.NewRecorder(&test.ResponseWriter{})

	// Call our plugin directly, and check the result.
	x.ServeDNS(ctx, rec, r)
}
