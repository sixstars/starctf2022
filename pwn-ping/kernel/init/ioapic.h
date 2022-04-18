#ifndef IOAPIC_H
#define IOAPIC_H

void ioapicinit(void);
void ioapicenable(int irq, int cpunum);


#endif