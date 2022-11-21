NAME	=	konamicode_detector
$(NAME): all

KERNEL_VERSION	=	$(shell uname -r)
KERNEL_INCLUDE_PATH	=	/usr/src/linux-headers-$(KERNEL_VERSION)

build-ebpf:
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

build: build-ebpf
	go build -o $(NAME) .

build-with-sound: build-ebpf
	go build -tags libasound -o $(NAME) .

all: build

clean:
	find . -name "*~"|xargs rm -f
	rm -fr ebpf/bin/
	rm -f $(NAME)

run: $(NAME)
	sudo ./$(NAME)

run-with-sound: build-with-sound
	sudo ./$(NAME)
