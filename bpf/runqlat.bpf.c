#include "vmlinux.h"
#include <bpf/bpf_helpers.h>
#include <bpf/bpf_tracing.h>
#include <bpf/bpf_core_read.h>

char LICENSE[] SEC("license") = "Dual BSD/GPL";

#define MAX_SLOTS 64

struct {
    __uint(type, BPF_MAP_TYPE_HASH);
    __uint(max_entries, 131072);
    __type(key, __u64);   // MVP: tid in low32
    __type(value, __u64); // start timestamp (ns)
} start_ts SEC(".maps");

// per-cpu histogram (atomic 필요 없음)
struct {
    __uint(type, BPF_MAP_TYPE_PERCPU_ARRAY);
    __uint(max_entries, MAX_SLOTS);
    __type(key, __u32);
    __type(value, __u64);
} hist SEC(".maps");

// ✅ LLVM14 BPF 크래시 회피용: clz/div 없이 log2 bucket 계산
static __always_inline __u32 log2_bucket_u64(__u64 v)
{
    __u32 b = 0;
    // v가 0이면 bucket 0
    if (v == 0)
        return 0;

    // bounded + unrolled
#pragma unroll
    for (int i = 0; i < 63; i++) {
        if (v >> (i + 1))
            b = i + 1;
    }

    if (b >= MAX_SLOTS)
        b = MAX_SLOTS - 1;
    return b;
}

SEC("tracepoint/sched/sched_wakeup")
int runqlat_wakeup(struct trace_event_raw_sched_wakeup_template *ctx)
{
    __u64 now = bpf_ktime_get_ns();
    __u32 tid = BPF_CORE_READ(ctx, pid);

    __u64 key = (__u64)tid;
    bpf_map_update_elem(&start_ts, &key, &now, BPF_ANY);
    return 0;
}

SEC("tracepoint/sched/sched_switch")
int runqlat_switch(struct trace_event_raw_sched_switch *ctx)
{
    __u64 now = bpf_ktime_get_ns();
    __u32 next_tid = BPF_CORE_READ(ctx, next_pid);
    __u64 key = (__u64)next_tid;

    __u64 *tsp = bpf_map_lookup_elem(&start_ts, &key);
    if (!tsp)
        return 0;

    __u64 delta_ns = now - *tsp;

    // ✅ 여기서 us 변환(/1000) 하지 말고 ns로 바로 bucket
    __u32 b = log2_bucket_u64(delta_ns);

    __u64 *cnt = bpf_map_lookup_elem(&hist, &b);
    if (cnt)
        (*cnt)++;

    bpf_map_delete_elem(&start_ts, &key);
    return 0;
}