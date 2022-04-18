include ../.env


INCLUDE_PATH :=../include
CFLAGS := -I $(INCLUDE_PATH) -Wall -gstabs -fno-stack-protector -Os -nostdinc -fno-pic -c

main_src = bootmain.c
bootasm_src = bootasm.S
main_obj = $(main_src:.c=.o)
bootasm_obj = $(bootasm_src:.S=.o)
ld_script = boot.ld.S
target = bootloader
elf = $(target:=.elf)

all: $(target)
	@echo "bootloader compile success"

$(target) : $(bootasm_obj) $(main_obj) $(ld_script)
	$(LD) -T $(word 3,$^) -o $(elf) $(word 1,$^) $(word 2,$^)
	$(OBJCOPY) -S -O binary -j .text $(elf) $@

$(bootasm_obj) :$(bootasm_src)
	$(CC) $(CFLAGS) $^ -o $@ 

$(main_obj) :$(main_src)
	$(CC) $(CFLAGS) $^ -o $@ 

clean:
	-$(RM) $(target) $(elf) $(main_obj) $(bootasm_obj)
