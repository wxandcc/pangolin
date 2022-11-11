// Package pangolin is a CoreDNS plugin that prints "Pangolin" to stdout on every packet received.
//
// It serves as a Pangolin CoreDNS plugin with numerous code comments.
package pangolin

import (
	"context"
	"net"
	"time"

	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"

	"github.com/miekg/dns"
)

// Define log to be a logger with the plugin name in it. This way we can just use log.Info and
// friends to log.
var log = clog.NewWithPlugin("Pangolin")

// Pangolin is an Pangolin plugin to show how to write a plugin.
type Pangolin struct {
	Next plugin.Handler
}

type dnsQueryResponse struct {
	ips  []net.IPAddr
	name string
	err  error
}

// ServeDNS implements the plugin.Handler interface. This method gets called when Pangolin is used
// in a Server.
func (e Pangolin) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	// This function could be simpler. I.e. just fmt.Println("Pangolin") here, but we want to show
	// a slightly more complex Pangolin as to make this more interesting.
	// Here we wrap the dns.ResponseWriter in a new ResponseWriter and call the next plugin, when the
	// answer comes back, it will print "Pangolin".

	// Debug log that we've seen the query. This will only be shown when the debug plugin is loaded.
	log.Debug("Received response")

	var c = make(chan dnsQueryResponse, len(servers))
	cancelCtx, cancel := context.WithCancel(context.TODO())

	for _, dnsServer := range servers {
		go queryDns(r.Question[0].Name, dnsServer, cancelCtx, c)
	}

	for range servers {
		select {
		case rt := <-c:
			if rt.err == nil && len(rt.ips) > 0 {
				cancel()
				msg := newMsg(r, rt)
				err := w.WriteMsg(msg)
				if err != nil {
					log.Errorf("write response raise error %s", err)
					return plugin.NextOrFailure(e.Name(), e.Next, ctx, w, r)
				}
				return dns.RcodeSuccess, nil
			}
		}
	}

	// Export metric with the server label set to the current server handling the request.
	//requestCount.WithLabelValues(metrics.WithServer(ctx)).Inc()
	cancel()
	return plugin.NextOrFailure(e.Name(), e.Next, ctx, w, r)
}

func newMsg(r *dns.Msg, res dnsQueryResponse) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true
	for _, v := range res.ips {
		msg.Answer = append(msg.Answer, &dns.A{
			Hdr: dns.RR_Header{Name: res.name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
			A:   v.IP,
		})
	}
	return msg
}

// Name implements the Handler interface.
func (e Pangolin) Name() string { return "Pangolin" }

// ResponsePrinter wrap a dns.ResponseWriter and will write Pangolin to standard output when WriteMsg is called.
type ResponsePrinter struct {
	dns.ResponseWriter
}

// WriteMsg calls the underlying ResponseWriter's WriteMsg method and prints "Pangolin" to standard output.
func (r *ResponsePrinter) WriteMsg(res *dns.Msg) error {
	log.Info("Pangolin")
	return r.ResponseWriter.WriteMsg(res)
}

// queryDns 根据host和dns server查询ip
func queryDns(name, dns string, ctx context.Context, c chan<- dnsQueryResponse) {

	log.Debugf("query %s from dns server %s", name, dns)
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(2000),
			}
			return d.DialContext(ctx, network, dns)
		},
	}
	ips, err := r.LookupIPAddr(context.Background(), name)
	log.Debugf("query %s on dns %s, the return ip is %#v; the error is [%s]", name, dns, ips, err)
	select {
	case <-ctx.Done():
		log.Debugf("query %s on dns %s, dns record already found, operation cancelled", name, dns)
		return
	case c <- dnsQueryResponse{
		ips,
		name,
		err,
	}:
		return
	}

}
