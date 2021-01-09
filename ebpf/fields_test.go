package ebpf

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mehrdadrad/tcpdog/config"
)

func TestGetReqFields(t *testing.T) {
	arg := []config.Field{
		{Name: "SRTT", Math: "/1000", Filter: ">1000"},
		{Name: "DPort", Math: "", Filter: ""},
	}

	fields := getReqFieldsV4(arg)

	assert.Len(t, fields, 2)
	assert.Equal(t, "srtt_us", fields[0].CField)
	assert.Equal(t, "/1000", fields[0].UMath)
	assert.Equal(t, ">1000", fields[0].Filter)
}
