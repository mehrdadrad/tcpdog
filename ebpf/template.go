package ebpf

import (
	"strings"
	"text/template"
)

const includes = `
#include <uapi/linux/ptrace.h>
#include <net/sock.h>
#include <bcc/proto.h>
#include <linux/tcp.h>
`

var funcMap = template.FuncMap{
	"isBPF": strings.HasPrefix,
}

const source = `
	{{if .Fields4}}

	{{if ne .Sample 0}}
	BPF_HASH(ipv4_sample, struct sock *, u64, 100000);
	{{- end}}

	struct ipv4_data{{.Suffix}}_t {
		{{- range $value := .Fields4}}
		{{if eq $value.CField "current_comm"}}
		{{printf "%s %s[TASK_COMM_LEN];" $value.CType $value.CField}}
		{{- else}}
		{{printf "%s %s;" $value.CType $value.CField}} 
		{{- end}}
		{{- end}}
	};
	BPF_PERF_OUTPUT(ipv4_events{{.Suffix}});
	{{end}}

	{{if .Fields6}}

	{{if ne .Sample 0}}
	BPF_HASH(ipv6_sample, struct sock *, u64, 100000);
	{{- end}}

	struct ipv6_data{{.Suffix}}_t {
		{{- range $value := .Fields6}}
		{{if eq $value.CField "current_comm"}}
		{{printf "%s %s[TASK_COMM_LEN];" $value.CType $value.CField}}
		{{- else}}
		{{printf "%s %s;" $value.CType $value.CField}} 
		{{- end}}
		{{- end}}
	};
	BPF_PERF_OUTPUT(ipv6_events{{.Suffix}});
	{{end}}

	
	int sk_trace{{.Suffix}}(struct tracepoint__{{.Tracepoint}}* args)
	{
		{{if eq .Tracepoint "sock__inet_sock_set_state"}}
		if (args->protocol != IPPROTO_TCP)
			return 0;

		if (args->newstate != {{.TCPState}}) {
			return 0;
		}
		{{end}}

		struct sock *sk = (struct sock *)args->skaddr;

		{{if .TCPInfo}}
		struct tcp_sock *tcpi = (struct tcp_sock *)sk;
		{{end}}

		u16 family = sk->__sk_common.skc_family;

		{{if .Fields4}}
		struct ipv4_data{{.Suffix}}_t data4 = {};
			
		if (family == AF_INET) {
			{{- range $value := .Fields4}}
			{{if not (isBPF $value.DS "bpf_") }}
			{{if $value.DSNP}}
			{{printf "data4.%s = %s.%s%s;" $value.CField $value.DS $value.CField $value.Func}}
			{{else}}
			{{printf "data4.%s = %s->%s%s;" $value.CField $value.DS $value.CField $value.Func}}
			{{- end}}
			{{- end}}
			{{- end}}

			{{- range $value := .Fields4}}
			{{if eq $value.DS "bpf_get_current_comm"}}	
			{{printf "bpf_get_current_comm(&data4.%s,sizeof(data4.%s));" $value.CField $value.CField}}
			{{- end}}
			{{if eq $value.DS "bpf_get_current_pid_tgid"}}	
			{{printf "data4.%s = bpf_get_current_pid_tgid() >> 32;" $value.CField}}
			{{- end}}
			{{- end}}

			{{- range $value := .Fields4}}
			{{if $value.Filter}}
			{{printf "if (%s) {" $value.Filter}}
				return 0;
			}
			{{- end}}
			{{- end}}

			{{if ne .Sample 0}}
			u64 *count;
			u64 zero = 0;
			count = ipv4_sample.lookup_or_try_init(&sk, &zero);
			if (!count) {
				bpf_probe_read_kernel(&count, sizeof(count), &zero);
				ipv4_sample.increment(sk);
				return 0;
			}

			if (*count < {{.Sample}}) {
				ipv4_sample.increment(sk);
				return 0;
			}
			
			ipv4_sample.delete(&sk);	
			
			{{- end}}

			ipv4_events{{.Suffix}}.perf_submit(args, &data4, sizeof(data4));

			return 0;
		}
		{{end}}

		{{if .Fields6}}
		struct ipv6_data{{.Suffix}}_t data6 = {};

		if (family == AF_INET6) {
			{{- range $value := .Fields6}}
			{{if and (not (isBPF $value.DS "bpf_")) (not (eq $value.CField "skc_v6_daddr")) (not (eq $value.CField "skc_v6_rcv_saddr"))}}
			
			{{if $value.DSNP}}
			{{printf "data6.%s = %s.%s%s;" $value.CField $value.DS $value.CField $value.Func}}
			{{else}}
			{{printf "data6.%s = %s->%s%s;" $value.CField $value.DS $value.CField $value.Func}}
			{{- end}}
			
			{{else}}

			{{if or (eq $value.CField "skc_v6_daddr") (eq $value.CField "skc_v6_rcv_saddr")}}
			{{printf "bpf_probe_read(&data6.%s, sizeof(data6.%s)," $value.CField $value.CField}} 
			{{printf "	%s.%s.in6_u.u6_addr32);" $value.DS $value.CField}}
			{{- end}}

			{{- end}}
			{{- end}}

			

			{{- range $value := .Fields6}}
			{{if eq $value.DS "bpf_get_current_comm"}}	
			{{printf "bpf_get_current_comm(&data6.%s,sizeof(data6.%s));" $value.CField $value.CField}}
			{{- end}}
			{{if eq $value.DS "bpf_get_current_pid_tgid"}}	
			{{printf "data6.%s = bpf_get_current_pid_tgid() >> 32;" $value.CField}}
			{{- end}}
			{{- end}}

			{{- range $value := .Fields6}}
			{{if $value.Filter}}
			{{printf "if (%s) {" $value.Filter}}
				return 0;
			}
			{{- end}}
			{{- end}}

			{{if ne .Sample 0}}
			u64 *count;
			u64 zero = 0;
			count = ipv6_sample.lookup_or_try_init(&sk, &zero);
			if (!count) {
				bpf_probe_read_kernel(&count, sizeof(count), &zero);
				ipv6_sample.increment(sk);
				return 0;
			}

			if (*count < {{.Sample}}) {
				ipv6_sample.increment(sk);
				return 0;
			}
			
			ipv6_sample.delete(&sk);	
			
			{{- end}}

			ipv6_events{{.Suffix}}.perf_submit(args, &data6, sizeof(data6));

			return 0;	
		}
		{{end}}

		return 0;
	}	
`
