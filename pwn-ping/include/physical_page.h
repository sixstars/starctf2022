/*
 * physical_page.h
 * Copyright (C) 2021 mac <hzshang15@gmail.com>
 *
 * Distributed under terms of the MIT license.
 */

#ifndef PHYSICAL_PAGE_H
#define PHYSICAL_PAGE_H

void physical_page_init(uint8_t* page,uint32_t size);
uint8_t* physical_alloc(uint32_t size,uint32_t align);
#define PAGE_COUNT(x) ((x+0xFFF)>>12)

#endif /* !PHYSICAL_PAGE_H */
