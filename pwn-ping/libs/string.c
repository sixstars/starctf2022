#include <string.h>

#include <x86.h>

/* *
 * strlen - calculate the length of the string @s, not including
 * the terminating '\0' character.
 * @s:        the input string
 *
 * The strlen() function returns the length of string @s.
 * */
size_t
strlen(const char *s) {
    size_t cnt = 0;
    while (*s ++ != '\0') {
        cnt ++;
    }
    return cnt;
}

/* *
 * strnlen - calculate the length of the string @s, not including
 * the terminating '\0' char acter, but at most @len.
 * @s:        the input string
 * @len:    the max-length that function will scan
 *
 * Note that, this function looks only at the first @len characters
 * at @s, and never beyond @s + @len.
 *
 * The return value is strlen(s), if that is less than @len, or
 * @len if there is no '\0' character among the first @len characters
 * pointed by @s.
 * */
size_t
strnlen(const char *s, size_t len) {
    size_t cnt = 0;
    while (cnt < len && *s ++ != '\0') {
        cnt ++;
    }
    return cnt;
}

/* *
 * strcpy - copies the string pointed by @src into the array pointed by @dst,
 * including the terminating null character.
 * @dst:    pointer to the destination array where the content is to be copied
 * @src:    string to be copied
 *
 * The return value is @dst.
 *
 * To avoid overflows, the size of array pointed by @dst should be long enough to
 * contain the same string as @src (including the terminating null character), and
 * should not overlap in memory with @src.
 * */
char *
strcpy(char *dst, const char *src) {
#ifdef __HAVE_ARCH_STRCPY
    return __strcpy(dst, src);
#else
    char *p = dst;
    while ((*p ++ = *src ++) != '\0')
        /* nothing */;
    return dst;
#endif /* __HAVE_ARCH_STRCPY */
}

/* *
 * strncpy - copies the first @len characters of @src to @dst. If the end of string @src
 * if found before @len characters have been copied, @dst is padded with '\0' until a
 * total of @len characters have been written to it.
 * @dst:    pointer to the destination array where the content is to be copied
 * @src:    string to be copied
 * @len:    maximum number of characters to be copied from @src
 *
 * The return value is @dst
 * */
char *
strncpy(char *dst, const char *src, size_t len) {
    char *p = dst;
    while (len > 0) {
        if ((*p = *src) != '\0') {
            src ++;
        }
        p ++, len --;
    }
    return dst;
}

/* *
 * strcmp - compares the string @s1 and @s2
 * @s1:        string to be compared
 * @s2:        string to be compared
 *
 * This function starts comparing the first character of each string. If
 * they are equal to each other, it continues with the following pairs until
 * the characters differ or until a terminanting null-character is reached.
 *
 * Returns an integral value indicating the relationship between the strings:
 * - A zero value indicates that both strings are equal;
 * - A value greater than zero indicates that the first character that does
 *   not match has a greater value in @s1 than in @s2;
 * - And a value less than zero indicates the opposite.
 * */
int
strcmp(const char *s1, const char *s2) {
#ifdef __HAVE_ARCH_STRCMP
    return __strcmp(s1, s2);
#else
    while (*s1 != '\0' && *s1 == *s2) {
        s1 ++, s2 ++;
    }
    return (int)((unsigned char)*s1 - (unsigned char)*s2);
#endif /* __HAVE_ARCH_STRCMP */
}

/* *
 * strncmp - compares up to @n characters of the string @s1 to those of the string @s2
 * @s1:        string to be compared
 * @s2:        string to be compared
 * @n:        maximum number of characters to compare
 *
 * This function starts comparing the first character of each string. If
 * they are equal to each other, it continues with the following pairs until
 * the characters differ, until a terminating null-character is reached, or
 * until @n characters match in both strings, whichever happens first.
 * */
int
strncmp(const char *s1, const char *s2, size_t n) {
    while (n > 0 && *s1 != '\0' && *s1 == *s2) {
        n --, s1 ++, s2 ++;
    }
    return (n == 0) ? 0 : (int)((unsigned char)*s1 - (unsigned char)*s2);
}

/* *
 * strchr - locates first occurrence of character in string
 * @s:        the input string
 * @c:        character to be located
 *
 * The strchr() function returns a pointer to the first occurrence of
 * character in @s. If the value is not found, the function returns 'NULL'.
 * */
char *
strchr(const char *s, char c) {
    while (*s != '\0') {
        if (*s == c) {
            return (char *)s;
        }
        s ++;
    }
    return NULL;
}

/* *
 * strfind - locates first occurrence of character in string
 * @s:        the input string
 * @c:        character to be located
 *
 * The strfind() function is like strchr() except that if @c is
 * not found in @s, then it returns a pointer to the null byte at the
 * end of @s, rather than 'NULL'.
 * */
char *
strfind(const char *s, char c) {
    while (*s != '\0') {
        if (*s == c) {
            break;
        }
        s ++;
    }
    return (char *)s;
}

/* *
 * strtol - converts string to long integer
 * @s:        the input string that contains the representation of an integer number
 * @endptr:    reference to an object of type char *, whose value is set by the
 *             function to the next character in @s after the numerical value. This
 *             parameter can also be a null pointer, in which case it is not used.
 * @base:    x
 *
 * The function first discards as many whitespace characters as necessary until
 * the first non-whitespace character is found. Then, starting from this character,
 * takes as many characters as possible that are valid following a syntax that
 * depends on the base parameter, and interprets them as a numerical value. Finally,
 * a pointer to the first character following the integer representation in @s
 * is stored in the object pointed by @endptr.
 *
 * If the value of base is zero, the syntax expected is similar to that of
 * integer constants, which is formed by a succession of:
 * - An optional plus or minus sign;
 * - An optional prefix indicating octal or hexadecimal base ("0" or "0x" respectively)
 * - A sequence of decimal digits (if no base prefix was specified) or either octal
 *   or hexadecimal digits if a specific prefix is present
 *
 * If the base value is between 2 and 36, the format expected for the integral number
 * is a succession of the valid digits and/or letters needed to represent integers of
 * the specified radix (starting from '0' and up to 'z'/'Z' for radix 36). The
 * sequence may optionally be preceded by a plus or minus sign and, if base is 16,
 * an optional "0x" or "0X" prefix.
 *
 * The strtol() function returns the converted integral number as a long int value.
 * */
long
strtol(const char *s, char **endptr, int base) {
    int neg = 0;
    long val = 0;

    // gobble initial whitespace
    while (*s == ' ' || *s == '\t') {
        s ++;
    }

    // plus/minus sign
    if (*s == '+') {
        s ++;
    }
    else if (*s == '-') {
        s ++, neg = 1;
    }

    // hex or octal base prefix
    if ((base == 0 || base == 16) && (s[0] == '0' && s[1] == 'x')) {
        s += 2, base = 16;
    }
    else if (base == 0 && s[0] == '0') {
        s ++, base = 8;
    }
    else if (base == 0) {
        base = 10;
    }

    // digits
    while (1) {
        int dig;

        if (*s >= '0' && *s <= '9') {
            dig = *s - '0';
        }
        else if (*s >= 'a' && *s <= 'z') {
            dig = *s - 'a' + 10;
        }
        else if (*s >= 'A' && *s <= 'Z') {
            dig = *s - 'A' + 10;
        }
        else {
            break;
        }
        if (dig >= base) {
            break;
        }
        s ++, val = (val * base) + dig;
        // we don't properly detect overflow!
    }

    if (endptr) {
        *endptr = (char *) s;
    }
    return (neg ? -val : val);
}


