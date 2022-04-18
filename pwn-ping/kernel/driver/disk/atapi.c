#include <ide.h>
#include <x86.h>
#include <stdio.h>
#include <physical_page.h>

unsigned char ide_atapi_dma_read(unsigned char idx, unsigned int lba,
                                 unsigned char numsects, void* addr) {
	unsigned int channel = ide_devices[idx].channel;
	unsigned int drive = ide_devices[idx].drive;
	unsigned int words = 2048 / 2; // Sector Size in Words, Almost All ATAPI Drives has a sector size of 2048 bytes.
	uint8_t err;
	unsigned int bus = channels[channel].base;
	ide_write(channel, ATA_REG_CONTROL, channels[channel].nIEN = ide_irq_invoked = 0x0);
	ide_write(channel, ATA_REG_HDDEVSEL, drive << 4);
	ide_write(channel, ATA_REG_FEATURES, 1); // PIO mode
	ide_write(channel, ATA_REG_LBA1, 0);   // Lower Byte of Sector Size.
	ide_write(channel, ATA_REG_LBA2, 0);   // Upper Byte of Sector Size.
	ide_write(channel, ATA_REG_COMMAND, ATA_CMD_PACKET);      // Send the Command.
	atapi_packet[ 0] = ATAPI_CMD_READ;
	atapi_packet[ 1] = 0x0;
	atapi_packet[ 2] = (lba >> 24) & 0xFF;
	atapi_packet[ 3] = (lba >> 16) & 0xFF;
	atapi_packet[ 4] = (lba >> 8) & 0xFF;
	atapi_packet[ 5] = (lba >> 0) & 0xFF;
	atapi_packet[ 6] = 0x0;
	atapi_packet[ 7] = 0x0;
	atapi_packet[ 8] = 0x0;
	atapi_packet[ 9] = numsects;
	atapi_packet[10] = 0x0;
	atapi_packet[11] = 0x0;
	err = ide_polling(channel, 1);
	if (err) return err;
	region_desc* prd_table = channels[channel].prd_table;
	for (int i = 0; i < numsects; i++) {
		prd_table[i].address = (uint32_t)&channels[channel].dma_buffer[i*words*2];
		prd_table[i].count = words * 2;
		prd_table[i].end = 0;
	}
	prd_table[numsects - 1].end = 0x8000;
	outl(channels[channel].bmide + IDE_DMA_PRD,(uint32_t)prd_table);
	ide_write(channel, ATA_REG_DMA_CMD, ide_read(channel, ATA_REG_DMA_CMD)|(1<<3));
	ide_write(channel, ATA_REG_DMA_STATUS, ide_read(channel, ATA_REG_DMA_STATUS) | 4 | 2);
	outsw(bus, atapi_packet, 6);
	ide_write(channel, ATA_REG_DMA_CMD, ide_read(channel, ATA_REG_DMA_CMD) | 1);
	ide_wait_irq();
	ide_write(channel, ATA_REG_DMA_CMD, ide_read(channel, ATA_REG_DMA_CMD) & (~1));
	memcpy(addr,channels[channel].dma_buffer,numsects*words*2);
	return 0;
}

unsigned char ide_atapi_dma_write(unsigned char idx, unsigned int lba,
                                  unsigned char numsects, void* addr) {
	unsigned int channel = ide_devices[idx].channel;
	unsigned int drive = ide_devices[idx].drive;
	unsigned int words = 2048 / 2; // Sector Size in Words, Almost All ATAPI Drives has a sector size of 2048 bytes.
	uint8_t err;
	unsigned int bus = channels[channel].base;
	ide_write(channel, ATA_REG_CONTROL, channels[channel].nIEN = ide_irq_invoked = 0x0);
	ide_write(channel, ATA_REG_HDDEVSEL, drive << 4);
	ide_write(channel, ATA_REG_FEATURES, 1); // PIO mode
	ide_write(channel, ATA_REG_LBA1, 0);   // Lower Byte of Sector Size.
	ide_write(channel, ATA_REG_LBA2, 0);   // Upper Byte of Sector Size.
	ide_write(channel, ATA_REG_COMMAND, ATA_CMD_PACKET);      // Send the Command.
	atapi_packet[ 0] = ATAPI_CMD_WRITE;
	atapi_packet[ 1] = 0x0;
	atapi_packet[ 2] = (lba >> 24) & 0xFF;
	atapi_packet[ 3] = (lba >> 16) & 0xFF;
	atapi_packet[ 4] = (lba >> 8) & 0xFF;
	atapi_packet[ 5] = (lba >> 0) & 0xFF;
	atapi_packet[ 6] = 0x0;
	atapi_packet[ 7] = 0x0;
	atapi_packet[ 8] = 0x0;
	atapi_packet[ 9] = numsects;
	atapi_packet[10] = 0x0;
	atapi_packet[11] = 0x0;
	err = ide_polling(channel, 1);
	if (err) return err;
	region_desc* prd_table = channels[channel].prd_table;
	memcpy(channels[channel].dma_buffer,addr,numsects*words*2);

	for (int i = 0; i < numsects; i++) {
		prd_table[i].address = (uint32_t)&channels[channel].dma_buffer[i*words*2];
		prd_table[i].count = words * 2;
		prd_table[i].end = 0;
	}
	prd_table[numsects - 1].end = 0x8000;
	outl(channels[channel].bmide + IDE_DMA_PRD,(uint32_t)prd_table);
	ide_write(channel, ATA_REG_DMA_CMD, ide_read(channel, ATA_REG_DMA_CMD) &~(1 << 3));
	ide_write(channel, ATA_REG_DMA_STATUS, ide_read(channel, ATA_REG_DMA_STATUS) | 4 | 2);
	outsw(bus, atapi_packet, 6);
	ide_write(channel, ATA_REG_DMA_CMD, ide_read(channel, ATA_REG_DMA_CMD) | 1);
	ide_wait_irq();
	ide_write(channel, ATA_REG_DMA_CMD, ide_read(channel, ATA_REG_DMA_CMD) & (~1));
	return 0;
}

