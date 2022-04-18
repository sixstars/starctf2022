/* Simple linker script for the JOS kernel.
   See the GNU ld 'info' manual ("info ld") to learn the syntax. */

OUTPUT_FORMAT("elf32-i386", "elf32-i386", "elf32-i386")
OUTPUT_ARCH(i386)
ENTRY(loader)

SECTIONS {	
	. = 0x100000;
	.text : {
		*(.multiboot)
		*(.text .stub .text.* .gnu.linkonce.t.*)
	}
	PROVIDE(etext = .);	/* Define the 'etext' symbol to this value */
	.rodata : {
		*(.rodata .rodata.* .gnu.linkonce.r.*)
	}

	/* Include debugging information in kernel memory */
	.stab : {
		PROVIDE(__STAB_BEGIN__ = .);
		*(.stab);
		PROVIDE(__STAB_END__ = .);
		BYTE(0)		/* Force the linker to allocate space
				   for this section */
	}

	.stabstr : {
		PROVIDE(__STABSTR_BEGIN__ = .);
		*(.stabstr);
		PROVIDE(__STABSTR_END__ = .);
		BYTE(0)		/* Force the linker to allocate space
				   for this section */
	}

	/* Adjust the address for the data segment to the next page */
	. = ALIGN(0x1000);

	/* The data segment */
	.data : {
		*(.data)
	}
    . = ALIGN(0x1000);

	.bss : {
        PROVIDE(bss_start = .);
		*(.bss)
        PROVIDE(bss_end = .);
	}
    . = ALIGN(0x1000);
    . += 0x20000;
    PROVIDE(kernel_stack = .);


    . = 0x1000;

    .textcpu2 : {
    		*(.text2.cpu2)
    }

    . = 0x800000;
    PROVIDE(kernel_stack2 = .);

	/DISCARD/ : {
		*(.eh_frame .note.GNU-stack)
	}
}
