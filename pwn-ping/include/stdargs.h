#ifndef __INCLUDE_STDARGS_H
#define __INCLUDE_STDARGS_H
//treat stack memory as chars stream
#define args_list char *
//get the size of type in the stack
#define _arg_stack_size(type) (((sizeof(type)-1)/sizeof(int)+1)*sizeof(int))

//get the stack address of args after fmt
#define args_start(ap, fmt) do {        \
ap = (char *)((unsigned int)&fmt + _arg_stack_size(&fmt));      \
} while (0)
//do noting
#define args_end(ap)
// get the address of next arg
#define args_next(ap, type) (((type *)(ap+=_arg_stack_size(type)))[-1])


#endif
