

#include <x86.h>
#include <clock.h>
#include <libcc.h>
#include <smp.h>

#define STARTUP_CODE_ADDRESS 0x1000

#define PIC_ICR_ADDRESS 0xFEE00300

extern void initcpu2();

extern int lock[];
extern int kernel_stack2[];
extern char _binary_init_initcpu2_3_o_start[];
extern char _binary_init_initcpu2_3_o_end[];
extern int _binary_init_initcpu2_3_o_size[];
volatile DECLARE_LOCK(taskLock);
volatile void (*task)(void*);
void* task_arg;

void cpu2_run();
void start_thread(){
	LOCK(taskLock);
	kernel_stack2[0] = (uint32_t)&cpu2_run;
	uint32_t size = _binary_init_initcpu2_3_o_end - _binary_init_initcpu2_3_o_start;
	memcpy((void*)STARTUP_CODE_ADDRESS,_binary_init_initcpu2_3_o_start,size);
	uint32_t* lock = (uint32_t*)(size + STARTUP_CODE_ADDRESS - 4);
	// *lock = 0;
	uint32_t* icr_addr = (uint32_t*)PIC_ICR_ADDRESS;
	*icr_addr = 0x000C4500;
	babysleep(100);
	*icr_addr = STARTUP_CODE_ADDRESS/0x1000 + 0x000C4600;
	babysleep(100);
	while(*lock == 0xdead){
		babysleep(100);
	}
}

void cpu2_run(){
	debug("cpu2 start! wait task\n");
	while(1){
		LOCK(taskLock);
		void (*tmp)(void*);
		tmp = task;
		void* tmp_arg = task_arg;
		UNLOCK(taskLock);
		task = 0;
		tmp(tmp_arg);
	}
}


void create_task(void (*p)(void*),void* arg){
    task = p;
    task_arg = arg;
    UNLOCK(taskLock);
    while(task){
        cpu_relax();
    }
    LOCK(taskLock);
}


