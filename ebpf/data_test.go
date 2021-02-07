package ebpf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateField(t *testing.T) {
	f, err := ValidateField("srtt")
	assert.NoError(t, err)
	assert.Equal(t, "SRTT", f)

	_, err = ValidateField("SRTT")
	assert.NoError(t, err)

	_, err = ValidateField("ttr")
	assert.Error(t, err)
}

func TestValidateTCPStatus(t *testing.T) {
	_, err := ValidateTCPStatus("TCP_ESTABLISHED")
	assert.NoError(t, err)
	_, err = ValidateTCPStatus("TCP_UNKNOWN")
	assert.Error(t, err)
}

func TestValidateTracepoint(t *testing.T) {
	assert.NoError(t, ValidateTracepoint("tcp:tcp_probe"))
	assert.Error(t, ValidateTracepoint("tcp:unknown"))
}
