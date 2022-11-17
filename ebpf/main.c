/* #pragma clang diagnostic push */
/* #pragma clang diagnostic ignored "-Waddress-of-packed-member" */
/* #pragma clang diagnostic ignored "-Warray-bounds" */
/* #pragma clang diagnostic ignored "-Wunused-label" */
/* #pragma clang diagnostic ignored "-Wgnu-variable-sized-type-not-at-end" */
/* #pragma clang diagnostic ignored "-Wframe-address" */
/* #pragma clang diagnostic ignored "-Wpass-failed" */

/* #include <uapi/linux/input-event-codes.h> */
#include <linux/input-event-codes.h>
/* #include <linux/input.h> */
/* #include <linux/atomic.h> */
/* #include <linux/atomic/atomic-long.h> */

#include "include/all.h"

struct input_dev {}; // fake

SEC("kprobe/input_handle_event")
int kprobe_input_handle_event(struct pt_regs *ctx)
/* int kprobe_input_handle_event(struct pt_regs *ctx, struct input_dev *dev, */
/*                               unsigned int type, unsigned int code, int value) */
{
    unsigned int type = (unsigned int)PT_REGS_PARM2(ctx);
    unsigned int code = (unsigned int)PT_REGS_PARM3(ctx);
    int value = (int)PT_REGS_PARM4(ctx);

    if (type == EV_KEY && value) { /* key down */
        bpf_printk("keydown code: %u\n", code);
        /* bpf_trace_printk("value %d\n", value); */
    } else if (type == EV_KEY && !value) { /* key up */
        bpf_printk("keyup code: %u\n", code);
    }

    return 0;
}

char _license[] SEC("license") = "GPL";
__u32 _version SEC("version") = 0xFFFFFFFE;
