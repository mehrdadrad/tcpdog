package ebpf

import (
	"log"
	"strings"

	"github.com/mehrdadrad/tcpdog/config"
)

type CType uint8
type DType uint8

const (
	u8 CType = iota + 1
	u16
	u32
	u64
	u128
	char
)

const (
	IP DType = 1
)

// FieldAttrs represents
type FieldAttrs struct {
	CType  CType
	DType  DType
	CField string
	DS     string
	Func   string
	Filter string
	DSNP   bool
}

func (c CType) String() string {
	switch c {
	case u8:
		return "u8"
	case u16:
		return "u16"
	case u32:
		return "u32"
	case u64:
		return "u64"
	case u128:
		return "unsigned __int128"
	case char:
		return "char"
	}

	return "na"
}

func getReqFieldsV4(cfgFields []config.Field) []FieldAttrs {
	var reqFields []FieldAttrs

	for _, v := range cfgFields {
		if attrs, ok := fieldsModel4[v.Name]; ok {
			reqFields = append(reqFields, FieldAttrs{
				CField: attrs.CField,
				CType:  attrs.CType,
				DS:     attrs.DS,
				DSNP:   attrs.DSNP,
				Func:   getValue(v.Func, attrs.Func),
				Filter: replaceNameWithCFieldV4(getValue(v.Filter, attrs.Filter), v.Name, attrs.CField),
			})
		} else {
			log.Fatal("unknown field")
		}
	}

	return reqFields
}

func getReqFieldsV6(cfgFields []config.Field) []FieldAttrs {
	var reqFields []FieldAttrs

	for _, v := range cfgFields {
		if attrs, ok := fieldsModel6[v.Name]; ok {
			reqFields = append(reqFields, FieldAttrs{
				CField: attrs.CField,
				CType:  attrs.CType,
				DS:     attrs.DS,
				DSNP:   attrs.DSNP,
				Func:   getValue(v.Func, attrs.Func),
				Filter: replaceNameWithCFieldV6(getValue(v.Filter, attrs.Filter), v.Name, attrs.CField),
			})
		}
	}

	return reqFields
}

func getValue(v string, d string) string {
	if v != "" {
		return v
	}
	return d
}

func replaceNameWithCFieldV4(filter, name, cField string) string {
	return strings.Replace(filter, name, "data4."+cField, -1)
}

func replaceNameWithCFieldV6(filter, name, cField string) string {
	return strings.Replace(filter, name, "data6."+cField, -1)
}
