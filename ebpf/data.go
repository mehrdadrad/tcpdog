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
		"GSOSegs": {
			CType:  u16,
			CField: "gso_segs",
			DS:     "tcpi",
			Desc:   "Max number of segs per GSO packet",
		},
		"DataSegsIn": {
			CType:  u32,
			CField: "data_segs_in",
			DS:     "tcpi",
			Desc:   "Total number of data segments in",
		},
		"MaxWindow": {
			CType:  u32,
			CField: "max_window",
			DS:     "tcpi",
			Desc:   "Maximal window ever seen from peer",
		},
		"SndWnd": {
			CType:  u32,
			CField: "snd_wnd",
			DS:     "tcpi",
			Desc:   "The window we expect to receive",
		},
		"WindowClamp": {
			CType:  u32,
			CField: "window_clamp",
			DS:     "tcpi",
			Desc:   "Maximal window to advertise",
		},
		"RcvSSThresh": {
			CType:  u32,
			CField: "rcv_ssthresh",
			DS:     "tcpi",
			Desc:   "Current window clamp",
		},
		"ECNFlags": {
			CType:  u8,
			CField: "ecn_flags",
			DS:     "tcpi",
			Desc:   "ECN status bits",
		},
		"SndCwnd": {
			CType:  u32,
			CField: "snd_cwnd",
			DS:     "tcpi",
			Desc:   "Sending congestion window",
		},
		"PrrOut": {
			CType:  u32,
			CField: "prr_out",
			DS:     "tcpi",
			Desc:   "Total number of pkts sent during Recovery",
		},
		"Delivered": {
			CType:  u32,
			CField: "delivered",
			DS:     "tcpi",
			Desc:   "Total data packets delivered incl. rexmits",
		},
		"DeliveredCe": {
			CType:  u32,
			CField: "delivered_ce",
			DS:     "tcpi",
			Desc:   "Like the above but only ECE marked packets",
		},
		"Lost": {
			CType:  u32,
			CField: "lost",
			DS:     "tcpi",
			Desc:   "Total data packets lost incl. rexmits",
		},
		"LostOut": {
			CType:  u32,
			CField: "lost_out",
			DS:     "tcpi",
			Desc:   "Lost packets",
		},
		"PriorSSThresh": {
			CType:  u32,
			CField: "prior_ssthresh",
			DS:     "tcpi",
			Desc:   "ssthresh saved at recovery start",
		},
		"DataSegsOut": {
			CType:  u32,
			CField: "data_segs_out",
			DS:     "tcpi",
			Desc:   "Total number of data segments sent RFC4898",
		},
	}

	validTracepoints = map[string]bool{
		"tcp:tcp_retransmit_skb":    true,
		"tcp:tcp_retransmit_synack": true,
		"tcp:tcp_destroy_sock":      true,
		"tcp:tcp_send_reset":        true,
		"tcp:tcp_receive_reset":     true,
		"tcp:tcp_probe":             true,
		"sock:inet_sock_set_state":  true,
	}

	validTCPStatus = map[string]uint8{
		"TCP_ESTABLISHED":  1,
		"TCP_SYN_SENT":     2,
		"TCP_SYN_RECV":     3,
		"TCP_FIN_WAIT1":    4,
		"TCP_FIN_WAIT2":    5,
		"TCP_TIME_WAIT":    6,
		"TCP_CLOSE":        7,
		"TCP_CLOSE_WAIT":   8,
		"TCP_LAST_ACK":     9,
		"TCP_LISTEN":       10,
		"TCP_CLOSING":      11,
		"TCP_NEW_SYN_RECV": 12,
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

// ValidateField validates a field
func ValidateField(f string) (string, error) {
	if _, ok := fieldsModel4[f]; ok {
		return f, nil
	}

	if v, ok := fieldsLowerCaseMap[strings.ToLower(f)]; ok {
		return v, nil
	}

	return f, fmt.Errorf("invalid field: %s", f)
}

// ValidateTCPStatus validates a TCP status
func ValidateTCPStatus(status string) (string, error) {
	statusUpper := strings.ToUpper(status)
	if _, ok := validTCPStatus[statusUpper]; !ok {
		return statusUpper, fmt.Errorf("invalid TCP status: %s", status)
	}
	return statusUpper, nil
}

// ValidateTracepoint validates a tracepoint
func ValidateTracepoint(tp string) error {
	if _, ok := validTracepoints[tp]; !ok {
		return fmt.Errorf("invalid tracepoint: %s", tp)
	}
	return nil
}
