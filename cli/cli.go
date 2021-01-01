package cli

import (
	"errors"

	"github.com/mehrdadrad/tcpdog/ebpf"
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

// Request represents command line request arguments.
type Request struct {
	Tracepoint string
	Fields     []string
	IPv4       bool
	IPv6       bool
	TCPState   string
	Output     string
}

var flags = []cli.Flag{
	&cli.BoolFlag{Name: "ipv4", Aliases: []string{"4"}, Usage: "enable IPv4 address", DefaultText: "true if IPv6 is false"},
	&cli.BoolFlag{Name: "ipv6", Aliases: []string{"6"}, Usage: "enable IPv6 address"},
	&cli.StringFlag{Name: "tracepoint", Aliases: []string{"tp"}, Value: "sock:inet_sock_set_state", Usage: "tracepoint name"},
	&cli.StringFlag{Name: "fields", Aliases: []string{"f"}, Value: "srtt,saddr,daddr,dport", Usage: "fields"},
	&cli.StringFlag{Name: "state", Aliases: []string{"s"}, Value: "tcp_close", Usage: "tcp state"},
}

// Get returns cli requested parameters.
func Get(args []string) (*Request, error) {
	var r = &Request{}

	app := &cli.App{
		Flags:  flags,
		Action: action(r),
	}

	err := app.Run(args)

	return r, err
}

func action(r *Request) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		r = &Request{
			Tracepoint: c.String("tp"),
			Fields:     c.StringSlice("fields"),
			IPv4:       c.Bool("4"),
			IPv6:       c.Bool("6"),
			TCPState:   c.String("state"),
		}
		return validate(r)
	}
}

func validate(r *Request) error {
	var err error

	// tracepoint
	err = validateTracepoint(r)
	if err != nil {
		return err
	}

	// fields
	r.Fields, err = validateFields(r)
	if err != nil {
		return err
	}

	return nil
}

func validateTracepoint(r *Request) error {
	if _, ok := validTracepoints[r.Tracepoint]; r.Tracepoint != "" && !ok {
		return errors.New("invalid tracepoint")
	}

	return nil
}

func validateFields(r *Request) ([]string, error) {
	fields := []string{}
	for _, f := range r.Fields {
		f, err := ebpf.ValidateField(f)
		if err != nil {
			return nil, err
		}
		fields = append(fields, f)
	}

	return fields, nil
}
