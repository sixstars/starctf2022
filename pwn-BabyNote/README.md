# BabyNote

category: Pwn
solved: 17
final score: 555 pt.

## Bug

The bug is in deleteNote function.

```C
        if (list_head == n && list_head->next == NULL)
        {
            list_head = NULL;
        }
        else if (n->next != NULL)
        {
            struct node **p = &list_head;
            while (*p != n)
            {
                p = &(*p)->next;
            }
            *p = n->next;
        }
```

When there are more than 1 elements (>= 2) in the linked list, deleting the first note will not clean its reference ptr. It is a typical UAF bug.

## solution

First of all, many players have problems in the local debugging environment. You can refer to environment scaffolding in Dockerfile for this question. The path of musl libc which is `/usr/lib/x86_64-linux-musl/libc.so` is different from glibc.
(sorry about this)


1. Use UAF to leak heap addres.
2. calculate the  libc base, stdout, system and other address with heap address.
3. alloc a node as a note to fake a new node whose note point to the secret in `__malloc_context`, and leak it.
4. prepare a fake chunk and a fake store
5. prepare fake meta_area and inject it into malloc_context
6. change mem pointer and rewrite malloc_replaced
7. change mem pointer and rewrite stdout_write
8. enjoy your shell :)

Some offset may be unstable ,but this won't affect much.