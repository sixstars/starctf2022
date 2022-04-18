/*
 * physical_page.c
 * Copyright (C) 2021 mac <hzshang15@gmail.com>
 *
 * Distributed under terms of the MIT license.
 */
#include <types.h>
#include <libcc.h>
#include <physical_page.h>
static uint8_t* page_ptr;

uint8_t* physical_alloc(uint32_t size,uint32_t align){
    size = (size+0xfff)&(~0xfff);
    page_ptr = (uint8_t*)(((uint32_t)page_ptr + align - 1)&~(align-1));
    uint8_t* page = page_ptr;
    page_ptr += size;
    *page = 1;
    if(*page != 1){
        debug("!!!!!!!!!!!!!page memory alloc fail!!!!!!!!!!!!\n");
    }
    memset(page,0,size);
    return page;
}


void physical_page_init(uint8_t* page,uint32_t size){
//    uint32_t pool_size = size >> (12 + 3);
//    page_pool = create_bitmap(pool_size);
    page_ptr = page;
    *page = 1;
    if(*page != 1){
        debug("!!!!!!!!!!page memory not available\n!!!!!!!!!!!!!!!");
    }
}







