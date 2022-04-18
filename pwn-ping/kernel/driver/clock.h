#ifndef __KERN_DRIVER_CLOCK_H__
#define __KERN_DRIVER_CLOCK_H__

#include <types.h>

extern volatile size_t ticks;

void clock_init(void);
void babysleep(size_t sleep);
#endif /* !__KERN_DRIVER_CLOCK_H__ */

