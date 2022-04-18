#include "asm.h"

.set PROT_MODE_CSEG,	0x8         # kernel code segment selector
.set PROT_MODE_DSEG,    0x10        # kernel data segment selector
.set CR0_PE_ON, 	0x1         # protected mode enable flag
.set SMAP,          0x534d4150

.section .entry
.code16

.globl _begin
_begin:
        cli
        cld
	# clean register
        xorw    %ax,%ax
        movw    %ax,%ds
        movw    %ax,%es
        movw    %ax,%ss
        jmp .

#Enable A20
seta20.1:
        inb     $0x64,%al           # Wait for not busy(8042 input buffer empty).
        testb   $0x2,%al            #
        jnz     seta20.1
        movb    $0xd1,%al           # 0xd1 -> port 0x64
        outb    %al,$0x64           # 0xd1 means: write data to 8042's P2 port
seta20.2:
        inb     $0x64,%al           # Wait for not busy(8042 input buffer empty).
        testb   $0x2,%al            # Busy?
        jnz     seta20.2
        movb    $0xdf,%al           # 0xdf -> port 0x60
        outb    %al,$0x60           # 0xdf = 11011111, means set P2's A20 bit(the 1 bit) to 1

/*
#now probe memory
probe_memory:
        movl $0,0x8000
        xorl %ebx, %ebx
        movw $0x8004, %di
start_probe:
        movl $SMAP, %edx
        movl $20, %ecx
        movl $0xe820, %eax
        int $0x15
        jnc cont
        movw $12345, 0x8000
        jmp finish_probe
cont:
        addw $20, %di
        incl 0x8000
        cmpl $0, %ebx
        jnz start_probe
*/
finish_probe: 
# Switch from real to protected mode, using a bootstrap GDT
# and segment translation that makes virtual addresses
# identical to physical addresses, so that the
# effective memory map does not change during the switch.
        lgdt gdtdesc
        movl %cr0, %eax
    	orl $CR0_PE_ON, %eax
    	movl %eax, %cr0
        ljmp $PROT_MODE_CSEG, $protcseg

.code32
protcseg:
        # Set up the protected-mode data segment registers
        movw $PROT_MODE_DSEG, %ax   # Our data segment selector
        movw %ax, %ds               # -> DS: Data Segment
        movw %ax, %es               # -> ES: Extra Segment
        movw %ax, %fs               # -> FS
        movw %ax, %gs               # -> GS
        movw %ax, %ss               # -> SS: Stack Segment
        movl $0x0, %ebp
        movl $_begin, %esp
        call bootmain
loop:
        jmp loop

.section .data
# Bootstrap GDT
.p2align 2                                          # force 4 byte alignment
gdt:
    SEG_NULLASM                                     # null seg
    SEG_ASM(STA_X|STA_R, 0x0, 0xffffffff)           # code seg for bootloader and kernel
    SEG_ASM(STA_W, 0x0, 0xffffffff)                 # data seg for bootloader and kernel

gdtdesc:
    .word 0x17                                      # sizeof(gdt) - 1
    .long gdt                                       # address gdt
