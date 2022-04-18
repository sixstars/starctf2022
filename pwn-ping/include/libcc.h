#ifndef __INCLUDE_LIBCC_H
#define __INCLUDE_LIBCC_H
#include <types.h>

void *memset(void *s, int c, size_t n);
void *memmove(void *dst, const void *src, size_t n);
void *memcpy(void *dst, const void *src, size_t n);
int memcmp(const void *v1, const void *v2, size_t n);
int printf(const char *format, ...);
int debug(const char *format, ...);
void set_loglevel(int lv);
void abort(const char *fmt, ...);
void dumpmem(void *addr,uint32_t size);
static inline uint32_t min(uint32_t a,uint32_t b){
	return a>b?b:a;
}
static inline uint32_t max(uint32_t a,uint32_t b){
	return a>b?a:b;
}
#endif
