// global PIC controller

#include <types.h>
#include <trap.h>
#include <stdio.h>
#define IOAPIC  0xFEC00000   // Default physical address of IO APIC

#define REG_ID     0x00  // Register index: ID
#define REG_VER    0x01  // Register index: version
#define REG_TABLE  0x10  // Redirection table base

// The redirection table starts at REG_TABLE and uses
// two registers to configure each interrupt.
// The first (low) register in a pair contains configuration bits.
// The second (high) register contains a bitmask telling which
// CPUs can serve that interrupt.
#define INT_DISABLED   0x00010000  // Interrupt disabled
#define INT_LEVEL      0x00008000  // Level-triggered (vs edge-)
#define INT_ACTIVELOW  0x00002000  // Active low (vs high)
#define INT_LOGICAL    0x00000800  // Destination is CPU id (vs APIC ID)

// IO APIC MMIO structure: write reg, then read or write data.
struct ioapic {
  uint reg;
  uint pad[3];
  uint data;
};
volatile struct ioapic *ioapic;
static uint ioapicread(int reg)
{
  ioapic->reg = reg;
  return ioapic->data;
}

static void ioapicwrite(int reg, uint data)
{
  ioapic->reg = reg;
  ioapic->data = data;
}


void ioapicinit(void){
  int i, maxintr;

  // if(!ismp)
  //   return;

  ioapic = (volatile struct ioapic*)IOAPIC;
  maxintr = (ioapicread(REG_VER) >> 16) & 0xFF;
  // id = ioapicread(REG_ID) >> 24;
  // if(id != ioapicid)
    // printf("ioapicinit: id isn't equal to ioapicid; not a MP\n");

  // Mark all interrupts edge-triggered, active high, disabled,
  // and not routed to any CPUs.
  for(i = 0; i <= maxintr; i++){
    ioapicwrite(REG_TABLE+2*i, INT_DISABLED | (IRQ_OFFSET + i));
    ioapicwrite(REG_TABLE+2*i+1, 0);
  }
  printf("support max intr number: %d\n",maxintr);
}

void ioapicenable(int irq, int cpunum)
{
  // Mark interrupt edge-triggered, active high,
  // enabled, and routed to the given cpunum,
  // which happens to be that cpu's APIC ID.
  ioapicwrite(REG_TABLE+2*irq, IRQ_OFFSET + irq);
  ioapicwrite(REG_TABLE+2*irq+1, cpunum << 24);
}




