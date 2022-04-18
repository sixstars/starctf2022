#include <setjmp.h>
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <signal.h>

const char STX = '\002', ETX = '\003';
char ret[60] = {0x3, 0x6a, 0x6d, 0x47, 0x6e, 0x5f, 0x3d, 0x75, 0x61, 0x53, 0x5a, 0x4c, 0x76, 0x4e, 0x34, 0x77, 0x46, 0x78, 0x45, 0x36, 0x52, 0x2b, 0x70, 0x2, 0x44, 0x32, 0x71, 0x56, 0x31, 0x43, 0x42, 0x54, 0x63, 0x6b};
char flag[60]; 
char r[60];
int LEN = 34;
char *ss, *str;
char **table;


jmp_buf bufferA, bufferB, bufferC;

static int compareStrings(const void *a, const void *b)
{
    char *aa = *(char **)a;
    char *bb = *(char **)b;
    return strcmp(aa, bb);
}


void setionA()
{
    if (strchr(flag, STX) || strchr(flag, ETX))
    {
        longjmp(bufferA, 1);
    }
    int jmp_ret = setjmp(bufferB);
    if (jmp_ret == 0)
    {
        longjmp(bufferA, 2);
    }
    sprintf(ss, "%c%s%c", STX, flag, ETX);
    jmp_ret = setjmp(bufferB);
    if (jmp_ret == 0)
        longjmp(bufferA, 1);
    if (jmp_ret >= 1)
    {
        str = calloc(LEN + 1, sizeof(char));
        strcpy(str, ss + jmp_ret - 1);
        if (jmp_ret - 1 > 0)
            strncat(str, ss, jmp_ret - 1);
        table[jmp_ret - 1] = str;
    }
    longjmp(bufferA, jmp_ret);
}

void setionB()
{
    int jmp_ret = setjmp(bufferC);
    if (jmp_ret == 0)
    {
        longjmp(bufferA, 1);
    }
    else if (jmp_ret < LEN + 1)
    {
        r[jmp_ret - 1] = table[jmp_ret - 1][LEN - 1];
        free(table[jmp_ret - 1]);
        longjmp(bufferA, jmp_ret + 1);
    }
    if (memcmp(r, ret, LEN) == 0)
        longjmp(bufferA, 1);
    else
        longjmp(bufferA, 2);
}

int main()
{
    gets(flag);
    int i;
    int jmp_ret = setjmp(bufferA);
    if (jmp_ret == 0)
    {
        setionA();
    }
    else if (jmp_ret == 1)
    {
        return 1;
    }
    ss = calloc(LEN + 1, sizeof(char));
    jmp_ret = setjmp(bufferA);
    if (jmp_ret == 0)
        longjmp(bufferB, 1);
    table = malloc(LEN * sizeof(const char *));

    jmp_ret = setjmp(bufferA);
    if (jmp_ret == 0)
        longjmp(bufferB, 1);
    else if (jmp_ret < LEN + 1)
    {
        longjmp(bufferB, jmp_ret + 1);
    }
    qsort(table, LEN, sizeof(char*), &compareStrings);
    jmp_ret = setjmp(bufferA);
    if (jmp_ret == 0)
        setionB();
    else if (jmp_ret < LEN + 1)
    {
        longjmp(bufferC, jmp_ret);
    }
    free(table);
    free(ss);
    jmp_ret = setjmp(bufferA);
    if (jmp_ret == 0)
        longjmp(bufferC, LEN + 1);
    if (jmp_ret == 1)
        printf("*CTF{%s}\n", flag);
    return 0;
}
