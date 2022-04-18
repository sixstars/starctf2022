/*
 * bitmap.h
 * Copyright (C) 2021 mac <hzshang15@gmail.com>
 *
 * Distributed under terms of the MIT license.
 */

#ifndef BITMAP_H
#define BITMAP_H

typedef unsigned char* bitmap_t;

static inline void set_bitmap(bitmap_t b, int i) {
    b[i / 8] |= 1 << (i & 7);
}

static inline void unset_bitmap(bitmap_t b, int i) {
    b[i / 8] &= ~(1 << (i & 7));
}

static inline void get_bitmap(bitmap_t b, int i) {
    return b[i / 8] & (1 << (i & 7)) ? 1 : 0;
}

static inline bitmap_t create_bitmap(int n) {
    void* buf = malloc((n + 7) / 8);
    memset(buf,0,(n + 7) / 8);
    return buf;
}

#endif /* !BITMAP_H */
