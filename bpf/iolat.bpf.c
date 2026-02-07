// bpf/iolat.bpf.c
#include "vmlinux.h"

#include <bpf/bpf_helpers.h>
#include <bpf/bpf_core_read.h>
#include <bpf/bpf_tracing.h>

char LICENSE[] SEC("license") = "Dual BSD/GPL";

#define MAX_SLOTS 26

struct iolat_start {
    __u64 ts;
    __u64 pid_tgid;
    char comm[16];
};

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 131072);
    __type(key, __u64);               // rq_id (casted pointer)
    __type(value, struct iolat_start);
} start_ts SEC(".maps");

struct {
    __uint(type, BPF_MAP_TYPE_ARRAY);
    __uint(max_entries, MAX_SLOTS);
    __type(key, __u32);
    __type(value, __u64);
} hist SEC(".maps");

// Workaround: LLVM-14 BPF backend can crash with __builtin_clzll()
// So compute floor(log2(v)) via an unrolled loop.
static __always_inline __u32 log2_bucket(__u64 v)
{
    if (v == 0)
        return 0;

    __u32 b = 0;

#pragma clang loop unroll(full)
    for (int i = 0; i < 63; i++) {
        if (v > 1) {
            v >>= 1;
            b++;
        } else {
            break;
        }
    }

    if (b >= MAX_SLOTS)
        b = MAX_SLOTS - 1;

    return b;
}

SEC("tp_btf/block_rq_issue")
int BPF_PROG(iolat_issue, struct request *rq)
{
    __u64 rq_id = (__u64)rq;

    struct iolat_start info = {};
    info.ts = bpf_ktime_get_ns();
    info.pid_tgid = bpf_get_current_pid_tgid();
    bpf_get_current_comm(&info.comm, sizeof(info.comm));

    bpf_map_update_elem(&start_ts, &rq_id, &info, BPF_ANY);
    return 0;
}

SEC("tp_btf/block_rq_complete")
int BPF_PROG(iolat_complete, struct request *rq, blk_status_t error, unsigned int nr_bytes)
{
    __u64 rq_id = (__u64)rq;

    struct iolat_start *info = bpf_map_lookup_elem(&start_ts, &rq_id);
    if (!info)
        return 0;

    __u64 now = bpf_ktime_get_ns();
    __u64 delta_ns = now - info->ts;
    __u64 delta_us = delta_ns / 1000;

    __u32 b = log2_bucket(delta_us);

    __u64 *cnt = bpf_map_lookup_elem(&hist, &b);
    if (cnt)
        __sync_fetch_and_add(cnt, 1);

    bpf_map_delete_elem(&start_ts, &rq_id);
    return 0;
}
