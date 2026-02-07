#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>

char LICENSE[] SEC("license") = "Dual BSD/GPL";

/*
 * Minimal "hello" eBPF for 0-week milestone:
 * Attach to tracepoint sched:sched_switch and increment a counter.
 */

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, 1);
    __type(key, __u32);
    __type(value, __u64);
} counter SEC(".maps");

SEC("tracepoint/sched/sched_switch")
int tp_sched_switch(struct trace_event_raw_sched_switch *ctx)
{
    __u32 key = 0;
    __u64 *val = bpf_map_lookup_elem(&counter, &key);
    if (val) {
        __sync_fetch_and_add(val, 1);
    }
    return 0;
}
