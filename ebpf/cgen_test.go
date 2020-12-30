package ebpf

import (
	"log"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/mehrdadrad/tcpdog/config"
)

func TestX(t *testing.T) {
	var funcMap = template.FuncMap{
		"isBPF": strings.HasPrefix,
	}

	cfgTracepoint := config.Tracepoint{
		Name:            "tcp:tcp_retransmit_skb",
		Fields:          "custom_fields1",
		TCPState:        "close",
		PollingInterval: 100,
		Inet:            []int{4, 6},
		Geo:             "enable",
	}

	cfgFileds := map[string][]config.Field{
		"custom_fields1": {
			{Name: "SRTT", Func: "/1000", Filter: ">1000"},
			{Name: "TotalRetrans", Func: "", Filter: ""},
		},
	}

	fields4 := getReqFieldsV4(cfgFileds[cfgTracepoint.Fields])
	log.Printf("%#v", fields4)

	tt := TracepointTemplate{
		Fields4:    fields4,
		Tracepoint: cfgTracepoint.Name,
		TCPState:   cfgTracepoint.TCPState,
		Suffix:     1,
	}
	tt.Init()
	tmpl, err := template.New("source").Funcs(funcMap).Parse(source)
	if err != nil {
		log.Fatal(err)
	}
	tmpl.Execute(os.Stdout, tt)
}
