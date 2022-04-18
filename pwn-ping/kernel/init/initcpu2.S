#include "asm.h"

.section .text

.code16
.global initcpu2
.global initcpu2_end

# .set SPINLOCK_ADDRESS, 0x2000
.set PROT_MODE_CSEG,	0x8         # kernel code segment selector
.set PROT_MODE_DSEG,    0x10        # kernel data segment selector
.set CR0_PE_ON, 	0x1             # protected mode enable flag

initcpu2:
	cli
	cld
	xor %ax, %ax
	mov %ax, %ds
	movb $1, lock
    wbinvd
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
    movl $kernel_stack2, %esp
    ret

# Bootstrap GDT
.p2align 2                                          # force 4 byte alignment
gdt:
    SEG_NULLASM                                     # null seg
    SEG_ASM(STA_X|STA_R, 0x0, 0xffffffff)           # code seg for bootloader and kernel
    SEG_ASM(STA_W, 0x0, 0xffffffff)                 # data seg for bootloader and kernel

gdtdesc:
    .word 0x17                                      # sizeof(gdt) - 1
    .long gdt                                       # address gdt
lock:
    .word 0xdead

initcpu2_end:
	.word 0
