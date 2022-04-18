#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include <time.h>
#include <unistd.h>

#include <sys/types.h> 
#include <sys/stat.h>
#include <fcntl.h>
/* #define ED25519_DLL */
#include "src/ed25519.h"

#include "src/ge.h"
#include "src/sc.h"
const char base[] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="; 
char *base64(const char* data, int data_len) 
{ 
    //int data_len = strlen(data); 
    int prepare = 0; 
    int ret_len; 
    int temp = 0; 
    char *ret = NULL; 
    char *f = NULL; 
    int tmp = 0; 
    char changed[4]; 
    int i = 0; 
    ret_len = data_len / 3; 
    temp = data_len % 3; 
    if (temp > 0) 
    { 
        ret_len += 1; 
    } 
    ret_len = ret_len*4 + 1; 
    ret = (char *)malloc(ret_len); 

    if ( ret == NULL) 
    { 
        printf("No enough memory.\n"); 
        exit(0); 
    } 
    memset(ret, 0, ret_len); 
    f = ret; 
    while (tmp < data_len) 
    { 
        temp = 0; 
        prepare = 0; 
        memset(changed, '\0', 4); 
        while (temp < 3) 
        { 
            //printf("tmp = %d\n", tmp); 
            if (tmp >= data_len) 
            { 
                break; 
            } 
            prepare = ((prepare << 8) | (data[tmp] & 0xFF)); 
            tmp++; 
            temp++; 
        } 
        prepare = (prepare<<((3-temp)*8)); 
        //printf("before for : temp = %d, prepare = %d\n", temp, prepare); 
        for (i = 0; i < 4 ;i++ ) 
        { 
            if (temp < i) 
            { 
                changed[i] = 0x40; 
            } 
            else 
            { 
                changed[i] = (prepare>>((3-i)*6)) & 0x3F; 
            } 
            *f = base[changed[i]]; 
            //printf("%.2X", changed[i]); 
            f++; 
        } 
    } 
    *f = '\0'; 
    return ret; 
} 

int main() {
    setbuf(stdout,NULL);
    setbuf(stdin,NULL);
	unsigned char public_key[32]={0}, private_key[64]={0}, scalar[32]={0};
	unsigned char signature[64]={0};
	unsigned char message[1024]={0};
	int fd=open("flag",O_RDONLY);
	int fd2=open("nonce",O_RDONLY);
	read(fd,private_key,31);
	read(fd2,private_key+32,32);
	ed25519_create_keypair(public_key, private_key);
	printf("public_key:%s\n",base64(public_key,32));
	puts("input message:");
	read(0,message,1024);
	puts("input scalar:");
	read(0,scalar,32);
	ed25519_add_scalar(public_key, private_key, scalar);
	ed25519_sign_add_scalar(signature, message, 1024, public_key, private_key,scalar);
	printf("signature:%s\n",base64(signature,64));

}
