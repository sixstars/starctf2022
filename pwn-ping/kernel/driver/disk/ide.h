#ifndef __IDE_H
#define __IDE_H
#include <types.h>
#define ATA_REG_DATA       0x00
#define ATA_REG_ERROR      0x01
#define ATA_REG_FEATURES   0x01
#define ATA_REG_SECCOUNT0  0x02
#define ATA_REG_LBA0       0x03
#define ATA_REG_LBA1       0x04
#define ATA_REG_LBA2       0x05
#define ATA_REG_HDDEVSEL   0x06
#define ATA_REG_COMMAND    0x07 // for read
#define ATA_REG_STATUS     0x07 // for write
#define ATA_REG_SECCOUNT1  0x08
#define ATA_REG_LBA3       0x09
#define ATA_REG_LBA4       0x0A
#define ATA_REG_LBA5       0x0B
#define ATA_REG_CONTROL    0x0C
#define ATA_REG_ALTSTATUS  0x0C
#define ATA_REG_DEVADDRESS 0x0D

#define ATA_REG_DMA_CMD    0x0E
#define ATA_REG_DMA_STATUS 2 + 0x0E
#define ATA_REG_DMA_PRD    4 + 0x0E

#define IDE_DMA_CMD    0
#define IDE_DMA_STATUS 2
#define IDE_DMA_PRD    4

#define ATA_CMD_READ_PIO          0x20
#define ATA_CMD_READ_PIO_EXT      0x24
#define ATA_CMD_READ_DMA          0xC8
#define ATA_CMD_READ_DMA_EXT      0x25
#define ATA_CMD_WRITE_PIO         0x30
#define ATA_CMD_WRITE_PIO_EXT     0x34
#define ATA_CMD_WRITE_DMA         0xCA
#define ATA_CMD_WRITE_DMA_EXT     0x35
#define ATA_CMD_CACHE_FLUSH       0xE7
#define ATA_CMD_CACHE_FLUSH_EXT   0xEA
#define ATA_CMD_PACKET            0xA0
#define ATA_CMD_IDENTIFY_PACKET   0xA1
#define ATA_CMD_IDENTIFY          0xEC

#define      ATAPI_CMD_READ       0xA8
#define      ATAPI_CMD_EJECT      0x1B
#define      ATAPI_CMD_WRITE	  0xAA

#define ATA_IDENT_DEVICETYPE   0
#define ATA_IDENT_CYLINDERS    2
#define ATA_IDENT_HEADS        6
#define ATA_IDENT_SECTORS      12
#define ATA_IDENT_SERIAL       20
#define ATA_IDENT_MODEL        54
#define ATA_IDENT_UDMASUP      88
#define ATA_IDENT_MDMASUP      63
#define ATA_IDENT_CAPABILITIES 98
#define ATA_IDENT_FIELDVALID   106
#define ATA_IDENT_MAX_LBA      120
#define ATA_IDENT_COMMANDSETS  164
#define ATA_IDENT_MAX_LBA_EXT  200


#define IDE_ATA        0x00
#define IDE_ATAPI      0x01
 
#define ATA_MASTER     0x00
#define ATA_SLAVE      0x01

#define ATA_ER_BBK      0x80    // Bad block
#define ATA_ER_UNC      0x40    // Uncorrectable data
#define ATA_ER_MC       0x20    // Media changed
#define ATA_ER_IDNF     0x10    // ID mark not found
#define ATA_ER_MCR      0x08    // Media change request
#define ATA_ER_ABRT     0x04    // Command aborted
#define ATA_ER_TK0NF    0x02    // Track 0 not found
#define ATA_ER_AMNF     0x01    // No address mark

#define ATA_SR_BSY     0x80    // Busy
#define ATA_SR_DRDY    0x40    // Drive ready
#define ATA_SR_DF      0x20    // Drive write fault
#define ATA_SR_DSC     0x10    // Drive seek complete
#define ATA_SR_DRQ     0x08    // Data request ready
#define ATA_SR_CORR    0x04    // Corrected data
#define ATA_SR_IDX     0x02    // Index
#define ATA_SR_ERR     0x01    // Error

// Channels:
#define      ATA_PRIMARY      0x00
#define      ATA_SECONDARY    0x01
 
// Directions:
#define      ATA_READ      0x00
#define      ATA_WRITE     0x01

#define      ATAPI_READ      0x00
#define      ATAPI_WRITE     0x01


typedef struct {
	uint32_t address;
	uint16_t count;
	uint16_t end; // the bit 
} region_desc;
struct ide_device {
   unsigned char  reserved;    // 0 (Empty) or 1 (This Drive really exists).
   unsigned char  channel;     // 0 (Primary Channel) or 1 (Secondary Channel).
   unsigned char  drive;       // 0 (Master Drive) or 1 (Slave Drive).
   unsigned short type;        // 0: ATA, 1:ATAPI.
   unsigned short signature;   // Drive Signature
   unsigned short capabilities;// Features.
   unsigned int   commandSets; // Command Sets Supported.
   unsigned int   size;        // Size in Sectors.
   unsigned char  model[41];   // Model in string.
   
   uint8_t dma;
   int udma;
   int mdma;
};
struct IDEChannelRegisters {
   unsigned short base;  // I/O Base.
   unsigned short ctrl;  // Control Base
   unsigned short bmide; // Bus Master IDE
   unsigned char  nIEN;  // nIEN (No Interrupt);
   region_desc* prd_table;
   uint8_t* dma_buffer;
};
extern struct IDEChannelRegisters channels[2];

extern struct ide_device ide_devices[4];
extern unsigned char ide_irq_invoked;
extern unsigned char atapi_packet[12];
void ide_init();
void ide_wait_irq();
unsigned char ide_print_error(unsigned int drive, unsigned char err);
void ide_irq();
void ide_read_sectors(unsigned char drive, unsigned char numsects, unsigned int lba, void* addr);
void ide_write_sectors(unsigned char drive, unsigned char numsects, unsigned int lba, void* addr);
void ide_enable_dma(int idx);

void ide_disable_dma(int idx);

void ide_write(unsigned char channel, unsigned char reg, unsigned char data);
unsigned char ide_read(unsigned char channel, unsigned char reg);
void ide_delay(int channel);

unsigned char ide_atapi_access(uint8_t direction,unsigned char drive, unsigned int lba, unsigned char numsects,
          void* addr);
unsigned char ide_polling(unsigned char channel, unsigned int advanced_check);



unsigned char ide_ata_dma_read(unsigned char idx, unsigned int lba, unsigned char numsects, void* addr);
unsigned char ide_ata_dma_write(unsigned char idx, unsigned int lba, unsigned char numsects, void* addr);
unsigned char ide_ata_pio_read(unsigned char idx, unsigned int lba, unsigned char numsects, void* addr);
unsigned char ide_ata_pio_write(unsigned char idx, unsigned int lba, unsigned char numsects, void* addr);
uint8_t ide_access_drive(uint8_t channel,uint8_t drive,uint32_t lba,uint32_t numsects);

unsigned char ide_atapi_dma_read(unsigned char idx, unsigned int lba, unsigned char numsects, void* addr);
unsigned char ide_atapi_dma_write(unsigned char idx, unsigned int lba, unsigned char numsects, void* addr);
unsigned char ide_atapi_pio_read(unsigned char idx, unsigned int lba, unsigned char numsects, void* addr);
unsigned char ide_atapi_pio_write(unsigned char idx, unsigned int lba, unsigned char numsects, void* addr);
unsigned char ide_atapi_bug_trigger();


#endif




