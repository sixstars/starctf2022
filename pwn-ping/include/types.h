#ifndef _INCLUDE_TYPES_H
#define _INCLUDE_TYPES_H

#define NULL ((void *)0)

typedef unsigned char		uint8_t;
typedef signed char			int8_t;
typedef unsigned short		uint16_t;
typedef signed short		int16_t;
typedef unsigned int		uint32_t;
typedef signed int		int32_t;
typedef unsigned long long	uint64_t;
typedef signed long long	int64_t;
typedef unsigned int        bool;
typedef unsigned int        uint;
/* *
 * Pointers and addresses are 32 bits long.
 * We use pointer types to represent addresses,
 * uintptr_t to represent the numerical values of addresses.
 * */

typedef int32_t intptr_t;
typedef uint32_t uintptr_t;
typedef __SIZE_TYPE__ size_t;

#define false 0
#define true 1

#endif
