/*
 * pci.h
 * Copyright (C) 2021 mac <hzshang15@gmail.com>
 *
 * Distributed under terms of the MIT license.
 */

#ifndef PCI_H
#define PCI_H
#include <types.h>
#include <x86.h>

#define CONFIG_ADDRESS 0xCF8
#define CONFIG_DATA 0xCFC
#define BAR0 0x10
#define COMMAND 0x4


#define OFFSET_STATUS 6
#define OFFSET_CAP 0x34
#define PCI_MAPREG_START        0x10
#define PCI_MAPREG_END          0x28

#define PCI_MAPREG_NUM(offset)                      \
    (((unsigned)(offset)-PCI_MAPREG_START)/4)

#define PCI_MAPREG_TYPE_MASK            0x00000001
#define PCI_MAPREG_TYPE(mr)                     \
            ((mr) & PCI_MAPREG_TYPE_MASK)
#define PCI_MAPREG_TYPE_MEM         0x00000000
#define PCI_MAPREG_TYPE_IO          0x00000001

#define PCI_MAPREG_MEM_TYPE_MASK        0x00000006
#define PCI_MAPREG_MEM_TYPE(mr)                     \
            ((mr) & PCI_MAPREG_MEM_TYPE_MASK)

#define PCI_MAPREG_MEM_TYPE_32BIT       0x00000000
#define PCI_MAPREG_MEM_TYPE_64BIT       0x00000004

#define PCI_MAPREG_MEM_ADDR_MASK        0xfffffff0
#define PCI_MAPREG_MEM_ADDR(mr)                     \
            ((mr) & PCI_MAPREG_MEM_ADDR_MASK)

#define PCI_MAPREG_MEM_SIZE(mr)                     \
            (PCI_MAPREG_MEM_ADDR(mr) & -PCI_MAPREG_MEM_ADDR(mr))


#define PCI_MAPREG_IO_ADDR_MASK         0xfffffffc
#define PCI_MAPREG_IO_ADDR(mr)                      \
            ((mr) & PCI_MAPREG_IO_ADDR_MASK)
#define PCI_MAPREG_IO_SIZE(mr)                      \
            (PCI_MAPREG_IO_ADDR(mr) & -PCI_MAPREG_IO_ADDR(mr))


typedef struct {
    uint16_t bus_id;
    uint16_t device_id;
    uint8_t func;
    // uint32_t addr; // (bus<<16) | (device<<11) | func<<8) 
    uint16_t vendor;
    uint16_t device;

    uint32_t iobase;
    uint32_t membase;
    uint32_t iosize;

    uint8_t irq;
    uint8_t intpin;
    uint8_t ioapicPin;
    uint8_t ioapicid;
    uint8_t subsystem_id;
    uint8_t subsystem_vendorid;
    uint8_t class_code;
    uint8_t subclass;
    uint8_t prog_if;
    uint8_t revision_id;
    uint32_t reg_base[6];
    uint32_t reg_size[6];
}Device;
uint32_t PCI_read(uint8_t bus,uint8_t device,uint16_t func,uint8_t offset);
void PCI_write(uint8_t bus,uint8_t device,uint16_t func,uint8_t offset,uint32_t);
Device* registerDevice(uint16_t bus,uint16_t device,uint16_t func);
void PCI_enableBusmaster(Device*);
void pci_init();

uint32_t PCI_readBar(Device* d,int idx);
void PCI_writeBar(Device* d,int idx,uint32_t value);
uint32_t PCI_readoff(Device*d,uint32_t offset);
void PCI_loadbars(Device* f);

extern Device devices[];
extern int device_num;



#endif /* !PCI_H */




