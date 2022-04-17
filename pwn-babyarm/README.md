## babyarm

> It is so simple, simplest, cannot be simpler…
> 
> solved: 5

Just a baby kernel stack overflow challenge. The novelty is that there are few chals about aarch64 kernel exploit before.

This arm64 kernel is built with nokaslr, stack-protector, PXN. We can simply leak the canary cookie and do ROP to leverage privilege and return to user mode.

As for how to switch back to user mode, I use gadgets of `ret_to_user` below in arch/arm64/kernel/entry.S:406

```
=> 0xffff800008012024 <ret_to_user+112>:   msr     elr_el1, x21
   0xffff800008012028 <ret_to_user+116>:   msr     spsr_el1, x22
   0xffff80000801202c <ret_to_user+120>:   ldp     x0, x1, [sp]
   0xffff800008012030 <ret_to_user+124>:   ldp     x2, x3, [sp, #16]
   0xffff800008012034 <ret_to_user+128>:   ldp     x4, x5, [sp, #32]
   0xffff800008012038 <ret_to_user+132>:   ldp     x6, x7, [sp, #48]
   0xffff80000801203c <ret_to_user+136>:   ldp     x8, x9, [sp, #64]
   0xffff800008012040 <ret_to_user+140>:   ldp     x10, x11, [sp, #80]
   0xffff800008012044 <ret_to_user+144>:   ldp     x12, x13, [sp, #96]
   0xffff800008012048 <ret_to_user+148>:   ldp     x14, x15, [sp, #112]
   0xffff80000801204c <ret_to_user+152>:   ldp     x16, x17, [sp, #128]
   0xffff800008012050 <ret_to_user+156>:   ldp     x18, x19, [sp, #144]
   0xffff800008012054 <ret_to_user+160>:   ldp     x20, x21, [sp, #160]
   0xffff800008012058 <ret_to_user+164>:   ldp     x22, x23, [sp, #176]
   0xffff80000801205c <ret_to_user+168>:   ldp     x24, x25, [sp, #192]
   0xffff800008012060 <ret_to_user+172>:   ldp     x26, x27, [sp, #208]
   0xffff800008012064 <ret_to_user+176>:   ldp     x28, x29, [sp, #224]
```
