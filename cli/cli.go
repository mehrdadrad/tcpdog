package cli

import (
	"strings"

	"github.com/mehrdadrad/tcpdog/config"
	cli "github.com/urfave/cli/v2"
)

var validTracepoints = map[string]bool{
	"tcp:tcp_retransmit_skb":    true,
	"tcp:tcp_retransmit_synack": true,
	"tcp:tcp_destroy_sock":      true,
	"tcp:tcp_send_reset":        true,
	"tcp:tcp_receive_reset":     true,
	"tcp:tcp_probe":             true,
	"sock:inet_sock_set_state":  true,
}

var flags = []cli.Flag{
	&cli.BoolFlag{Name: "ipv4", Aliases: []string{"4"}, Usage: "enable IPv4 address", DefaultText: "true if IPv6 is false"},
	&cli.BoolFlag{Name: "ipv6", Aliases: []string{"6"}, Usage: "enable IPv6 address"},
	&cli.StringFlag{Name: "tracepoint", Aliases: []string{"tp"}, Value: "sock:inet_sock_set_state", Usage: "tracepoint name"},
	&cli.StringFlag{Name: "fields", Aliases: []string{"f"}, Value: "srtt,saddr,daddr,dport", Usage: "fields"},
	&cli.StringFlag{Name: "state", Aliases: []string{"s"}, Value: "TCP_CLOSE", Usage: "tcp state"},
	&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Value: "", Usage: "path to a file in yaml format to read configuration"},
}

// Get returns cli config.CLIRequested parameters.
func Get(args []string) (*config.CLIRequest, error) {
	var r = &config.CLIRequest{}

	app := &cli.App{
		Flags:  flags,
		Action: action(r),
	}

	err := app.Run(args)

	return r, err
}

func action(r *config.CLIRequest) cli.ActionFunc {
	return func(c *cli.Context) error {

		r.Tracepoint = c.String("tp")
		r.Fields = strings.Split(c.String("fields"), ",")
		r.IPv4 = c.Bool("4")
		r.IPv6 = c.Bool("6")
		r.TCPState = c.String("state")
		r.Config = c.String("config")

		return nil
	}
}
