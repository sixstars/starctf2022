
ENTRY(_begin)
MEMORY
{
    MBR (rwx) :
        ORIGIN = 0x7C00,
        LENGTH = 512

    BSS (rw) :
        ORIGIN = 0x8100,
        LENGTH = 77200
}
SECTIONS
{
    .text : {
        PROVIDE(mbr = .);
        *(.entry);
        *(.text);
        *(.rodata);
        *(.data);
    } > MBR

    .bss : {
        PROVIDE(bss_start = .);
        *(.bss);
        PROVIDE(bss_end = .);
    } > BSS

    /DISCARD/ : {
        *(.eh_frame);
    }
}
