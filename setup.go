package flatten

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"

	clog "github.com/coredns/coredns/plugin/pkg/log"
)

var log = clog.NewWithPlugin("flatten")

func init() {
	caddy.RegisterPlugin("flatten", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

func setup(c *caddy.Controller) error {
	f, err := parse(c)
	if err != nil {
		return plugin.Error("flatten", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		f.Next = next
		return f
	})

	return nil
}

func parse(c *caddy.Controller) (*Flatten, error) {
	f := &Flatten{}

	for c.Next() {
		args := c.RemainingArgs()
		if len(args) > 3 {
			return nil, c.ArgErr()
		}
		f.Original = args[0] + "."
		f.Target = args[1] + "."
		f.Upstream = args[2]
	}

	return f, nil
}
