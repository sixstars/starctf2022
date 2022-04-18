#include <stdio.h>
#include <pmm.h>
#include <trap.h>
#include <x86.h>
#include <picirq.h>
#include <clock.h>
#include <multiboot.h>
#include <heap.h>
#include <pci.h>
#include <keyboard.h>
#include <virtio_dev.h>
#include <physical_page.h>
#include <smp.h>
#include <virtio_gpu.h>
#include <virtio.h>
#include <smb.h>
#include <screen.h>
#include <mp.h>
#include <ioapic.h>
#include <lapic.h>
#include <ide.h>
void dma_test();
void network_send_packet(uint8_t* pkt,size_t length);
void* flag_ptr;
void kern_init(multiboot_info_t* mbd, unsigned int magic){
    extern char bss_start[],bss_end[];
    memset(bss_start,0,bss_end-bss_start);
    cga_init();
    pmm_init();
    heap_init((uint8_t*)0x200000,0x100000);
    physical_page_init((uint8_t*)0x300000,0x500000);

    intr_init();
    pic_init();
    clock_init();
    keyboard_init();

    pci_init();
    virtio_dev_install();
    smb_init();
    intr_enable();
    ide_init();
    flag_ptr = physical_alloc(0x400,0x10000);
    ide_disable_dma(0);
    ide_read_sectors(0,1,0,flag_ptr);
    printf("load flag:%s\n",flag_ptr);
    printf("ping me...\n");
    
extern uint8_t outnetbuf[];
extern uint32_t outnetsize;
extern uint8_t innetbuf[];
extern uint32_t innetsize;
void network_main(uint8_t*,int size,uint8_t*,uint32_t*);
    while(1){
        if (innetsize){
            network_main(innetbuf + 12,innetsize - 12,outnetbuf,&outnetsize);
            innetsize = 0;
        }
        if(outnetsize){
            network_send_packet(outnetbuf,outnetsize);
            outnetsize = 0;
        }
        asm("hlt");
    }

}





