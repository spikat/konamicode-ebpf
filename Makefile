KERNEL_VERSION	=	$(shell uname -r)
KERNEL_INCLUDE_PATH	=	/usr/src/linux-headers-$(KERNEL_VERSION)

build-ebpf:
	echo $(KERNEL_VERSION)
	mkdir -p ebpf/bin
	clang -D__KERNEL__ -D__ASM_SYSREG_H \
		-Wno-unused-value \
		-Wno-pointer-sign \
		-Wno-compare-distinct-pointer-types \
		-Wunused \
		-Wall \
		-Werror \
		-I${KERNEL_INCLUDE_PATH}/include \
		-I$(KERNEL_INCLUDE_PATH)/include/uapi \
		-I$(KERNEL_INCLUDE_PATH)/include/generated/uapi \
		-I$(KERNEL_INCLUDE_PATH)/arch/x86/include \
		-I$(KERNEL_INCLUDE_PATH)/arch/x86/include/uapi \
		-I$(KERNEL_INCLUDE_PATH)/arch/x86/include/generated \
		-O2 -emit-llvm \
		ebpf/main.c \
		-c -o - | llc -march=bpf -filetype=obj -o ebpf/bin/probe.o

build:
	go build -o bin/main .

all: build-ebpf build

clean:
	rm -fr *~ bin/ ebpf/bin/ ebpf/*~

run:
	sudo bin/main
