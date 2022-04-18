### Jump writeup

#### 1. Please analyse code by using ida, or dynamic debugging, you can get this challenge's logic that uses the setjmp and longjmp functions to obfuscate the function flow.

#### 2. the logic of this challenge you can get is as follows:

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

#### 3. Then you will solve the inverse function:

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
	
	printf flag then you will got the flag
```

#### 4. This problem is actually a bzip2 algorithm. The key of this challenge is to analyze the setjmp and longjmp to hide the function flow, the get the correct logic flow. Lastly, you can write the decoding function to get the flag.
