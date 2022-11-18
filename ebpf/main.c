#include <linux/input-event-codes.h>

#include "include/all.h"

struct konamicode_status {
    int completion;
};

struct bpf_map_def SEC("maps/konamicode_sequence") konamicode_sequence = {
    .type = BPF_MAP_TYPE_ARRAY,
    .key_size = sizeof(int),
    .value_size = sizeof(struct konamicode_status),
    .max_entries = 1,
};

struct bpf_map_def SEC("maps/notes") notes = {
    .type = BPF_MAP_TYPE_PERF_EVENT_ARRAY,
    .max_entries = 0,
    .pinning = 0,
    .namespace = "",
};

struct sound_note {
    u64 freq;
    u64 duration;
};

long __attribute__((always_inline)) push_sound_note(void *ctx, struct sound_note *n) {
    bpf_printk("pushing note to perf buffer\n");
    u32 cpu = bpf_get_smp_processor_id();
    return bpf_perf_event_output(ctx, &notes, cpu, n, sizeof(*n));
}

int __attribute__((always_inline)) validate_konamicode_input(struct konamicode_status* ks) {
    ks->completion++;
    int key = 0;
    return bpf_map_update_elem(&konamicode_sequence, &key, ks, BPF_ANY);
}

int __attribute__((always_inline)) reset_konamicode(struct konamicode_status* ks) {
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

int __attribute__((always_inline)) activate_konamicode(void *ctx) {
    int key = 0;
    int* counter = bpf_map_lookup_elem(&konamicode_activation_counter, &key);
    if (counter == NULL) {
        return -1;
    }
    *counter += 1;
    bpf_printk("KONAMI CODE entered \\o/ (%i times)\n", *counter);
    int err = bpf_map_update_elem(&konamicode_activation_counter, &key, counter, BPF_ANY);
    if (err) {
        return err;
    }

    struct sound_note n = {
        .freq = NOTE_AS,
        .duration = 1000,
    };
    push_sound_note(ctx, &n);

    n.freq = NOTE_B;
    return push_sound_note(ctx, &n);
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
            return 0;
        }

        unsigned int completion = ks->completion;
        switch (completion) {
        case 0:
            if (code == KEY_UP) {
                bpf_printk("UP\n");
                struct sound_note n = {
                    .freq = NOTE_C,
                    .duration = 1000,
                };
                push_sound_note(ctx, &n);
                return (validate_konamicode_input(ks));
            }
            break;
        case 1:
            if (code == KEY_UP) {
                bpf_printk("UP\n");
                struct sound_note n = {
                    .freq = NOTE_CS,
                    .duration = 1000,
                };
                push_sound_note(ctx, &n);
                return (validate_konamicode_input(ks));
            }
            break;
        case 2:
            if (code == KEY_DOWN) {
                bpf_printk("DOWN\n");
                struct sound_note n = {
                    .freq = NOTE_D,
                    .duration = 1000,
                };
                push_sound_note(ctx, &n);
                return (validate_konamicode_input(ks));
            }
            break;
        case 3:
            if (code == KEY_DOWN) {
                bpf_printk("DOWN\n");
                struct sound_note n = {
                    .freq = NOTE_DS,
                    .duration = 1000,
                };
                push_sound_note(ctx, &n);
                return (validate_konamicode_input(ks));
            }
            break;
        case 4:
            if (code == KEY_LEFT) {
                bpf_printk("LEFT\n");
                struct sound_note n = {
                    .freq = NOTE_E,
                    .duration = 1000,
                };
                push_sound_note(ctx, &n);
                return (validate_konamicode_input(ks));
            }
        case 5:
            if (code == KEY_RIGHT) {
                bpf_printk("RIGHT\n");
                struct sound_note n = {
                    .freq = NOTE_F,
                    .duration = 1000,
                };
                push_sound_note(ctx, &n);
                return (validate_konamicode_input(ks));
            }
            break;
        case 6:
            if (code == KEY_LEFT) {
                bpf_printk("LEFT\n");
                struct sound_note n = {
                    .freq = NOTE_FS,
                    .duration = 1000,
                };
                push_sound_note(ctx, &n);
                return (validate_konamicode_input(ks));
            }
        case 7:
            if (code == KEY_RIGHT) {
                bpf_printk("RIGHT\n");
                struct sound_note n = {
                    .freq = NOTE_G,
                    .duration = 1000,
                };
                push_sound_note(ctx, &n);
                return (validate_konamicode_input(ks));
            }
            break;
        case 8:
            if (code == KEY_B) {
                bpf_printk("B\n");
                struct sound_note n = {
                    .freq = NOTE_GS,
                    .duration = 1000,
                };
                push_sound_note(ctx, &n);
                return (validate_konamicode_input(ks));
            }
            break;
        case 9:
            if (code == KEY_A || code == KEY_Q) { // workaround for azerty. TODO: validate keyboard mapping inputs ?
                bpf_printk("A\n");
                struct sound_note n = {
                    .freq = NOTE_A,
                    .duration = 1000,
                };
                push_sound_note(ctx, &n);
                return (validate_konamicode_input(ks));
            }
            break;
        case 10:
            if (code == KEY_ENTER) {
                activate_konamicode(ctx);
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
