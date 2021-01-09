package ebpf

import (
	"fmt"
	"strings"

	"github.com/mehrdadrad/tcpdog/config"
)

// CType represents clang type
type CType uint8

// DType represents data type
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
	// IP represents IP data type
	IP DType = 1
)

// FieldAttrs represents
type FieldAttrs struct {
	CType     CType
	DType     DType
	CField    string
	DS        string
	UMath     string
	Math      string
	Func      string
	Filter    string
	Desc      string
	DSNP      bool
	BigEndian bool
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

	for i, v := range cfgFields {
		attrs := fieldsModel4[v.Name]
		reqFields = append(reqFields, FieldAttrs{
			CField: attrs.CField,
			CType:  attrs.CType,
			DS:     attrs.DS,
			DSNP:   attrs.DSNP,
			UMath:  v.Math,
			Math:   attrs.Math,
			Func:   attrs.Func,
			Filter: replaceNameWithCFieldV4(getValue(v.Filter, attrs.Filter), v.Name, attrs.CField, i),
		})
	}

	return reqFields
}

func getReqFieldsV6(cfgFields []config.Field) []FieldAttrs {
	var reqFields []FieldAttrs

	for i, v := range cfgFields {
		attrs := fieldsModel6[v.Name]
		reqFields = append(reqFields, FieldAttrs{
			CField: attrs.CField,
			CType:  attrs.CType,
			DS:     attrs.DS,
			DSNP:   attrs.DSNP,
			UMath:  v.Math,
			Math:   attrs.Math,
			Func:   attrs.Func,
			Filter: replaceNameWithCFieldV6(getValue(v.Filter, attrs.Filter), v.Name, attrs.CField, i),
		})
	}

	return reqFields
}

func getValue(v string, d string) string {
	if v != "" {
		return v
	}
	return d
}

func replaceNameWithCFieldV4(filter, name, cField string, index int) string {
	return strings.Replace(filter, name, fmt.Sprintf("data4.%s%d", cField, index), -1)
}

func replaceNameWithCFieldV6(filter, name, cField string, index int) string {
	return strings.Replace(filter, name, fmt.Sprintf("data6.%s%d", cField, index), -1)
}
