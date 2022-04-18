
OUTPUT_FORMAT("elf32-i386", "elf32-i386", "elf32-i386")
OUTPUT_ARCH(i386)
ENTRY(initcpu2)

SECTIONS {	
	. = 0x1000;
	.text : {
		*(.text)
	}
	/DISCARD/ : {
		*(.eh_frame .note.GNU-stack)
	}
    . = 0x800000;
    PROVIDE(kernel_stack2 = .);
}
