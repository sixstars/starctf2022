/*
 * ip.h
 * Copyright (C) 2022 user <hzshang15@gmail.com>
 *
 * Distributed under terms of the MIT license.
 */

#ifndef IP_H
#define IP_H


#define ARP 0x608
#define IP 0x8


// ARP header
typedef struct __attribute__((packed))
{
	uint16_t	ar_hrd;		// Hardware type : ethernet
	uint16_t	ar_pro;     // Protocol		 : IP
	uint8_t	ar_hln;     // Hardware size
	uint8_t	ar_pln;     // Protocal size
	uint16_t	ar_op;      // Opcode replay
	uint8_t	ar_sha[6];  // Sender MAC
	uint32_t	ar_sip;  // Sender IP
	uint8_t	ar_tha[6];  // Target mac
	uint32_t	ar_tip;  // Target IP
} ARPFrame;

typedef struct __attribute__((__packed__)) 
{
    uint8_t ihl : 4;
    uint8_t version : 4;
    uint8_t tos;
    uint16_t len;
    uint16_t id;
    uint16_t flags;
    uint8_t ttl;
    uint8_t proto;
    uint16_t csum;
    uint32_t saddr;
    uint32_t daddr;       
} IPFrame;

#define ECHO 8
#define REPLY 0

typedef struct __attribute__((packed)){
    uint8_t type;
    uint8_t code;
    uint16_t csum;
    uint8_t data[];
} ICMPFrame;



typedef struct __attribute__((__packed__)) {
    uint8_t dst_mac[6];
    uint8_t src_mac[6];
    uint16_t type;
    union {
        IPFrame ip;
        ARPFrame arp;
    };
} EthFrame;

static inline unsigned short
htons (unsigned int __arg)
{
  register unsigned short __result;

  __asm__ ("xchg%B0 %b0,%h0" : "=q" (__result) : "0" (__arg));
  return __result;
}

void network_main(uint8_t* ipbuf,int size,uint8_t* outbuf,uint32_t* outsize_ptr);
void ipmain(IPFrame* ip,int size,uint8_t* buf,uint32_t* size_ptr);
void arpmain(ARPFrame* arp,int size,uint8_t* buf,uint32_t* size_ptr);


#endif /* !IP_H */
