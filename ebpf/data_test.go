package ebpf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateField(t *testing.T) {
	f, err := ValidateField("srtt")
	assert.NoError(t, err)
	assert.Equal(t, "SRTT", f)

	_, err = ValidateField("ttr")
	assert.Error(t, err)
}
