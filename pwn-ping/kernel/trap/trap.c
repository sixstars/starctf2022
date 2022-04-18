#include <trap.h>
#include <types.h>
#include <x86.h>
#include <mmu.h>
#include <memlayout.h>
#include <stdio.h>
#include <clock.h>
#include <picirq.h>
#include <ioapic.h>
#include <lapic.h>
static struct gatedesc idt[256] = {{0}};
static struct pseudodesc idt_pd = {
    sizeof(idt) - 1, (uintptr_t)idt
};
void (*intr_array[0x100])(struct trapframe*);

void intr_init(){

    extern uintptr_t __vectors[];
    for(int i=0;i<sizeof(idt)/sizeof(struct gatedesc);i++){
        SETGATE(idt[i],0,GD_KTEXT, __vectors[i], DPL_KERNEL);
    }
    SETGATE(idt[T_SWITCH_TOK], 0, GD_KTEXT, __vectors[T_SWITCH_TOK], DPL_USER);
    memset(intr_array,0,sizeof(intr_array));
    lidt(&idt_pd);
}

void intr_enable(){
    sti();
}

void intr_disable(){
    cli();
}

void trap(struct trapframe *tf){
    // uint32_t trap_num = tf->tf_trapno - IRQ_OFFSET;
    if(intr_array[tf->tf_trapno]){
        intr_array[tf->tf_trapno](tf);
    }else{
        printf("unknown intr number: %d\n",tf->tf_trapno);
        printf("eip at: 0x%08x\n",tf->tf_eip);
        while(1){
            asm("hlt");
        }
    }
}

void register_intr_handler(int intr,void (*fn)(struct trapframe*)){
    intr_array[intr] = fn;
}











