

inline uint32_t ROR(uint32_t x, int r)
{
    return (x >> r) | (x << (32 - r));
}

inline uint32_t ROL(uint32_t x, int r)
{
    return (x << r) | (x >> (32 - r));
}

static  char key[16] = {0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0xA, 0xB, 0xC, 0xD, 0xE, 0xF};

static void encipher(unsigned int num_rounds, uint32_t *v)
{
    uint32_t *key2 = (uint32_t *)key;
    unsigned int i;
    uint32_t v0 = v[0], v1 = v[1], sum = 0, delta = 0x10325476;
    for (i = 0; i < num_rounds; i++)
    {
        v0 += (((v1 << 4) ^ (v1 >> 5)) + v1) ^ (sum + key2[sum & 3]);
        sum += delta;
        v1 += (((v0 << 4) ^ (v0 >> 5)) + v0) ^ (sum + key2[(sum >> 11) & 3]);
    }
    v[0] = v0;
    v[1] = v1;
}


// Convert words (input) into bytes.
static uint32_t Words32ToBytes(uint32_t value)
{
    return ((uint32_t)((uint8_t)value) << 24 | (uint32_t)(uint8_t)(value >> 8) << 16 |
            (uint32_t)(uint8_t)(value >> 16) << 8 | (uint32_t)(uint8_t)(value >> 24));
}
// Covert bytes into words (output).
static uint32_t BytesToWords32(uint32_t value)
{

    return ((uint32_t)(uint8_t)(value >> 24) | (uint32_t)(uint8_t)(value >> 16) << 8 |
            (uint32_t)(uint8_t)(value >> 8) << 16 | (uint32_t)((uint8_t)value) << 24);
}
static uint32_t roundKeys[60];

// function that generates subKeys from the key according to the SIMON key scheduling algorithm for a 128-bit key
static uint32_t *generateSubkeys()
{
    // the 128 bit key is placed in two integers, both of them are 64 bit
    uint64_t KeyHigh = *((uint64_t *)key);
    uint64_t KeyLow = *((uint64_t *)&key[8]);
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

// function for encrypting a block using a key
 static uint64_t encrypt(char *plaintext)
{
    // generate the subkeys using the function defined above
    uint32_t *roundKeys = generateSubkeys();
    // convert the plaintext from a Hex String to a 64-bit integer
    uint64_t state = *((uint64_t *)plaintext);
    // split block of plain text into 2 blocks.
    uint32_t rightPlainBlock = Words32ToBytes(state >> 32);
    uint32_t leftPlainBlock = Words32ToBytes(state);

    for (int i = 0; i < 44; i++)
    {

        uint32_t temp = leftPlainBlock;
        leftPlainBlock = rightPlainBlock ^ ((ROL(leftPlainBlock, 1) & ROL(leftPlainBlock, 8)) ^ ROL(leftPlainBlock, 2)) ^ roundKeys[i];
        rightPlainBlock = temp;
    }

    state = state & 0;
    state = ((state | leftPlainBlock) << 32) | rightPlainBlock;

    return state;
}


static char result[32] = {0x66, 0xc2, 0xf5, 0xfd, 0x86, 0x82, 0x32, 0x7a, 0x04, 0x40, 0x94, 0xce, 0xdc, 0x8a, 0xe0, 0x5d, 0x0a, 0xbd, 0xe4, 0xa6, 0xdc, 0xad, 0xca, 0x16, 0x0c, 0x6f, 0xcd, 0x13, 0x36, 0xd9, 0x75, 0x1a};

static char enc_data[32];
// Test main function
int sfimain(char *flag)
{
    char *plaintext;
    int i = 0;
    
    for (i = 0; i < 4; i++)
    {
        uint64_t ciphertext;
        plaintext = flag + i * 8;
        ciphertext = encrypt(plaintext);
        encipher(1 << (i + 1), (uint32_t *)&ciphertext);
        memcpy(enc_data+8*i,(char*)&ciphertext,8);
    }
    if (memcmp(result, enc_data, 32) == 0)
    {
        return 1;
    }
    return 0;
}
