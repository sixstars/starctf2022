### NaCl writeup

#### 1. main logic in `src.c`
#### 2. this challenge is likely Google Native NaCl project. It includes jump check, memory check and align check. But I alse replace the stack registers, use r15 to replace rsp, r14 to replace rbp. And there are no call, ret and leave. Here is only jmp. Access memory by addding r13 to get address. 
#### 3. i put the loader function and the native client binary together. the section SFI is NaCl's code, and SFI_DATA is data section. 
#### 4. expected solution is as follows: firstly, dump section SFI and SFI_data. Then, you modify the binary by script code. you need to recover stack, register and instrcution call, ret, etc. Lastly, you can disassemble the new binary by ida or other tools. Then you can got simple logic.
#### 5. i am sorry that all players solved this challenge by dynamic debugging and static Analysis of orginal assembly code. Because my hide logic code is short. So bad.
#### 6. decoding logic is as follows:
```C
#include <stdio.h>
#include <string.h>

#define uint32_t unsigned int
#define uint64_t unsigned long long
#define uint8_t unsigned char
#define ROR(x, r) ((x >> r) | (x << (32 - r)))
#define ROL(x, r) ((x << r) | (x >> (32 - r)))


unsigned char key[16] = {0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xA, 0xB, 0xC, 0xD, 0xE, 0xF};


void decipher(unsigned int num_rounds, uint32_t *v)
{
    uint32_t *key2 = (uint32_t *)key;
    unsigned int i;
    uint32_t v0 = v[0], v1 = v[1], delta = 0x10325476, sum = delta * num_rounds;
    for (i = 0; i < num_rounds; i++)
    {
        v1 -= (((v0 << 4) ^ (v0 >> 5)) + v0) ^ (sum + key2[(sum >> 11) & 3]);
        sum -= delta;
        v0 -= (((v1 << 4) ^ (v1 >> 5)) + v1) ^ (sum + key2[sum & 3]);
    }
    v[0] = v0;
    v[1] = v1;
}

// Convert words (input) into bytes.
uint32_t Words32ToBytes(uint32_t value)
{
    return ((uint32_t)((uint8_t)value) << 24 | (uint32_t)(uint8_t)(value >> 8) << 16 |
            (uint32_t)(uint8_t)(value >> 16) << 8 | (uint32_t)(uint8_t)(value >> 24));
}
// Covert bytes into words (output).
uint32_t BytesToWords32(uint32_t value)
{

    return ((uint32_t)(uint8_t)(value >> 24) | (uint32_t)(uint8_t)(value >> 16) << 8 |
            (uint32_t)(uint8_t)(value >> 8) << 16 | (uint32_t)((uint8_t)value) << 24);
}
uint32_t roundKeys[60];

// function that generates subKeys from the key according to the SIMON key scheduling algorithm for a 128-bit key
uint32_t *generateSubkeys()
{
    // the 128 bit key is placed in two integers, both of them are 64 bit
    uint64_t KeyHigh = *((uint64_t *)key);
    uint64_t KeyLow = *((uint64_t *)&key[8]);
    // 0x67452301，0xEFCDAB89，0x98BADCFE，0x10325476
    // 0xfc2ce51207a635dbLL
    //  uint32_t c = 0xfffffffc;
    uint64_t z3 = 0x67452301EFCDAB89LL;
    uint32_t c = 0x98BADCFE;

    // we allocate space for 32 subkeys, since there are 32 rounds

    roundKeys[0] = Words32ToBytes(KeyHigh >> 32);
    roundKeys[1] = Words32ToBytes(KeyHigh);
    roundKeys[2] = Words32ToBytes(KeyLow >> 32);
    roundKeys[3] = Words32ToBytes(KeyLow);

    for (int i = 4; i < 44; ++i)
    {
        uint32_t test = ROR(roundKeys[i - 1], 3);
        roundKeys[i] = c ^ (z3 & 1) ^ roundKeys[i - 4] ^ ROR(roundKeys[i - 1], 3) ^ roundKeys[i - 3] ^ ROR(roundKeys[i - 1], 4) ^ ROR(roundKeys[i - 3], 1);
        z3 >>= 1;
    }
    return roundKeys;
}

// function for decrypting a block using a key
uint64_t decrypt(uint64_t ciphertext)
{
    // generate the subkeys using the function defined above
    uint32_t *roundKeys = generateSubkeys();
    // convert the plaintext from a Hex String to a 64-bit integer
    uint64_t state = ciphertext;
    // split block of plain text into 2 blocks.
    uint32_t rightCipherBlock = state;
    uint32_t leftCipherBlock = state >> 32;

    for (int i = 43; i >= 0; i--)
    {
        uint32_t temp = rightCipherBlock;
        rightCipherBlock =
            leftCipherBlock ^ ((ROL(rightCipherBlock, 1) & ROL(rightCipherBlock, 8)) ^ ROL(rightCipherBlock, 2)) ^ roundKeys[i];
        leftCipherBlock = temp;
    }

    state = state & 0;
    state = ((state | BytesToWords32(rightCipherBlock)) << 32) | BytesToWords32(leftCipherBlock);

    return state;
}

// Test main function
int main()
{
    char ret[32] = {0x66,0xc2,0xf5,0xfd,0x86,0x82,0x32,0x7a,0x04,0x40,0x94,0xce,0xdc,0x8a,0xe0,0x5d,0x0a,0xbd,0xe4,0xa6,0xdc,0xad,0xca,0x16,0x0c,0x6f,0xcd,0x13,0x36,0xd9,0x75,0x1a};
    int i = 0;
    for (i = 0; i < 4; i++)
    {
        uint64_t ciphertextv = *ret;
        decipher(1<<(i+1), (uint32_t *)&ciphertext);
        ciphertext = decrypt(ciphertext);
        printf("after decrypt %llx\n", ciphertext);
    }

    return 0;
}
```
