/*
 * heap.h
 * Copyright (C) 2018 hzshang <hzshang15@gmail.com>
 *
 * Distributed under terms of the MIT license.
 */

#ifndef HEAP_H
#define HEAP_H
#include <types.h>

#define HEAP_SIZE 0x100000
#define BINS_SIZE 0x50
#define USE 0x1
#define NOUSE 0x0
#define MAP 0x2
#define NOMAP 0x0

/*
 *    Chunk struct
 *   |--------------------|
 *   | Last Chunk Size    |  last bit 0:free, 1:use
 *   |--------------------|  
 *   | Chunk Size         |
 *   |--------------------|
 *   | Content | Next ptr |
 *   |--------------------|
 */
struct chunk{
    size_t last_size;
    size_t size;
    struct chunk* next;
};
typedef struct chunk* chunk_ptr;
//READ/WRITE
#define chunk_content(ptr) ((void *)&(ptr->next))
#define get_chunk_by_content(content) ((chunk_ptr)(content-2*sizeof(size_t)))

//ONLY READ
#define chunk_is_use(ptr) (ptr->last_size & 0x1)
#define chunk_flag(ptr) (ptr->last_size & 0x3)
#define chunk_last_size(ptr) (ptr->last_size & (~0x3))

#define chunk_down(ptr) ((chunk_ptr)((void*)ptr + ptr->size))
#define chunk_up(ptr) ((chunk_ptr)((void*)ptr - chunk_last_size(ptr)))

#define index_of_bins(size) ((size-2*sizeof(void*))/sizeof(chunk_ptr))
#define mem_chunk_size(size) (((size+sizeof(void*)-1) & (~(sizeof(void*)-1))) + 2*sizeof(void*))
#define MAX_SMALL_BIN_SIZE (BINS_SIZE*sizeof(void*)+sizeof(void*))

#define update_chunk(ptr,self_size,l_size,flag) do{ \
    ptr->size=self_size; \
    ptr->last_size = (l_size & (~0x3))|flag; \
} while(0)

#define update_chunk_flag(ptr,flag) do{ \
    ptr->last_size=(ptr->last_size & (~0x3))|flag; \
} while(0)



chunk_ptr merge_chunk(chunk_ptr, chunk_ptr);
void* malloc_from_bins(int, size_t);
void install_chunk(chunk_ptr);
chunk_ptr uninstall_chunk(chunk_ptr);
int heap_init(uint8_t* base,uint32_t size);
void* malloc(size_t size);
int free(void*);




#endif /* !HEAP_H */
