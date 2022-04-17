# 2022 *CTF010 wp - Simple File System


### Introduction

I implemented a [file system project from umn](https://www-users.cselabs.umn.edu/classes/Fall-2019/csci5103/tmp/project3/project3.html) for practice. Here is the [project source code](https://github.com/stefan1wan/Simple_FS/tree/main) and here is the [problem source code](https://github.com/stefan1wan/Simple_FS/tree/starctf_simple_fs).

###  Solution
#### keypoint
As the following code, the *plantflag* command will generate two random numbers *num1* and *num2* firstly. Then the real flag will be buried between *num1* random flags and *num2* random flags. As the logics in function `do_copyin`, for the real flag(`flag=1`), it will call `do_encode` to encrypt the flag; for the random flags(`flag=2`), it will call `do_random` to generate a random number.

```c
else if(!strcmp(cmd,"plantflag")) {
            time_t t;
            srand((unsigned) time(&t));
            int num1 = rand()%100;
            int num2 = rand()%100;
            for(int i=0; i<num1; i++){
                inumber = create_inode();
                if(do_copyin("flag",inumber, 2)) {
                        printf("copied file %s to inode %d\n",arg1,inumber);
                } else {
                        printf("copy failed!\n");
                }
            }
    
            inumber = create_inode();
            if(do_copyin("flag",inumber, 1)) {
                    printf("plant flag to inode %d!\n", inumber);
            } else {
                    printf("copy failed!\n");
            }

            for(int i=0; i<num2; i++){
                inumber = create_inode();
                if(do_copyin("flag",inumber, 2)) {
                        printf("copied file %s to inode %d\n",arg1,inumber);
                } else {
                        printf("copy failed!\n");
                }
            }
}
```

The encode function is naive
```c
static int do_encode(char* buffer, int length){
    unsigned int mask = fs_getmask();
    for(int i=0; i<length; i++){
        unsigned char* buf = &buffer[i];
        *buf  = (*buf >> 1) | (*buf << 7);
        *buf ^= (unsigned char)(mask&0xff);
        *buf  = (*buf >> 2) | (*buf << 6);
        *buf ^= (unsigned char)((mask>>8)&0xff);
        *buf  = (*buf >> 3) | (*buf << 5);
        *buf ^= (unsigned char)((mask>>16)&0xff);
        *buf  = (*buf >> 4) | (*buf << 4);
        *buf ^= (unsigned char)((mask>>24)&0xff);
        *buf  = (*buf >> 5) | (*buf << 3);
    }
    return 0;
}
```

#### solution steps

Run the following command to mount the file system.
```
./simplefs image.flag 500
mount
```

We could run `debug` to get the inodes numbers(175 or 0-174) and run `copyout <inode id> <your file>` to get all the encrypted data. Then decode all the data by a script like the following, you will find the flag.
```c
#include <stdio.h>

static int do_decode(unsigned char* buf, int length){
    unsigned int mask = 0xdeedbeef;
    for(int i=0; i<length; i++){
        unsigned char* buf = &buffer[i];
        buf[i]  = (buf[i] >> 1) | (buf[i] << 7);
        buf[i] ^= (unsigned char)(mask&0xff);
        buf[i]  = (buf[i] >> 2) | (buf[i] << 6);
        buf[i] ^= (unsigned char)((mask>>8)&0xff);
        buf[i]  = (buf[i] >> 3) | (buf[i] << 5);
        buf[i] ^= (unsigned char)((mask>>16)&0xff);
        buf[i]  = (buf[i] >> 4) | (buf[i] << 4);
        buf[i] ^= (unsigned char)((mask>>24)&0xff);
        buf[i]  = (buf[i] >> 5) | (buf[i] << 3);
    }
    return 0;
}

int main(){
    int fd = open("output_file", r);
    unsigned char* buf[10000];
    int len = read(fd, buf,100);
    do_decode(buf, len);
    printf("%s\n", buf);
}
```
