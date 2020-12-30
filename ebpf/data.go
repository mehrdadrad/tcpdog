package ebpf

var fieldsModel4 = map[string]FieldAttrs{
	"TCPHeaderLen": {
		CType:  u16,
		CField: "tcp_header_len",
		DS:     "tcpi",
		DSNP:   false,
	},
	"SRTT": {
		DS:     "tcpi",
		CField: "srtt_us",
		CType:  u32,
	},
	"DPort": {
		DS:     "sk->__sk_common",
		DSNP:   true,
		CField: "skc_dport",
		CType:  u16,
	},
	"TotalRetrans": {
		DS:     "tcpi",
		CField: "total_retrans",
		CType:  u32,
	},
	"AdvMSS": {
		DS:     "tcpi",
		CField: "advmss",
		CType:  u16,
	},
	"SAddr": {
		DS:     "sk->__sk_common",
		DSNP:   true,
		CField: "skc_rcv_saddr",
		CType:  u32,
		DType:  IP,
	},
	"DAddr": {
		DS:     "sk->__sk_common",
		DSNP:   true,
		CField: "skc_daddr",
		CType:  u32,
		DType:  IP,
	},
	"BytesReceived": {
		DS:     "tcpi",
		CField: "bytes_received",
		CType:  u64,
	},
	"BytesSent": {
		DS:     "tcpi",
		CField: "bytes_sent",
		CType:  u64,
	},
	"BytesAcked": {
		DS:     "tcpi",
		CField: "BytesAcked",
		CType:  u64,
	},
	"Task": {
		DS:     "bpf_get_current_comm",
		CField: "current_comm",
		CType:  char,
	},
	"PID": {
		DS:     "bpf_get_current_pid_tgid",
		CField: "pid",
		CType:  u32,
	},
	"SegsIn": {
		DS:     "tcpi",
		CField: "segs_in",
		CType:  u32,
	},
	"SegsOut": {
		DS:     "tcpi",
		CField: "segs_out",
		CType:  u32,
	},
	"DsackDups": {
		DS:     "tcpi",
		CField: "dsack_dups",
		CType:  u32,
	},
}

var fieldsModel6 = map[string]FieldAttrs{}

func init() {
	for k, v := range fieldsModel4 {
		if k == "SAddr" {
			v.CType = u128
			v.CField = "skc_v6_rcv_saddr"
		}
		if k == "DAddr" {
			v.CType = u128
			v.CField = "skc_v6_daddr"
		}
		fieldsModel6[k] = v
	}
}
