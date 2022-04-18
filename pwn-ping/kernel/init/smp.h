#ifndef SMP_H
#define SMP_H

void start_thread();

extern volatile DECLARE_LOCK(taskLock);

extern volatile void (*task)(void*);
extern void* task_arg;
void create_task(void (*p)(void*),void* arg);
#endif
