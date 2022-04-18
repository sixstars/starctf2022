/*
 * keyboard.c
 * Copyright (C) 2021 mac <hzshang15@gmail.com>
 *
 * Distributed under terms of the MIT license.
 */

#include <keyboard.h>
#include <x86.h>
#include <screen.h>
#include <picirq.h>
#include <mp.h>
#include <libcc.h>
#include <trap.h>
#define BUFFER_SIZE 0x100
#define COMMAND_PORT 0x64
#define DATA_PORT 0x60

uint8_t keybuffer[BUFFER_SIZE];
volatile uint32_t buffer_read;
volatile uint32_t buffer_write;
uint8_t unshift_map[0x100] = {
[0xc]'-',[0xd]'=',
[0x1a]'[',[0x1b]']',
[0x27]';',[0x28]'\'',
[0x02]'1',[0x03]'2',[0x04]'3',[0x05]'4',[0x06]'5',[0x07]'6',[0x08]'7',
[0x09]'8',[0x0A]'9',[0x0B]'0',[0x10]'q',[0x11]'w',[0x12]'e',[0x13]'r',
[0x14]'t',[0x15]'z',[0x16]'u',[0x17]'i',[0x18]'o',[0x19]'p',[0x1E]'a',
[0x1F]'s',[0x20]'d',[0x21]'f',[0x22]'g',[0x23]'h',[0x24]'j',[0x25]'k',
[0x26]'l',[0x2C]'y',[0x2D]'x',[0x2E]'c',[0x2F]'v',[0x30]'b',[0x31]'n',
[0x32]'m',[0x33]',',[0x34]'.',[0x35]'/',[0x1c]'\n',[0x39]' ',
};
uint8_t shift_map[0x100] = {
[0xc]'_',[0xd]'+',
[0x1a]'{',[0x1b]'}',
[0x27]':',[0x28]'"',
[0x02]'!',[0x03]'@',[0x04]'#',[0x05]'$',[0x06]'%',[0x07]'^',
[0x08]'&',[0x09]'*',[0x0A]'(',[0x0B]')',[0x10]'Q',[0x11]'W',
[0x12]'E',[0x13]'R',[0x14]'T',[0x15]'Z',[0x16]'U',[0x17]'I',
[0x18]'O',[0x19]'P',[0x1E]'A',[0x1F]'S',[0x20]'D',[0x21]'F',
[0x22]'G',[0x23]'H',[0x24]'J',[0x25]'K',[0x26]'L',[0x2C]'Y',
[0x2D]'X',[0x2E]'C',[0x2F]'V',[0x30]'B',[0x31]'N',[0x32]'M',
[0x33]'<',[0x34]'>',[0x35]'?',};
void keyboard_init(){
    buffer_read = 0;
    buffer_write = 0;
    outb(COMMAND_PORT,0xae);
    outb(COMMAND_PORT,0x20);
    uint8_t status = (inb(DATA_PORT) | 1) & ~0x10;
    outb(COMMAND_PORT,0x60);
    outb(DATA_PORT,status);
    outb(DATA_PORT,0xf4);

    pic_enable(IRQ_KBD);
    register_intr_handler(IRQ_KBD + IRQ_OFFSET,keyboard_callback);
}

//     // search
//     int pin = getIRQPin("ISA ",IRQ_KBD);
//     debug("keyboard pin number: %d\n",pin);
//     if(pin != -1)
//         registerIRQ(0,pin,keyboard_callback);
// }

bool shift = false;
void keyboard_callback(struct trapframe* tf){
    uint8_t value = inb(DATA_PORT);
//    kprintf("keyboard callback,recv char: %x\n",value);
    switch(value){
        case 0x2a:
        case 0x36:
            shift = true;
            break;
        case 0xaa:
        case 0xb6:
            shift = false;
            break;
        default:
            if(value <= 0x80){
                uint8_t c;
                if(shift){
                    c = shift_map[value];
                }else{
                    c = unshift_map[value];
                }
                put_c(c);
                keybuffer[(buffer_write++)%BUFFER_SIZE] = c;
            }
            break;
    }
    //TODO: add lock
}

uint8_t get_c(){
    while(buffer_read == buffer_write){
        stop_cpu();
    }
    return keybuffer[(buffer_read++)%BUFFER_SIZE];
}






