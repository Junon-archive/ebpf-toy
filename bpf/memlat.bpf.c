#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

char LICENSE[] SEC("license") = "Dual BSD/GPL";

#define MAX_SLOTS 64

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 8192);
    __type(key, __u64);    // pid_tgid
    __type(value, __u64);  // start ts (ns)
} start_ts SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, MAX_SLOTS);
    __type(key, __u32);    // bucket
    __type(value, __u64);  // count
} hist SEC(".maps");

static __always_inline __u32 log2_bucket_u64(__u64 v)
{
    // v must be > 0
    __u32 b = 0;
    // simple log2; verifier-friendly
    while (v >>= 1) {
        b++;
        if (b >= (MAX_SLOTS - 1)) break;
    }
    return b;
}

// Entry: handle_mm_fault()
SEC("kprobe/handle_mm_fault")
int BPF_KPROBE(memlat_entry)
{
    __u64 key = bpf_get_current_pid_tgid();
    __u64 ts = bpf_ktime_get_ns();
    bpf_map_update_elem(&start_ts, &key, &ts, BPF_ANY);
    return 0;
}

// Exit: handle_mm_fault() return
SEC("kretprobe/handle_mm_fault")
int BPF_KRETPROBE(memlat_exit)
{
    __u64 key = bpf_get_current_pid_tgid();
    __u64 *tsp = bpf_map_lookup_elem(&start_ts, &key);
    if (!tsp)
        return 0;

    __u64 delta_ns = bpf_ktime_get_ns() - *tsp;
    bpf_map_delete_elem(&start_ts, &key);

    // convert to us; avoid zero
    __u64 delta_us = delta_ns / 1000;
    if (delta_us == 0)
        delta_us = 1;

    __u32 b = log2_bucket_u64(delta_us);
    __u64 *cnt = bpf_map_lookup_elem(&hist, &b);
    if (cnt)
        __sync_fetch_and_add(cnt, 1);

    return 0;
}
