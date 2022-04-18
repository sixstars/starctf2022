#include <ide.h>
#include <x86.h>

#include <stdio.h>


unsigned char ide_ata_dma_read(unsigned char idx, unsigned int lba, 
	unsigned char numsects, void* addr){
	unsigned int  drive      = ide_devices[idx].drive;
	unsigned int channel = ide_devices[idx].channel;
	uint16_t words = 256;
	uint16_t port = channels[channel].bmide;
	channels[channel].nIEN = 0;
	ide_write(channel, ATA_REG_CONTROL, 0x00);
	region_desc* prd_table = channels[channel].prd_table;
	memset(channels[channel].dma_buffer,'\xcc',numsects*words*2);
	// prepare dma
	for(int i=0;i<numsects;i++){
		prd_table[i].address = (uint32_t)&channels[channel].dma_buffer[i*2*words];
		prd_table[i].count = words*2;
		prd_table[i].end = 0;
	}
	prd_table[numsects-1].end = 0x8000;
	outl(port + 4,(uint32_t)prd_table);
	ide_write(channel, ATA_REG_DMA_CMD, ide_read(channel,ATA_REG_DMA_CMD)|1<<3);
	ide_write(channel, ATA_REG_DMA_STATUS, ide_read(channel,ATA_REG_DMA_STATUS)|4|2);//clear err intr
	// access driver
	uint16_t lba_mode = ide_access_drive(channel,drive,lba,numsects);
	ide_delay(channel);
	ide_write(channel, ATA_REG_COMMAND, lba_mode ==2 ? ATA_CMD_READ_DMA_EXT:ATA_CMD_READ_DMA);
	ide_write(channel, ATA_REG_DMA_CMD, ide_read(channel,ATA_REG_DMA_CMD)|1);
	ide_wait_irq();
	ide_write(channel,ATA_REG_DMA_CMD,ide_read(channel,ATA_REG_DMA_CMD)&(~1));
	memcpy(addr,channels[channel].dma_buffer,numsects*words*2);
	return 0;
}
unsigned char ide_ata_dma_write(unsigned char idx, unsigned int lba, 
	unsigned char numsects, void* addr){
	unsigned int  drive      = ide_devices[idx].drive;
	unsigned int channel = ide_devices[idx].channel;
	uint16_t words = 256;
	// uint16_t port = channels[channel].bmide;
	region_desc* prd_table = channels[channel].prd_table;
	channels[channel].nIEN = 0;
	ide_write(channel, ATA_REG_CONTROL, 0x00);
	ide_irq_invoked = 0;
	memcpy(channels[channel].dma_buffer,addr,numsects*words*2);
	// prepare dma
	for(int i=0;i<numsects;i++){
		prd_table[i].address = (uint32_t)&channels[channel].dma_buffer[i*2*words];
		prd_table[i].count = words*2;
		prd_table[i].end = 0;
	}
	prd_table[numsects-1].end = 0x8000;
	ide_write(channel, ATA_REG_DMA_CMD, ide_read(channel,ATA_REG_DMA_CMD)&~(1<<3));//write
	ide_write(channel, ATA_REG_DMA_STATUS, ide_read(channel,ATA_REG_DMA_STATUS)|4|2);//clear err intr
	// access driver
	uint16_t lba_mode = ide_access_drive(channel,drive,lba,numsects);
	ide_delay(channel);
	ide_write(channel, ATA_REG_COMMAND, lba_mode == 2 ? ATA_CMD_WRITE_DMA_EXT:ATA_CMD_WRITE_DMA);
	ide_write(channel, ATA_REG_DMA_CMD, ide_read(channel,ATA_REG_DMA_CMD)|1);
	ide_wait_irq();
	ide_write(channel,ATA_REG_DMA_CMD,ide_read(channel,ATA_REG_DMA_CMD)&(~1));
	return 0;
}
unsigned char ide_ata_pio_read(unsigned char idx, unsigned int lba, 
	unsigned char numsects, void* addr){
	unsigned int drive = ide_devices[idx].drive;
	unsigned int channel = ide_devices[idx].channel;
	uint16_t words = 256;
	uint16_t bus = channels[channel].base;
	uint16_t lba_mode = ide_access_drive(channel,drive,lba,numsects);
	ide_delay(channel);
	ide_write(channel, ATA_REG_COMMAND, lba_mode ==2 ? ATA_CMD_READ_PIO_EXT:ATA_CMD_READ_PIO);
	for (int i = 0; i < numsects; i++) {
	 	uint8_t err = ide_polling(channel, 1);
		if (err){
			return err;
		}
		insw(bus,addr,words);
        addr += words*2;
	}
	return 0;
}
unsigned char ide_ata_pio_write(unsigned char idx, unsigned int lba, 
	unsigned char numsects, void* addr){
	unsigned int drive = ide_devices[idx].drive;
	unsigned int channel = ide_devices[idx].channel;
	uint16_t words = 256;
	uint16_t bus = channels[channel].base;
	uint16_t lba_mode = ide_access_drive(channel,drive,lba,numsects);
	ide_delay(channel);
	ide_write(channel, ATA_REG_COMMAND, lba_mode ==2 ? ATA_CMD_WRITE_PIO_EXT:ATA_CMD_WRITE_PIO);
	for (int i = 0; i < numsects; i++) {
	 	uint8_t err = ide_polling(channel, 1);
		if (err){
			return err;
		}
		outsw(bus,addr,words);
        addr += words*2;
	}
	ide_write(channel, ATA_REG_COMMAND,lba_mode == 2?ATA_CMD_CACHE_FLUSH_EXT:ATA_CMD_CACHE_FLUSH);
	ide_polling(channel,1);
	return 0;
}
