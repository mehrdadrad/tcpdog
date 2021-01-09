package ebpf

import (
	"testing"

	"github.com/mehrdadrad/tcpdog/config"
	"github.com/stretchr/testify/assert"
)

func TestGetBPFCode(t *testing.T) {
	cfgTracepoint := config.Tracepoint{
		Name:     "sock:inet_sock_set_state",
		Fields:   "custom_fields1",
		TCPState: "TCP_CLOSE",
		Inet:     []int{4, 6},
	}

	cfgFileds := map[string][]config.Field{
		"custom_fields1": {
			{Name: "SRTT", Math: "/1000", Filter: "SRTT>1000"},
			{Name: "TotalRetrans", Math: "", Filter: ""},
			{Name: "SAddr"},
			{Name: "DAddr"},
		},
	}

	source, err := GetBPFCode(&config.Config{
		Tracepoints: []config.Tracepoint{cfgTracepoint},
		Fields:      cfgFileds,
	})

	assert.NoError(t, err)

	assert.Contains(t, source, "struct tcp_sock *tcpi = tcp_sk(sk);")
	assert.Contains(t, source, "u32 srtt_us0;")
	assert.Contains(t, source, "u32 total_retrans1;")
	assert.Contains(t, source, "args->newstate != TCP_CLOSE")
	assert.Contains(t, source, "args->protocol != IPPROTO_TCP")

	// v4
	assert.Contains(t, source, "(data4.srtt_us0>1000)")
	assert.Contains(t, source, "ipv4_events0.perf_submit(args, &data4, sizeof(data4)")
	assert.Contains(t, source, "data4.srtt_us0 = (tcpi->srtt_us) /1000;")
	assert.Contains(t, source, "data4.total_retrans1 = (tcpi->total_retrans)")
	assert.Contains(t, source, "BPF_PERF_OUTPUT(ipv4_events0);")
	assert.Contains(t, source, "u32 skc_rcv_saddr2;")
	assert.Contains(t, source, "u32 skc_daddr3;")

	// v6
	assert.Contains(t, source, "(data6.srtt_us0>1000)")
	assert.Contains(t, source, "ipv6_events0.perf_submit(args, &data6, sizeof(data6)")
	assert.Contains(t, source, "data6.srtt_us0 = (tcpi->srtt_us) /1000;")
	assert.Contains(t, source, "data6.total_retrans1 = (tcpi->total_retrans)")
	assert.Contains(t, source, "BPF_PERF_OUTPUT(ipv6_events0);")
	assert.Contains(t, source, "unsigned __int128 skc_v6_rcv_saddr2;")
	assert.Contains(t, source, "unsigned __int128 skc_v6_daddr3;")
}
