package flatten

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/miekg/dns"
)

// Flatten is a plugin that flattens CNAME chains in responses.
type Flatten struct {
	Next     plugin.Handler
	Original string // Original domain name to flatten
	Target   string // Target CNAME to flatten to
	Upstream string // Upstream DNS server to resolve the target CNAME
}

// ServeDNS implements the plugin.Handler interface.
func (f Flatten) ServeDNS(ctx context.Context, w dns.ResponseWriter, r *dns.Msg) (int, error) {
	state := request.Request{W: w, Req: r}

	// Check if the query is for the original domain name and an A or AAAA record
	if state.QName() != f.Original || (state.QType() != dns.TypeA && state.QType() != dns.TypeAAAA) {
		return plugin.NextOrFailure(f.Name(), f.Next, ctx, w, r)
	}

	// Create a new response message
	a := new(dns.Msg)
	a.SetReply(r)
	a.Authoritative = true

	// Resolve the target CNAME A records
	if err := f.resolveAorAAAA(f.Target, dns.TypeA, a, ctx); err != nil {
		return dns.RcodeServerFailure, err
	}

	// Resolve the target CNAME AAAA records
	if err := f.resolveAorAAAA(f.Target, dns.TypeAAAA, a, ctx); err != nil {
		return dns.RcodeServerFailure, err
	}

	// Rewrite the RR headers to match the original domain name
	for _, rr := range a.Answer {
		rr.Header().Name = f.Original
	}

	// Write the response
	w.WriteMsg(a)

	// Log the response
	log.Infof("%s:%s - [%s] flattened to [%s] via %s", state.IP(), state.Port(), state.Name(), f.Target, f.Upstream)

	return dns.RcodeSuccess, nil
}

func (f Flatten) resolveAorAAAA(name string, qtype uint16, msg *dns.Msg, ctx context.Context) error {
	m := new(dns.Msg)
	m.SetQuestion(name, qtype)

	c := new(dns.Client)
	in, _, err := c.Exchange(m, f.Upstream)
	if err != nil {
		return err
	}

	msg.Answer = append(msg.Answer, in.Answer...)

	return nil
}

// Name implements the Handler interface.
func (f Flatten) Name() string { return "flatten" }

// Compile-time assertion
var _ plugin.Handler = Flatten{}
