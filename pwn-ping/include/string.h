#ifndef __INCLUDE_STRING_H__
#define __INCLUDE_STRING_H__

#include <types.h>

size_t strlen(const char *s);
size_t strnlen(const char *s, size_t len);

char *strcpy(char *dst, const char *src);
char *strncpy(char *dst, const char *src, size_t len);

int strcmp(const char *s1, const char *s2);
int strncmp(const char *s1, const char *s2, size_t n);

char *strchr(const char *s, char c);
char *strfind(const char *s, char c);
long strtol(const char *s, char **endptr, int base);



#endif /* !__INCLUDE_STRING_H__ */

