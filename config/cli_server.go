package config

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	cli "github.com/urfave/cli/v2"
)

var flagsServer = []cli.Flag{
	&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Value: "", Usage: "path to a file in yaml format to read configuration"},
}

// Get returns server cli request
func getServer(args []string, version string) (*serverCLIRequest, error) {
	var r = &serverCLIRequest{}

	initCLIServer()

	app := &cli.App{
		Version: version,
		Flags:   flagsServer,
		Action:  actionServer(r),
	}

	err := app.Run(args)

	return r, err
}

func actionServer(r *serverCLIRequest) cli.ActionFunc {
	return func(c *cli.Context) error {
		r.Config = c.String("config")

		return nil
	}
}

func initCLIServer() {
	cli.AppHelpTemplate = `usage: tcpdog server options
	
options:

   {{range .VisibleFlags}}{{.}}
   {{end}}
`

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("TCPDog version: %s [server]\n", c.App.Version)
		cli.OsExiter(0)
	}

	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		funcMap := template.FuncMap{
			"join": strings.Join,
		}
		t := template.Must(template.New("help").Funcs(funcMap).Parse(templ))
		t.Execute(w, data)
		cli.OsExiter(0)
	}
}
