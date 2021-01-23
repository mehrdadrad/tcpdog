package cli

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/mehrdadrad/tcpdog/server/config"
	cli "github.com/urfave/cli/v2"
)

var flags = []cli.Flag{
	&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Value: "", Usage: "path to a file in yaml format to read configuration"},
}

// Get returns cli request
func Get(args []string, version string) (*config.CLIRequest, error) {
	var r = &config.CLIRequest{}

	app := &cli.App{
		Version: version,
		Flags:   flags,
		Action:  action(r),
	}

	err := app.Run(args)

	return r, err
}

func action(r *config.CLIRequest) cli.ActionFunc {
	return func(c *cli.Context) error {
		r.Config = c.String("config")

		return nil
	}
}

func init() {
	cli.AppHelpTemplate = `usage: tcpdog server options
	
options:

   {{range .VisibleFlags}}{{.}}
   {{end}}
`

	cli.VersionFlag = &cli.BoolFlag{
		Name: "version", Aliases: []string{"v"},
		Usage: "print only the version",
	}

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
