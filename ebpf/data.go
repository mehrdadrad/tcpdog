package ebpf

import (
	"fmt"
	"strings"
)

var (
	fieldsLowerCaseMap = map[string]string{}
	fieldsModel6       = map[string]FieldAttrs{}
	fieldsModel4       = map[string]FieldAttrs{
		"TCPHeaderLen": {
			CType:  u16,
			CField: "tcp_header_len",
			DS:     "tcpi",
			Desc:   "Bytes of tcp header to send",
		},
		"SRTT": {
			DS:     "tcpi",
			CField: "srtt_us",
			CType:  u32,
			Desc:   "RTT measurement: smoothed round trip time << 3 in usecs",
		},
		"MDev": {
			DS:     "tcpi",
			CField: "mdev_us",
			CType:  u32,
			Desc:   "RTT measurement: medium deviation",
		},
		"MDevMax": {
			DS:     "tcpi",
			CField: "mdev_max_us",
			CType:  u32,
			Desc:   "RTT measurement: maximal mdev for the last rtt period",
		},
		"RTTVar": {
			DS:     "tcpi",
			CField: "rttvar_us",
			CType:  u32,
			Desc:   "RTT measurement: smoothed mdev_max",
		},
		"TotalRetrans": {
			DS:     "tcpi",
			CField: "total_retrans",
			CType:  u32,
			Desc:   "Total retransmits for entire connectio",
		},
		"AdvMSS": {
			DS:     "tcpi",
			CField: "advmss",
			CType:  u16,
			Desc:   "Advertised MSS",
		},
		"SAddr": {
			DS:        "sk->__sk_common",
			DSNP:      true,
			BigEndian: true,
			CField:    "skc_rcv_saddr",
			CType:     u32,
			DType:     IP,
			Desc:      "",
		},
		"DAddr": {
			DS:        "sk->__sk_common",
			DSNP:      true,
			BigEndian: true,
			CField:    "skc_daddr",
			CType:     u32,
			DType:     IP,
			Desc:      "",
		},
		"DPort": {
			DS:        "sk->__sk_common",
			DSNP:      true,
			BigEndian: true,
			CField:    "skc_dport",
			CType:     u16,
			Desc:      "",
		},
		"LPort": {
			DS:        "sk->__sk_common",
			DSNP:      true,
			BigEndian: true,
			CField:    "skc_num",
			CType:     u16,
			Desc:      "",
		},
		"BytesReceived": {
			DS:     "tcpi",
			CField: "bytes_received",
			CType:  u64,
			Desc:   "RFC4898 tcpEStatsAppHCThruOctetsReceived",
		},
		"BytesSent": {
			DS:     "tcpi",
			CField: "bytes_sent",
			CType:  u64,
			Desc:   "RFC4898 tcpEStatsPerfHCDataOctetsOut",
		},
		"BytesAcked": {
			DS:     "tcpi",
			CField: "bytes_acked",
			CType:  u64,
			Desc:   "RFC4898 tcpEStatsAppHCThruOctetsAcked",
		},
		"NumSAcks": {
			DS:     "tcpi->rx_opt",
			CField: "num_sacks",
			CType:  u8,
			DSNP:   true,
			Desc:   "Number of SACK blocks",
		},
		"UserMSS": {
			DS:     "tcpi->rx_opt",
			CField: "user_mss",
			CType:  u16,
			DSNP:   true,
			Desc:   "MSS requested by user in ioctl",
		},
		"RTT": {
			DS:     "tcpi->rack",
			CField: "rtt_us",
			CType:  u32,
			DSNP:   true,
			Desc:   "Associated RTT",
		},
		"MSSClamp": {
			DS:     "tcpi->rx_opt",
			CField: "mss_clamp",
			CType:  u16,
			DSNP:   true,
			Desc:   "Maximal mss, negotiated at connection setup",
		},
		"Task": {
			DS:     "bpf_get_current_comm",
			CField: "current_comm",
			CType:  char,
			Desc:   "",
		},
		"PID": {
			DS:     "bpf_get_current_pid_tgid",
			CField: "pid",
			CType:  u32,
			Desc:   "",
		},
		"SegsIn": {
			DS:     "tcpi",
			CField: "segs_in",
			CType:  u32,
			Desc:   "Total number of segments in",
		},
		"SegsOut": {
			DS:     "tcpi",
			CField: "segs_out",
			CType:  u32,
			Desc:   "Total number of segments sent",
		},
		"DsackDups": {
			DS:     "tcpi",
			CField: "dsack_dups",
			CType:  u32,
			Desc:   "Total number of DSACK blocks received",
		},
		"RateDelivered": {
			DS:     "tcpi",
			CField: "rate_delivered",
			CType:  u32,
			Desc:   "Saved rate sample: packets delivered",
		},
		"RateInterval": {
			DS:     "tcpi",
			CField: "rate_interval_us",
			CType:  u32,
			Desc:   "Saved rate sample: time elapsed",
		},
		"SndSSThresh": {
			DS:     "tcpi",
			CField: "snd_ssthresh",
			CType:  u32,
			Desc:   "Slow start size threshold",
		},
	}
)

func init() {
	for k, v := range fieldsModel4 {
		fieldsLowerCaseMap[strings.ToLower(k)] = k

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

// ValidateField checks if field exist
func ValidateField(f string) (string, error) {
	if _, ok := fieldsModel4[f]; ok {
		return f, nil
	}

	if v, ok := fieldsLowerCaseMap[strings.ToLower(f)]; ok {
		return v, nil
	}

	return f, fmt.Errorf("%s, invalid field", f)
}
