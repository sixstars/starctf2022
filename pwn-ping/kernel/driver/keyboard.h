/*
 * keyboard.h
 * Copyright (C) 2021 mac <hzshang15@gmail.com>
 *
 * Distributed under terms of the MIT license.
 */

#ifndef KEYBOARD_H
#define KEYBOARD_H
#include <types.h>
#include <trap.h>
uint8_t get_c();
void keyboard_callback(struct trapframe* tf);
void keyboard_init();
#endif /* !KEYBOARD_H */
