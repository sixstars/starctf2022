#ifndef CPU_H
#define CPU_H

#include <types.h>
typedef struct {
	uint8_t apicid;
} CPU;

extern CPU cpus[];
extern int ncpu;


#endif