unsigned char ide_atapi_pio_read(unsigned char idx, unsigned int lba,
                                 unsigned char numsects, void* addr) {
	unsigned int channel = ide_devices[idx].channel;
	unsigned int drive = ide_devices[idx].drive;
	unsigned int   words      = 2048 / 2; // Sector Size in Words, Almost All ATAPI Drives has a sector size of 2048 bytes.
	uint8_t err;
	unsigned int   bus      = channels[channel].base;
	ide_write(channel, ATA_REG_CONTROL, channels[channel].nIEN = ide_irq_invoked = 0x0);
	ide_write(channel, ATA_REG_HDDEVSEL, drive << 4);
	ide_write(channel, ATA_REG_FEATURES, 0); // PIO mode
	ide_write(channel, ATA_REG_LBA1, (words * 2) & 0xFF);   // Lower Byte of Sector Size.
	ide_write(channel, ATA_REG_LBA2, (words * 2) >> 8); // Upper Byte of Sector Size.
	ide_write(channel, ATA_REG_COMMAND, ATA_CMD_PACKET);      // Send the Command.
	atapi_packet[ 0] = ATAPI_CMD_READ;
	atapi_packet[ 1] = 0x0;
	atapi_packet[ 2] = (lba >> 24) & 0xFF;
	atapi_packet[ 3] = (lba >> 16) & 0xFF;
	atapi_packet[ 4] = (lba >> 8) & 0xFF;
	atapi_packet[ 5] = (lba >> 0) & 0xFF;
	atapi_packet[ 6] = 0x0;
	atapi_packet[ 7] = 0x0;
	atapi_packet[ 8] = 0x0;
	atapi_packet[ 9] = numsects;
	atapi_packet[10] = 0x0;
	atapi_packet[11] = 0x0;
	err = ide_polling(channel, 1);
	if (err) return err;
	outsw(bus, atapi_packet, 6);
	for (int i = 0; i < numsects; i++) {
		ide_wait_irq();                  // Wait for an IRQ.
		err = ide_polling(channel, 1);
		if (err) return err;      // Polling and return if error.
		insw(bus, addr, words);
		addr += (words * 2);
	}
	ide_wait_irq();
	return 0;
}
unsigned char ide_atapi_pio_write(unsigned char idx, unsigned int lba,
                                  unsigned char numsects, void* addr) {
	unsigned int channel = ide_devices[idx].channel;
	unsigned int drive = ide_devices[idx].drive;
	unsigned int   words      = 2048 / 2; // Sector Size in Words, Almost All ATAPI Drives has a sector size of 2048 bytes.
	uint8_t err;
	unsigned int   bus      = channels[channel].base;
	ide_write(channel, ATA_REG_CONTROL, channels[channel].nIEN = ide_irq_invoked = 0x0);
	ide_write(channel, ATA_REG_HDDEVSEL, drive << 4);
	ide_write(channel, ATA_REG_FEATURES, 0); // PIO mode
	ide_write(channel, ATA_REG_LBA1, (words * 2) & 0xFF);   // Lower Byte of Sector Size.
	ide_write(channel, ATA_REG_LBA2, (words * 2) >> 8); // Upper Byte of Sector Size.
	ide_write(channel, ATA_REG_COMMAND, ATA_CMD_PACKET);      // Send the Command.
	atapi_packet[ 0] = ATAPI_CMD_READ;
	atapi_packet[ 1] = 0x0;
	atapi_packet[ 2] = (lba >> 24) & 0xFF;
	atapi_packet[ 3] = (lba >> 16) & 0xFF;
	atapi_packet[ 4] = (lba >> 8) & 0xFF;
	atapi_packet[ 5] = (lba >> 0) & 0xFF;
	atapi_packet[ 6] = 0x0;
	atapi_packet[ 7] = 0x0;
	atapi_packet[ 8] = 0x0;
	atapi_packet[ 9] = numsects;
	atapi_packet[10] = 0x0;
	atapi_packet[11] = 0x0;
	err = ide_polling(channel, 1);
	if (err) return err;
	outsw(bus, atapi_packet, 6);
	for (int i = 0; i < numsects; i++) {
		ide_wait_irq();                  // Wait for an IRQ.
		err = ide_polling(channel, 1);
		if (err) return err;      // Polling and return if error.
		outsw(bus, addr, words);
		addr += (words * 2);
	}
	ide_wait_irq();
	return 0;
}


