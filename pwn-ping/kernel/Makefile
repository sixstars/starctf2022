include ../.env

kernel :=kernel.elf
INCLUDE_PATH :=-I ../include -Imm -Itrap -Iinit -Idriver -Idriver/disk
ld_script :=kernel.ld.S
libc :=../libs/libc.so

object = init/init.o init/initcpu2.o init/smp.o init/initcpu2.o\
		 init/mp.o init/lapic.o init/ioapic.o\
		 init/cpu.o \
		 driver/keyboard.o \
		 driver/virtio.o driver/virtio_pci.o driver/virtio_ops.o driver/virtio_net.o\
		 driver/virtio_dev.o driver/virtio_gpu.o \
		 driver/smb.o driver/pci.o driver/picirq.o driver/clock.o \
		 driver/disk/ide.o driver/disk/ata.o driver/disk/atapi.o \
		 mm/pmm.o\
		 trap/vector.o trap/trap.o trap/trapenter.o \
		 entry.o \
		 heap/heap.o heap/physical_page.o

CFLAGS :=$(INCLUDE_PATH) -fno-builtin-printf -c -static -Wall -fno-stack-protector -nostdlib -Wno-builtin-declaration-mismatch -Wno-address-of-packed-member -fcf-protection=none

all: $(kernel)
	@echo "kernel compile success"

$(kernel): $(object)
	$(LD) -T $(ld_script) $^ $(libc) -o $(kernel)
	strip $(kernel)

trap/vector.o: trap/vectors.S
	$(CC) $(CFLAGS) -c $^ -o $@

init/initcpu2.o: init/initcpu2.S
	$(CC) $(CFLAGS) -c $^ -o init/initcpu2.1.o
	$(LD) -T init/initcpu2.ld.S init/initcpu2.1.o -o init/initcpu2.2.o
	objcopy -S -O binary -j .text init/initcpu2.2.o init/initcpu2.3.o
	objcopy -I binary -O elf32-i386 -B i386 init/initcpu2.3.o init/initcpu2.o

trap/trapenter.o: trap/trapenter.S
	$(CC) $(CFLAGS) -c $^ -o $@

clean:
	$(RM) $(kernel)
	$(RM) */*.o
	$(RM) */*/*.o
	$(RM) entry.o
