#include <linux/input-event-codes.h>

#include "include/all.h"

struct konamicode_status {
    u64 last_press;
    int completion;
};

struct bpf_map_def SEC("maps/konamicode_sequence") konamicode_sequence = {
    .type = BPF_MAP_TYPE_ARRAY,
    .key_size = sizeof(int),
    .value_size = sizeof(struct konamicode_status),
    .max_entries = 1,
};

int __attribute__((always_inline)) validate_konamicode_input(struct konamicode_status* ks) {
    ks->completion++;
    int key = 0;
    return bpf_map_update_elem(&konamicode_sequence, &key, ks, BPF_ANY);
}

int __attribute__((always_inline)) reset_konamicode(struct konamicode_status* ks) {
    bpf_printk("konamicode reset\n");
    ks->completion = 0;
    int key = 0;
    return bpf_map_update_elem(&konamicode_sequence, &key, ks, BPF_ANY);
}

struct bpf_map_def SEC("maps/konamicode_activation_counter") konamicode_activation_counter = {
    .type = BPF_MAP_TYPE_ARRAY,
    .key_size = sizeof(int),
    .value_size = sizeof(int),
    .max_entries = 1,
};

int __attribute__((always_inline)) activate_konamicode() {
    int key = 0;
    int* counter = bpf_map_lookup_elem(&konamicode_activation_counter, &key);
    if (counter == NULL) {
        return -1;
    }
    *counter += 1;
    bpf_printk("konamicode counter: %i\n", *counter);
    return bpf_map_update_elem(&konamicode_activation_counter, &key, counter, BPF_ANY);
}


SEC("kprobe/input_handle_event")
int kprobe_input_handle_event(struct pt_regs *ctx)
{
    unsigned int type = (unsigned int)PT_REGS_PARM2(ctx);
    unsigned int code = (unsigned int)PT_REGS_PARM3(ctx);
    int value = (int)PT_REGS_PARM4(ctx);

    if (type == EV_KEY && value) { /* key down */
        int key = 0;
        struct konamicode_status *ks = bpf_map_lookup_elem(&konamicode_sequence, &key);
        if (ks == NULL) {
            bpf_printk("NULL\n");
            return 0;
        }

        unsigned int completion = ks->completion;
        switch (completion) {
        case 0:
            if (code == KEY_UP) {
                bpf_printk("UP 1\n");
                return (validate_konamicode_input(ks));
            }
            break;
        case 1:
            if (code == KEY_UP) {
                bpf_printk("UP 2\n");
                return (validate_konamicode_input(ks));
            }
            break;
        case 2:
            if (code == KEY_DOWN) {
                bpf_printk("DOWN 1\n");
                return (validate_konamicode_input(ks));
            }
            break;
        case 3:
            if (code == KEY_DOWN) {
                bpf_printk("DOWN 2\n");
                return (validate_konamicode_input(ks));
            }
            break;
        case 4:
            if (code == KEY_LEFT) {
                bpf_printk("LEFT 1\n");
                return (validate_konamicode_input(ks));
            }
        case 5:
            if (code == KEY_RIGHT) {
                bpf_printk("RIGHT 1\n");
                return (validate_konamicode_input(ks));
            }
            break;
        case 6:
            if (code == KEY_LEFT) {
                bpf_printk("LEFT 2\n");
                return (validate_konamicode_input(ks));
            }
        case 7:
            if (code == KEY_RIGHT) {
                bpf_printk("RIGHT 2\n");
                return (validate_konamicode_input(ks));
            }
            break;
        case 8:
            if (code == KEY_B) {
                bpf_printk("B\n");
                return (validate_konamicode_input(ks));
            }
            break;
        case 9:
            if (code == KEY_A || code == KEY_Q) { // workaround for azerty. TODO: validate keyboard mapping inputs ?
                bpf_printk("A\n");
                return (validate_konamicode_input(ks));
            }
            break;
        case 10:
            if (code == KEY_ENTER) {
                bpf_printk("ENTER\n");
                activate_konamicode();
                return reset_konamicode(ks);
            }
            break;
        }
        reset_konamicode(ks);
    }
    return 0;
}

char _license[] SEC("license") = "GPL";
__u32 _version SEC("version") = 0xFFFFFFFE;
