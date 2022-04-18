/*
 * virtio_pci.c
 * Copyright (C) 2021 mac <hzshang15@gmail.com>
 *
 * Distributed under terms of the MIT license.
 */

#include <virtio_pci.h>

#define PCI_MSIX_ENABLE 0x8000
int virtio_read_caps(virtio_pci_dev* pdev){
    Device* dev = pdev->pci;
    uint8_t cap_offset = PCI_readoff(dev,PCI_CAPABILITY_LIST)&0xff;
    while(cap_offset){
        uint32_t value = PCI_readoff(dev,cap_offset);
        uint8_t vndr = value&0xff;
        uint8_t next = (value>>8)&0xff;
        if(vndr == PCI_CAP_ID_MSIX){
            uint16_t flags = (value>>16)&0xffff;
            if (flags & PCI_MSIX_ENABLE)
                pdev->msix_status = VIRTIO_MSIX_ENABLED;
            else
                pdev->msix_status = VIRTIO_MSIX_DISABLED;
        }
        if(vndr != PCI_CAP_ID_VNDR){
            goto next_cap;
        }
        // uint8_t cfg_len = (value>>16)&0xff;
        uint8_t cfg_type = (value>>24)&0xff;
        uint8_t bar = PCI_readoff(dev,cap_offset+4)&0xff;
        uint32_t offset = PCI_readoff(dev,cap_offset + 8);
        // uint32_t length = PCI_readoff(dev,cap_offset + 12);
        switch(cfg_type){
            case VIRTIO_PCI_CAP_COMMON_CFG:
                pdev->common_cfg = (struct virtio_pci_common_cfg*)(dev->reg_base[bar] + offset);
                break;
            case VIRTIO_PCI_CAP_NOTIFY_CFG:
                pdev->notify_base = (uint16_t*)(dev->reg_base[bar] + offset);
                pdev->notify_off_multiplier = PCI_readoff(dev,cap_offset + sizeof(struct virtio_pci_cap));
                break;
            case VIRTIO_PCI_CAP_DEVICE_CFG:
                pdev->device_cfg = (uint8_t*)(dev->reg_base[bar] + offset);
                break;
            case VIRTIO_PCI_CAP_ISR_CFG:
                pdev->isr = (uint8_t*)(dev->reg_base[bar] + offset);
                break;
        }
    next_cap:
        cap_offset = next;
    }
    if(pdev->common_cfg == NULL || pdev->notify_base == NULL 
        || pdev->device_cfg == NULL || pdev->isr == NULL){
        // kprintf("no modern virtio pci device found\n");
        return -1;
    }
    // kprintf("modern virtio pci device found\n");
    // kprintf("common_cfg map at: %x\n",pdev->common_cfg);
    // kprintf("notify_base map at: %x\n",pdev->notify_base);
    // kprintf("device_cfg map at: %x\n",pdev->device_cfg);
    // kprintf("isr map at: %x\n",pdev->isr);
    return 0;
}






