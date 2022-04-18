#include <elf.h>
#include <x86.h>
#define SECTSIZE 512
#define ELF_BUFFER 0x10000
extern uint8_t mbr[];

static void readseg(uintptr_t va, uint32_t count, uint32_t offset);
static void readsect(void *dst, uint32_t secno);
static void waitdisk(void);
static inline uint32_t get_lba(int n);


void bootmain(void) {

	struct elfhdr *elf=(struct elfhdr*)ELF_BUFFER;
	uint32_t base=get_lba(2)*SECTSIZE;//the offset of kernel in the disk

    readseg((uintptr_t)elf, SECTSIZE * 8, base);

    // is this a valid ELF?
    if (elf->e_magic != ELF_MAGIC) {
        goto bad;
    }
    struct proghdr *ph, *eph;
    ph = (struct proghdr *)((uintptr_t)elf + elf->e_phoff);
    eph = ph + elf->e_phnum;
    for (; ph < eph; ph ++) {
    	uintptr_t va=(uintptr_t)(ph->p_va & 0xFFFFFFFF);
        readseg(va, ph->p_memsz, base+ph->p_offset);
    }

    // call the entry point from the ELF header
    // note: does not return
    void (*entry)(void)=(void(*)(void))(elf->e_entry & 0xFFFFFF);
    entry();

bad:
    /* do nothing */
    while (1);
}

void waitdisk(void){
    while ((inb(0x1F7) & 0xC0) != 0x40)
        /* do nothing */;
}

void readsect(void *dst,uint32_t secno){
    // wait for disk to be ready
    waitdisk();

    outb(0x1F2, 1);                         // count = 1
    outb(0x1F3, secno & 0xFF);
    outb(0x1F4, (secno >> 8) & 0xFF);
    outb(0x1F5, (secno >> 16) & 0xFF);
    outb(0x1F6, ((secno >> 24) & 0xF) | 0xE0);
    outb(0x1F7, 0x20);                      // cmd 0x20 - read sectors

    // wait for disk to be ready
    waitdisk();

    // read a sector
    insl(0x1F0, dst, SECTSIZE);
}

void readseg(uintptr_t va,uint32_t count,uint32_t offset){
    uintptr_t end_va = va + count;

    // round down to sector boundary
    va -= offset % SECTSIZE;

    // translate from bytes to sectors; kernel starts at sector 1
    uint32_t secno = (offset / SECTSIZE);

    // If this is too slow, we could read lots of sectors at a time.
    // We'd write more to memory than asked, but it doesn't matter --
    // we load in increasing order.
    for (; va < end_va; va += SECTSIZE, secno ++) {
        readsect((void *)va, secno);
    }
}

static inline uint32_t get_lba(int n){
	uint32_t lba;
	for (int i = 0; i < 4; i++) {
		((uint8_t *)&lba)[i] = mbr[454 + 16 * (n - 1) + i];
	}
	return lba;
}
