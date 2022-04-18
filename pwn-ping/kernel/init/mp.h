#ifndef __MP_H
#define __MP_H
#include <types.h>
struct mp {             // floating pointer
  uint8_t signature[4];           // "_MP_"
  void *physaddr;               // phys addr of MP config table
  uint8_t length;                 // 1
  uint8_t specrev;                // [14]
  uint8_t checksum;               // all bytes must add up to 0
  uint8_t type;                   // MP system config type
  uint8_t imcrp;
  uint8_t reserved[3];
};

struct mpconf {         // configuration table header
  uint8_t signature[4];           // "PCMP"
  uint16_t length;                // total table length
  uint8_t version;                // [14]
  uint8_t checksum;               // all bytes must add up to 0
  uint8_t product[20];            // product id
  uint *oemtable;               // OEM table pointer
  uint16_t oemlength;             // OEM table length
  uint16_t entry;                 // entry count
  uint *lapicaddr;              // address of local APIC
  uint16_t xlength;               // extended table length
  uint8_t xchecksum;              // extended table checksum
  uint8_t reserved;
};

struct mpproc {         // processor table entry
  uint8_t type;                   // entry type (0)
  uint8_t apicid;                 // local APIC id
  uint8_t version;                // local APIC verison
  uint8_t flags;                  // CPU flags
    #define MPBOOT 0x02           // This proc is the bootstrap processor.
  uint8_t signature[4];           // CPU signature
  uint feature;                 // feature flags from CPUID instruction
  uint8_t reserved[8];
};

struct mpioapic {       // I/O APIC table entry
  uint8_t type;                   // entry type (2)
  uint8_t apicno;                 // I/O APIC id
  uint8_t version;                // I/O APIC version
  uint8_t flags;                  // I/O APIC flags
  uint *addr;                  // I/O APIC address
};
struct mpbus{
  uint8_t type;
  uint8_t bus_id;
  uint8_t name[6];
} __attribute__((packed));

struct mpinst{
  uint8_t type;
  uint8_t intr_type;
  uint16_t io_intr_flag;
  uint8_t src_bus_id;
  uint8_t src_bus_irq;
  uint8_t dst_apic_id;
  uint8_t dst_apic_intin;
} __attribute__((packed));

// Table entry types
#define MPPROC    0x00  // One per processor
#define MPBUS     0x01  // One per bus
#define MPIOAPIC  0x02  // One per I/O APIC
#define MPIOINTR  0x03  // One per bus interrupt source
#define MPLINTR   0x04  // One per system interrupt source
void mpinit();
int getIRQPin(char* bus,int irq);
#endif