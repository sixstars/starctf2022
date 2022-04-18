### Jump writeup

#### 1. 使用ida通过对代码的分析，或者使用相关库函数识别的工具，会发现本题是使用setjmp和longjmp函数对逻辑的函数流进行混淆的无符号binary。

#### 2. 逆向获得程序逻辑大概如下

```C
    len = strlen(flag) + 2;
    ss = calloc(len + 1, sizeof(char));
    sprintf(ss, "%c%s%c", "\002", flag, "\003");
    table = malloc(len * sizeof(const char *));
    for (i = 0; i < len; ++i)
    {
        str = calloc(len + 1, sizeof(char));
        strcpy(str, ss + i);
        if (i > 0)
            strncat(str, ss, i);
        table[i] = str;
    }
    qsort(table, len);
    for (i = 0; i < len; ++i)
    {
        r[i] = table[i][len - 1];
        free(table[i]);
    }

	r[] == ?{0x3,0x6a,0x6d,0x47,0x6e,0x5f,0x3d,0x75,0x61,0x53,0x5a,0x4c,0x76,0x4e,0x34,0x77,0x46,0x78,0x45,0x36,0x52,0x2b,0x70,0x2,0x44,0x32,0x71,0x56,0x31,0x43,0x42,0x54,0x63,0x6b}
```

#### 3. 写出简单的逆函数

```C
    len = strlen(r);
    for (i = 0; i < len; ++i)
        table[i] = calloc(len + 1, sizeof(char));
    for (i = 0; i < len; ++i)
    {
        for (j = 0; j < len; ++j)
        {
            memmove(table[j] + 1, table[j], len);
            table[j][0] = r[j];
        }
        qsort(table, len);
    }
    for (i = 0; i < len; ++i)
    {
        if (table[i][len - 1] == "\003")
        {
            strncpy(flag, table[i] + 1, len - 2);
            break;
        }
    }
	
	printf flag 便可得到flag
```

#### 4. 本题其实是一个bzip2算法，关键在于分析setjmp和longjmp对函数流的隐藏，分析出正确的逻辑流，然后就可以写出解码函数，获得flag。