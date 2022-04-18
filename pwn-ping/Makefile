include .env

INCLUDE_PATH :=include
KERN_LD_SCRIPT :=tools/kernel.ld.S

kernel := kernel/kernel.elf
libc := libs/libc.so
img := kernel.iso

.PHONY: $(img) $(kernel) clean build

all: $(img)
	@echo "make done"

$(img): $(kernel)
	mkdir iso
	mkdir iso/boot
	mkdir iso/boot/grub
	cp $< iso/boot/kernel.elf
	echo 'set timeout=0' > iso/boot/grub/grub.cfg
	echo 'set default=0' >> iso/boot/grub/grub.cfg
	echo '' >> iso/boot/grub/grub.cfg
	echo 'menuentry "my os" {' >> iso/boot/grub/grub.cfg
	echo '  multiboot /boot/kernel.elf' >> iso/boot/grub/grub.cfg
	echo '  boot' >> iso/boot/grub/grub.cfg
	echo '}' >> iso/boot/grub/grub.cfg
	grub-mkrescue --output=kernel.iso iso
	rm -rf iso

gdb:$(img)
	sudo $(QEMU) -s -cdrom kernel.iso \
		-netdev tap,id=n1,if=virtio,ifname=tap100 \
		-device virtio-net-pci,netdev=n1,mac=cc:dd:ee:ff:aa:bb,disable-modern="on" \
		-smp 2 -m 2048M -curses -monitor telnet:0.0.0.0:4444,server,nowait



qemu: 
	sudo $(QEMU) -cdrom kernel.iso \
		-hda flag.txt \
		-netdev tap,id=n1,ifname=tap100 \
		-device virtio-net-pci,netdev=n1,mac=01:02:03:04:05:06 \
		-smp 1 -m 64M -curses -monitor telnet:0.0.0.0:4444,server,nowait



$(kernel): |$(libc)
	$(MAKE) -C kernel

$(libc):
	+$(MAKE) -C libs

clean:
	+$(MAKE) clean -C libs
	+$(MAKE) clean -C kernel
	$(RM) $(img)
