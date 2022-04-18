/*
 * heap.c
 * Copyright (C) 2018 hzshang <hzshang15@gmail.com>
 *
 * Distributed under terms of the MIT license.
 */

#include <heap.h>
#include <stdio.h>
chunk_ptr* bins_ptr = NULL;

// install chunk to bins
void install_chunk(chunk_ptr ptr) {
    size_t size = ptr->size;
    if(size == 0) {
        abort("invalid chunk!\n");
    }
    chunk_ptr* head;
    if(size > MAX_SMALL_BIN_SIZE) {
        //install to first array
        head = bins_ptr;
    } else {
        head = bins_ptr + index_of_bins(size);
    }
    ptr->next = *head;
    *head = ptr;
}
// uninstall chunk from bins
chunk_ptr uninstall_chunk(chunk_ptr ptr) {
    size_t size = ptr->size;
    if(size == 0) {
        abort("invalid chunk!\n");
    }
    int index= size > MAX_SMALL_BIN_SIZE ? 0:index_of_bins(size);
    chunk_ptr temp=bins_ptr[index];
    if(temp == ptr) {
        bins_ptr[index] = ptr->next;
    } else {
        while(temp) {
            if(temp->next == ptr)
                break;
            temp = temp->next;
        }
        if(!temp) {
            abort("no such chunk in bins!");
        } else {
            temp->next = ptr->next;
        }
    }
    return ptr;
}

// init heap
int heap_init(uint8_t* base,uint32_t size) {
    // init bins
    bins_ptr = (chunk_ptr*)base;
    memset(bins_ptr, 0, BINS_SIZE * sizeof(chunk_ptr));
    chunk_ptr first_heap = (chunk_ptr)(base + BINS_SIZE*sizeof(chunk_ptr));
    update_chunk(first_heap, size - BINS_SIZE*sizeof(chunk_ptr), 0, NOUSE);
    install_chunk(first_heap);
    return 0;
}


void* malloc(size_t size) {
    if(!size)
        return 0;
    size_t chunk_size = mem_chunk_size(size);
    chunk_ptr ret;
    int index = 0;
    if (chunk_size <= MAX_SMALL_BIN_SIZE ) {
        index = index_of_bins(chunk_size);
        if(bins_ptr[index] == 0) {
            index = 0;
        }
    }
    ret = malloc_from_bins(index, chunk_size);
    if(ret)
        return chunk_content(ret);
    else
        return NULL;
}

int free(void* dev) {
    chunk_ptr ptr = get_chunk_by_content(dev);
    update_chunk_flag(ptr,NOUSE);
    chunk_ptr up = chunk_up(ptr);
    chunk_ptr down = chunk_down(ptr);
    if(up != ptr && !chunk_is_use(up)) {
        uninstall_chunk(up);
        ptr = merge_chunk(up, ptr);
    }
    if(!chunk_is_use(down)) {
        uninstall_chunk(down);
        ptr = merge_chunk(ptr, down);
    }
    install_chunk(ptr);
    return 0;
}

//merge two free chunk
chunk_ptr merge_chunk(chunk_ptr up, chunk_ptr down) {
    size_t size = up->size + down->size;
    size_t last_size = chunk_last_size(up);
    int flag = chunk_flag(up) | chunk_flag(down);
    // init new flag
    chunk_ptr new_chunk = up;
    update_chunk(new_chunk,size,last_size,flag);
    
    // change next chunk's last size
    chunk_ptr temp = chunk_down(new_chunk);
    int temp_flag = chunk_flag(temp);
    update_chunk(temp,temp->size,size,temp_flag);
    return new_chunk;
}

void* malloc_from_bins(int index, size_t size) {
    if(!bins_ptr[index]) {
        // no more chunk
        return NULL;
    }
    chunk_ptr new_chunk;
    //index > 0 , cut from fast bins
    if(index > 0) {
        new_chunk = uninstall_chunk(bins_ptr[index]);
        update_chunk_flag(new_chunk, USE);
        return new_chunk;
    }
    //index == 0 , cut from big chunk

    //find the chunk
    chunk_ptr temp = bins_ptr[0];
    while(temp) {
        if(temp->size >= size)
            break;
        temp = temp->next;
    }

    //malloc size is too big
    if(temp == 0)
        return NULL;

    // uninstall chunk first
    chunk_ptr uninstalled_chunk = uninstall_chunk(temp);
    size_t total_size = uninstalled_chunk->size;
    size_t last_size = chunk_last_size(uninstalled_chunk);

    //init new chunk
    new_chunk = uninstalled_chunk;
    update_chunk(new_chunk,size,last_size,USE);

    size_t left_size = total_size - size;
    chunk_ptr left_chunk;
    if(left_size < 0xc) {
        //It's a memory fragmentation
        //so merge it with new_chunk
        left_chunk = new_chunk;
        left_chunk->size=total_size;
        update_chunk_flag(left_chunk,USE);
    } else { // it's still a chunk after cut
        left_chunk = chunk_down(new_chunk);
        update_chunk(left_chunk,left_size,new_chunk->size,NOUSE);
        install_chunk(left_chunk);
    }
    // change down chunk size
    chunk_ptr down = chunk_down(left_chunk);
    update_chunk(down,down->size,left_chunk->size,chunk_flag(down));
    return new_chunk;
}




