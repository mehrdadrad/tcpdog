package ebpf

import (
	"bytes"
	"errors"
	"strings"
	"text/template"

	"github.com/mehrdadrad/tcpdog/config"
)

// TracepointTemplate represents template's data object
type TracepointTemplate struct {
	Fields4    []FieldAttrs
	Fields6    []FieldAttrs
	Tracepoint string
	TCPState   string
	Suffix     int
	Sample     int
	TCPInfo    bool
	ICSK       bool
}

// Init intializes tracepointTemplate
func (t *TracepointTemplate) Init() {
	for _, f := range append(t.Fields4, t.Fields6...) {
		if strings.Contains(f.DS, "tcpi") {
			t.TCPInfo = true
		}
		if strings.Contains(f.DS, "icsk") {
			t.ICSK = true
		}
	}
}

// CGen represents code generator
type CGen struct {
	conf *config.Config
}

// GetBPFCode returns BPF program
func GetBPFCode(conf *config.Config) (string, error) {
	var bpfCode string

	cg := CGen{conf: conf}
	for index, tracepoint := range conf.Tracepoints {
		code, err := cg.getTracepointBPFCode(index, tracepoint)
		if err != nil {
			return "", err
		}

		bpfCode += code
	}

	return includes + bpfCode, nil
}

func (c *CGen) getTracepointBPFCode(index int, tp config.Tracepoint) (string, error) {
	var (
		cfgFields []config.Field
		fields4   []FieldAttrs
		fields6   []FieldAttrs
		ok        bool
	)

	if cfgFields, ok = c.conf.Fields[tp.Fields]; !ok {
		return "", errors.New("field's template not exist")
	}

	for _, v := range tp.INet {
		if v == 4 {
			fields4 = getReqFieldsV4(cfgFields)
		} else {
			fields6 = getReqFieldsV6(cfgFields)
		}
	}

	tp.Name = strings.Replace(tp.Name, ":", "__", 1)

	tt := TracepointTemplate{
		Fields4:    fields4,
		Fields6:    fields6,
		Tracepoint: tp.Name,
		TCPState:   tp.TCPState,
		Suffix:     index,
		Sample:     tp.Sample,
	}

	tt.Init()

	buf := new(bytes.Buffer)
	tmpl, err := template.New("source").Funcs(funcMap).Parse(source)
	if err != nil {
		return "", err
	}

	if err = tmpl.Execute(buf, tt); err != nil {
		return "", err
	}

	return buf.String(), nil
}