unsigned char ide_atapi_access(uint8_t direction, unsigned char drive, unsigned int lba, unsigned char numsects,
                               void* addr) {
	unsigned int   channel      = ide_devices[drive].channel;
	unsigned int   slavebit      = ide_devices[drive].drive;
	unsigned int   bus      = channels[channel].base;
	unsigned int   words      = 2048 / 2; // Sector Size in Words, Almost All ATAPI Drives has a sector size of 2048 bytes.
	unsigned char  err;
	uint8_t dma = ide_devices[drive].dma;
	// Enable IRQs:
	ide_write(channel, ATA_REG_CONTROL, channels[channel].nIEN = ide_irq_invoked = 0x0);
	// (I): Setup SCSI Packet:
	// ------------------------------------------------------------------
	atapi_packet[ 0] = direction == ATAPI_READ ? ATAPI_CMD_READ : ATAPI_CMD_WRITE;
	atapi_packet[ 1] = 0x0;
	atapi_packet[ 2] = (lba >> 24) & 0xFF;
	atapi_packet[ 3] = (lba >> 16) & 0xFF;
	atapi_packet[ 4] = (lba >> 8) & 0xFF;
	atapi_packet[ 5] = (lba >> 0) & 0xFF;
	atapi_packet[ 6] = 0x0;
	atapi_packet[ 7] = 0x0;
	atapi_packet[ 8] = 0x0;
	atapi_packet[ 9] = numsects;
	atapi_packet[10] = 0x0;
	atapi_packet[11] = 0x0;
	// (II): Select the Drive:
	// ------------------------------------------------------------------
	ide_write(channel, ATA_REG_HDDEVSEL, slavebit << 4);
	// (III): Delay 400 nanosecond for select to complete:
	ide_delay(channel);
	// (IV): Inform the Controller that we use PIO mode:
	// ------------------------------------------------------------------
	ide_write(channel, ATA_REG_FEATURES, 0);         // PIO mode.
	// (V): Tell the Controller the size of buffer:
	// ------------------------------------------------------------------
	ide_write(channel, ATA_REG_LBA1, (words * 2) & 0xFF);   // Lower Byte of Sector Size.
	ide_write(channel, ATA_REG_LBA2, (words * 2) >> 8); // Upper Byte of Sector Size.
	// (VI): Send the Packet Command:
	// ------------------------------------------------------------------
	ide_write(channel, ATA_REG_COMMAND, ATA_CMD_PACKET);      // Send the Command.
	// (VII): Waiting for the driver to finish or invoke an error:
	// ------------------------------------------------------------------
	err = ide_polling(channel, 1);
	if (err) return err;         // Polling and return if error.
	if (dma) {
		uint16_t port = 1;
		region_desc* prd_table = channels[channel].prd_table;
		for (int i = 0; i < numsects; i++) {
			prd_table[i].address = (uint32_t)addr;
			prd_table[i].count = words * 2;
			prd_table[i].end = 0;
		}
		prd_table[numsects - 1].end = 0x8000;
		if (direction == ATAPI_READ)
			outb(port + ATA_REG_DMA_CMD, inb(port + ATA_REG_DMA_CMD) | (1 << 3));
		else
			outb(port + ATA_REG_DMA_CMD, inb(port + ATA_REG_DMA_CMD) & ~(1 << 3));
		outb(port + ATA_REG_DMA_STATUS, inb(port + ATA_REG_DMA_STATUS) | 4 | 2);
		outsw(bus, atapi_packet, 6);
		outb(port + ATA_REG_DMA_CMD, inb(port + ATA_REG_DMA_CMD) | 1);
		debug("wait irq for dma\n");
		ide_wait_irq();
		debug("wait irq for dma done\n");
		// reset the Start/Stop bit
		// outb(port + IDE_DMA_CMD,0);
		if (direction == 1) {
			ide_polling(channel, 0); // Polling.
		}
	} else {
		outsw(bus, atapi_packet, 6);
		for (int i = 0; i < numsects; i++) {
			ide_wait_irq();                  // Wait for an IRQ.
			err = ide_polling(channel, 1);
			if (err) return err;      // Polling and return if error.
			if (direction == ATAPI_READ)
				insw(bus, addr, words);
			else
				outsw(bus, addr, words);
			addr += (words * 2);
		}
		// (X): Waiting for an IRQ:
		// ------------------------------------------------------------------
		ide_wait_irq();
	}
	// (XI): Waiting for BSY & DRQ to clear:
	// ------------------------------------------------------------------
	while (ide_read(channel, ATA_REG_STATUS) & (ATA_SR_BSY | ATA_SR_DRQ));

	return 0; // Easy, ... Isn't it?
}




