#include <pmm.h>
#include <types.h>
#include <mmu.h>
#include <x86.h>
#include <memlayout.h>

static struct taskstate ts = {0};

static struct segdesc gdt[] = {
    SEG_NULL,
    [SEG_KTEXT] = SEG(STA_X | STA_R, 0x0, 0xFFFFFFFF, DPL_KERNEL),
    [SEG_KDATA] = SEG(STA_W, 0x0, 0xFFFFFFFF, DPL_KERNEL),
    [SEG_UTEXT] = SEG(STA_X | STA_R, 0x0, 0xFFFFFFFF, DPL_USER),
    [SEG_UDATA] = SEG(STA_W, 0x0, 0xFFFFFFFF, DPL_USER),
    [SEG_TSS]    = SEG_NULL,
};

static struct pseudodesc gdt_pd = {
    sizeof(gdt) - 1, (uint32_t)gdt
};
static inline void lgdt( struct pseudodesc *pd ){
    asm("lgdt (%0)"::"r"(pd));
    asm("movw %%ax,%%gs"::"a"(USER_DS));
    asm("movw %%ax,%%fs"::"a"(USER_DS));
    asm("movw %%ax,%%es"::"a"(KERNEL_DS));
    asm("movw %%ax,%%ds"::"a"(KERNEL_DS));
    asm("movw %%ax,%%ss"::"a"(KERNEL_DS));
    asm("ljmp %0, $1f\n 1:\n"::"i"(KERNEL_CS));
}
uint8_t stack0[1024];

static void gdt_init(){
// create a TSS descriptor entry in GDT
// add enough information to the TSS in memory as needed
// load the TR register with a segment selector for that segment
    ts.ts_esp0 = (uint32_t)&stack0 + sizeof(stack0);
    ts.ts_ss0 = KERNEL_DS;
    gdt[SEG_TSS] = SEG16(STS_T32A, (uint32_t)&ts, sizeof(ts), DPL_KERNEL);
    gdt[SEG_TSS].sd_s = 0;
    lgdt(&gdt_pd);
    ltr(GD_TSS);
}
void pmm_init(){
    gdt_init();
}
