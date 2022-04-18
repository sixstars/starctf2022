
#include <stdio.h>
#include <picirq.h>
#include <pci.h>
#include <trap.h>
#include <ide.h>
#include <clock.h>
#include <physical_page.h>


#define DESC_END 0x8000
#define DESC_NEXT 0

#define IDE_CMD_WRITE 0x80
#define IDE_CMD_READ 0x00
#define IDE_CMD_START 0x01
#define IDE_CMD_STOP 0x00

#define IDE_STATUS_INTR 2
#define IDE_STATUS_ERROR 1
#define CLEAR_BIT(x,offset) (x &= ~(1<<offset))
#define CHECK_BIT(x,offset) (x & (1<<offset))

static void ide_initialize(uint16_t bar0, uint16_t bar1,
                           uint16_t bar2, uint16_t bar3, uint16_t bar4);
void parse_progif(uint8_t pg);


struct IDEChannelRegisters channels[2];

unsigned char package[3];
unsigned char ide_buf[2048] = {0};
unsigned char ide_irq_invoked = 0;
unsigned char atapi_packet[12] = {0xA8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0};

struct ide_device ide_devices[4];

void ide_init() {
	Device* dev = 0;
	for (int i = 0; i < device_num; i++) {
		if (devices[i].class_code == 1 && devices[i].subclass == 1) {
			debug("find IDE device[%d] at: %02x:%02x:%1x\n",i, devices[i].bus_id, devices[i].device_id,
			      devices[i].func);
			dev = &devices[i];
			break;
		}
	}
	if (!dev)
		return;
	PCI_loadbars(dev);
	parse_progif(dev->prog_if);
	// we except is 0x80:ISA Compatibility mode-only controller, supports bus mastering
	// primary channel is in compatibility mode (ports 0x1F0-0x1F7, 0x3F6, IRQ14).
	// secondary channel is in compatibility mode (ports 0x170-0x177, 0x376, IRQ15).
	// bar0 0x1f0 bar1 0x3f6
	// bar2 0x170 bar3 0x376
	// bar4  dev->iobase
	register_intr_handler(IRQ_IDE_PRI + IRQ_OFFSET, ide_irq);
	register_intr_handler(IRQ_IDE_SEC + IRQ_OFFSET, ide_irq);
	pic_enable(IRQ_IDE_PRI);
	pic_enable(IRQ_IDE_SEC);

	ide_initialize(dev->reg_base[0], dev->reg_base[1], dev->reg_base[2],
	               dev->reg_base[3], dev->reg_base[4]);

	// uint8_t status;
//   status = inb(channels[0].bmide + IDE_DMA_STATUS);
//   status |= 0x60;
//   outb(channels[0].bmide + IDE_DMA_STATUS,status);

//   status = inb(channels[1].bmide + IDE_DMA_STATUS);
//   status |= 0x60;
//   outb(channels[1].bmide + IDE_DMA_STATUS,status);
	ide_enable_dma(0);
	ide_enable_dma(1);
}



void ide_enable_dma(int idx) {
	ide_devices[idx].dma = 1;

	uint8_t channel = ide_devices[idx].channel;
	uint8_t drive = ide_devices[idx].drive;
	channels[channel].nIEN = 0;
	if (!channels[idx].prd_table) {
		channels[idx].prd_table = (region_desc*)physical_alloc(0x8000, 0x10000);
		channels[idx].dma_buffer = (uint8_t*)physical_alloc(0x8000, 0x10000);
	}
	// uint16_t port = channels[idx].bmide;

	if (ide_devices[idx].udma > 0) {
		// outl(port + 4,(uint32_t)channels[idx].prd_table);
		ide_write(channel, ATA_REG_HDDEVSEL, 0xE0 | (drive << 4));
		ide_delay(channel);
		ide_write(channel, ATA_REG_SECCOUNT0, 8 << 3 | ide_devices[idx].udma);
		ide_write(channel, ATA_REG_FEATURES, 3);
		ide_write(channel, ATA_REG_COMMAND, 0xef);
		ide_wait_irq();
	} else if (ide_devices[idx].mdma > 0) {
		// outl(port + 4,(uint32_t)channels[idx].prd_table);
		ide_write(channel, ATA_REG_HDDEVSEL, 0xA0 | (drive << 4));
		ide_delay(channel);
		ide_write(channel, ATA_REG_SECCOUNT0, 8 << 3 | ide_devices[idx].mdma);
		ide_write(channel, ATA_REG_FEATURES, 3);
		ide_write(channel, ATA_REG_COMMAND, 0xef);
		ide_wait_irq();
	} else {
		debug("ide_enable_dma fail\n");
		ide_devices[idx].dma = 0;
	}
	ide_write(channel, ATA_REG_DMA_CMD, ide_read(channel, ATA_REG_DMA_CMD) & (~1));
}

void ide_disable_dma(int idx) {
	ide_devices[idx].dma = 0;
}


void ide_write(unsigned char channel, unsigned char reg, unsigned char data) {
	if (reg > 0x07 && reg < 0x0C)
		//Set this to read back the High Order Byte of the last LBA48 value sent to an IO port.
		ide_write(channel, ATA_REG_CONTROL, 0x80 | channels[channel].nIEN); //
	if (reg < 0x08)
		outb(channels[channel].base  + reg - 0x00, data);
	else if (reg < 0x0C)
		outb(channels[channel].base  + reg - 0x06, data);
	else if (reg < 0x0E)
		outb(channels[channel].ctrl  + reg - 0x0A, data);
	else if (reg < 0x16)
		outb(channels[channel].bmide + reg - 0x0E, data);
	if (reg > 0x07 && reg < 0x0C)
		ide_write(channel, ATA_REG_CONTROL, channels[channel].nIEN);
}

unsigned char ide_read(unsigned char channel, unsigned char reg) {
	unsigned char result;
	if (reg > 0x07 && reg < 0x0C)
		ide_write(channel, ATA_REG_CONTROL, 0x80 | channels[channel].nIEN);
	if (reg < 0x08)
		result = inb(channels[channel].base + reg - 0x00);
	else if (reg < 0x0C)
		result = inb(channels[channel].base  + reg - 0x06);
	else if (reg < 0x0E)
		result = inb(channels[channel].ctrl  + reg - 0x0A);
	else if (reg < 0x16)
		result = inb(channels[channel].bmide + reg - 0x0E);
	if (reg > 0x07 && reg < 0x0C)
		ide_write(channel, ATA_REG_CONTROL, channels[channel].nIEN);
	return result;
}

void ide_read_buffer(unsigned char channel, unsigned char reg, void* buffer,
                     unsigned int quads) {
	/* WARNING: This code contains a serious bug. The inline assembly trashes ES and
	 *           ESP for all of the code the compiler generates between the inline
	 *           assembly blocks.
	 */
	if (reg > 0x07 && reg < 0x0C)
		ide_write(channel, ATA_REG_CONTROL, 0x80 | channels[channel].nIEN);
	// asm("pushw %es; movw %ds, %ax; movw %ax, %es");
	if (reg < 0x08)
		insl(channels[channel].base  + reg - 0x00, buffer, quads);
	else if (reg < 0x0C)
		insl(channels[channel].base  + reg - 0x06, buffer, quads);
	else if (reg < 0x0E)
		insl(channels[channel].ctrl  + reg - 0x0A, buffer, quads);
	else if (reg < 0x16)
		insl(channels[channel].bmide + reg - 0x0E, buffer, quads);
	// asm("popw %es;");
	if (reg > 0x07 && reg < 0x0C)
		ide_write(channel, ATA_REG_CONTROL, channels[channel].nIEN);
}

static void ide_initialize(uint16_t bar0, uint16_t bar1,
                           uint16_t bar2, uint16_t bar3, uint16_t bar4) {
	uint8_t count = 0;
	// 1- Detect I/O Ports which interface IDE Controller:
	channels[ATA_PRIMARY  ].base  = (bar0 & 0xFFFFFFFC) + 0x1F0 * (!bar0);
	channels[ATA_PRIMARY  ].ctrl  = (bar1 & 0xFFFFFFFC) + 0x3F6 * (!bar1);
	channels[ATA_SECONDARY].base  = (bar2 & 0xFFFFFFFC) + 0x170 * (!bar2);
	channels[ATA_SECONDARY].ctrl  = (bar3 & 0xFFFFFFFC) + 0x376 * (!bar3);
	channels[ATA_PRIMARY  ].bmide = (bar4 & 0xFFFFFFFC) + 0; // Bus Master IDE
	channels[ATA_SECONDARY].bmide = (bar4 & 0xFFFFFFFC) + 8; // Bus Master IDE
	// 2- Disable IRQs:
	ide_write(ATA_PRIMARY  , ATA_REG_CONTROL, 2);
	ide_write(ATA_SECONDARY, ATA_REG_CONTROL, 2);
	// 3- Detect ATA-ATAPI Devices:
	for (int i = 0; i < 2; i++) {
		for (int j = 0; j < 2; j++) {
			uint8_t err = 0;
			uint8_t type = IDE_ATA;
			ide_devices[count].reserved = 0;
			// (I) Select Drive:
			ide_write(i, ATA_REG_HDDEVSEL, 0xA0 | (j << 4)); // Select Drive.
			ide_delay(i);
			// (II) Send ATA Identify Command:
			ide_write(i, ATA_REG_COMMAND, ATA_CMD_IDENTIFY);
			ide_delay(i); // This function should be implemented in your OS. which waits for 1 ms.
			// it is based on System Timer Device Driver.
			if (ide_read(i, ATA_REG_STATUS) == 0) continue;
			while (1) {
				uint8_t status = ide_read(i, ATA_REG_STATUS);
				if ((status & ATA_SR_ERR)) {err = 1; break;} // If Err, Device is not ATA.
				if (!(status & ATA_SR_BSY) && (status & ATA_SR_DRQ)) break; // Everything is right.
			}
			if (err) {
				unsigned char cl = ide_read(i, ATA_REG_LBA1);
				unsigned char ch = ide_read(i, ATA_REG_LBA2);
				if (cl == 0x14 && ch == 0xEB)
					type = IDE_ATAPI;
				else if (cl == 0x69 && ch == 0x96)
					type = IDE_ATAPI;
				else
					continue; // Unknown Type (may not be a device).
				ide_write(i, ATA_REG_COMMAND, ATA_CMD_IDENTIFY_PACKET);
				babysleep(50);
			}
			memset(ide_buf, 0, sizeof(ide_buf));

			// (V) Read Identification Space of the Device:
			ide_read_buffer(i, ATA_REG_DATA, ide_buf, 128);

			// (VI) Read Device Parameters:
			ide_devices[count].reserved     = 1;
			ide_devices[count].type         = type;
			ide_devices[count].channel      = i;
			ide_devices[count].drive        = j;
			ide_devices[count].signature    = *((unsigned short *)(ide_buf + ATA_IDENT_DEVICETYPE));
			ide_devices[count].capabilities = *((unsigned short *)(ide_buf + ATA_IDENT_CAPABILITIES));
			ide_devices[count].commandSets  = *((unsigned int *)(ide_buf + ATA_IDENT_COMMANDSETS));

			// (VII) Get Size:
			if (ide_devices[count].commandSets & (1 << 26)) {
				// Device uses 48-Bit Addressing:
				ide_devices[count].size   = *((unsigned int *)(ide_buf + ATA_IDENT_MAX_LBA_EXT));
			} else {
				// Device uses CHS or 28-bit Addressing:
				ide_devices[count].size   = *((unsigned int *)(ide_buf + ATA_IDENT_MAX_LBA));

			}

			// (VIII) String indicates model of device (like Western Digital HDD and SONY DVD-RW...):
			for (int k = 0; k < 40; k += 2) {
				ide_devices[count].model[k] = ide_buf[ATA_IDENT_MODEL + k + 1];
				ide_devices[count].model[k + 1] = ide_buf[ATA_IDENT_MODEL + k];
			}
			ide_devices[count].model[40] = 0; // Terminate String.

			ide_devices[count].udma = -1;
			ide_devices[count].mdma = -1;
			//check udma support
			for (int bit = 6; bit >= 0; bit--) {
				if (ide_buf[ATA_IDENT_UDMASUP] & (1 << bit)) {
					ide_devices[count].udma = bit;
					break;
				}
			}
			for (int bit = 2; bit >= 0; bit--) {
				if (ide_buf[ATA_IDENT_MDMASUP] & (1 << bit)) {
					ide_devices[count].mdma = bit;
					break;
				}
			}

			count++;
		}
	}
	// 4- Print Summary:
	for (int i = 0; i < 4; i++)
		if (ide_devices[i].reserved == 1) {
			debug("IDE[%d:%d] %s Drive %dMB - %s\n",
			      ide_devices[i].channel, ide_devices[i].drive,
			(const char *[]) {"ATA", "ATAPI"}[ide_devices[i].type],        /* Type */
			ide_devices[i].size / 1024 / 2,               /* Size */
			ide_devices[i].model);
		}
}


unsigned char ide_polling(unsigned char channel, unsigned int advanced_check) {
	// (I) Delay 400 nanosecond for BSY to be set:
	// -------------------------------------------------
	ide_delay(channel);
	// (II) Wait for BSY to be cleared:
	// -------------------------------------------------
	while (ide_read(channel, ATA_REG_STATUS) & ATA_SR_BSY); // Wait for BSY to be zero.

	if (advanced_check) {
		unsigned char state = ide_read(channel, ATA_REG_STATUS); // Read Status Register.

		// (III) Check For Errors:
		// -------------------------------------------------
		if (state & ATA_SR_ERR)
			return 2; // Error.

		// (IV) Check If Device fault:
		// -------------------------------------------------
		if (state & ATA_SR_DF)
			return 1; // Device Fault.

		// (V) Check DRQ:
		// -------------------------------------------------
		// BSY = 0; DF = 0; ERR = 0 so we should check for DRQ now.
		if ((state & ATA_SR_DRQ) == 0)
			return 3; // DRQ should be set

	}

	return 0; // No Error.

}







void ide_read_sectors(unsigned char drive, unsigned char numsects, unsigned int lba, void* addr) {
	// 1: Check if the drive presents:
	// ==================================
	if (drive > 3 || ide_devices[drive].reserved == 0) package[0] = 0x1;      // Drive Not Found!

	// 2: Check if inputs are valid:
	// ==================================
	else if (((lba + numsects) > ide_devices[drive].size) && (ide_devices[drive].type == IDE_ATA))
		package[0] = 0x2;                     // Seeking to invalid position.

	// 3: Read in PIO Mode through Polling & IRQs:
	// ============================================
	else {
		unsigned char err;
		if (ide_devices[drive].type == IDE_ATA) {
			if (ide_devices[drive].dma) {
				err = ide_ata_dma_read(drive, lba, numsects, addr);
			} else {
				err = ide_ata_pio_read(drive, lba, numsects, addr);
			}
		}
		if (ide_devices[drive].type == IDE_ATAPI) {
			for (int i = 0; i < numsects; i++) {
				if (ide_devices[drive].dma) {
					err = ide_atapi_dma_read(drive, lba + i, 1, addr + i * 2048);
				} else {
					err = ide_atapi_pio_read(drive, lba + i, 1, addr + i * 2048);
				}
			}
		}
		package[0] = ide_print_error(drive, err);
	}
}

void ide_write_sectors(unsigned char drive, unsigned char numsects, unsigned int lba, void* addr) {

	// 1: Check if the drive presents:
	// ==================================
	if (drive > 3 || ide_devices[drive].reserved == 0) package[0] = 0x1;      // Drive Not Found!
	// 2: Check if inputs are valid:
	// ==================================
	else if (((lba + numsects) > ide_devices[drive].size) && (ide_devices[drive].type == IDE_ATA))
		package[0] = 0x2;                     // Seeking to invalid position.
	// 3: Read in PIO Mode through Polling & IRQs:
	// ============================================
	else {
		unsigned char err;
		if (ide_devices[drive].type == IDE_ATA) {
			if (ide_devices[drive].dma) {
				debug("write from ata using dma\n");
				err = ide_ata_dma_write(drive, lba, numsects, addr);
			} else {
				debug("write from ata using pio\n");
				err = ide_ata_pio_write(drive, lba, numsects, addr);
			}
		}
		else if (ide_devices[drive].type == IDE_ATAPI) {
			if (ide_devices[drive].dma) {
				err = ide_atapi_dma_write(drive, lba, numsects, addr);
			} else {
				err = ide_atapi_pio_write(drive, lba, numsects, addr);
			}
		}
		package[0] = ide_print_error(drive, err);
	}
}



void parse_progif(uint8_t pg) {
	/*
	https://wiki.osdev.org/IDE
	Bit 0: When set, the primary channel is in PCI native mode. When clear, the primary channel is in compatibility mode (ports 0x1F0-0x1F7, 0x3F6, IRQ14).
	Bit 1: When set, you can modify bit 0 to switch between PCI native and compatibility mode. When clear, you cannot modify bit 0.
	Bit 2: When set, the secondary channel is in PCI native mode. When clear, the secondary channel is in compatibility mode (ports 0x170-0x177, 0x376, IRQ15).
	Bit 3: When set, you can modify bit 2 to switch between PCI native and compatibility mode. When clear, you cannot modify bit 2.
	Bit 7: When set, this is a bus master IDE controller. When clear, this controller doesn't support DMA.
	*/
	debug("IDE mode(0x%02x)", pg);
	switch (pg) {
	case 0:
		debug("ISA Compatibility mode-only controller\n");
		break;
	case 5:
		debug("PCI native mode-only controller\n");
		break;
	case 0xa:
		debug("ISA Compatibility mode controller, supports both channels switched to PCI native mode\n");
		break;
	case 0xf:
		debug("PCI native mode controller, supports both channels switched to ISA compatibility mode\n");
		break;
	case 0x80:
		debug("ISA Compatibility mode-only controller, supports bus mastering\n");
		break;
	case 0x85:
		debug("PCI native mode-only controller, supports bus mastering\n");
		break;
	case 0x8a:
		debug("ISA Compatibility mode controller, supports both channels switched to PCI native mode, supports bus mastering\n");
		break;
	case 0x8f:
		debug("PCI native mode controller, supports both channels switched to ISA compatibility mode, supports bus mastering\n");
		break;
	}
}


unsigned char ide_print_error(unsigned int drive, unsigned char err) {
	if (err == 0) return err;
	printf(" IDE:");
	if (err == 1) {printf("- Device Fault\n     "); err = 19;}
	else if (err == 2) {
		unsigned char st = ide_read(ide_devices[drive].channel, ATA_REG_ERROR);
		if (st & ATA_ER_AMNF)   {printf("- No Address Mark Found\n     ");   err = 7;}
		if (st & ATA_ER_TK0NF)   {printf("- No Media or Media Error\n     ");   err = 3;}
		if (st & ATA_ER_ABRT)   {printf("- Command Aborted\n     ");      err = 20;}
		if (st & ATA_ER_MCR)   {printf("- No Media or Media Error\n     ");   err = 3;}
		if (st & ATA_ER_IDNF)   {printf("- ID mark not Found\n     ");      err = 21;}
		if (st & ATA_ER_MC)   {printf("- No Media or Media Error\n     ");   err = 3;}
		if (st & ATA_ER_UNC)   {printf("- Uncorrectable Data Error\n     ");   err = 22;}
		if (st & ATA_ER_BBK)   {printf("- Bad Sectors\n     ");       err = 13;}
	} else  if (err == 3)           {printf("- Reads Nothing\n     "); err = 23;}
	else  if (err == 4)  {printf("- Write Protected\n     "); err = 8;}

	return err;
}


void ide_delay(int channel) {
	for (int i = 0; i < 4; i++)
		ide_read(channel, ATA_REG_ALTSTATUS);
}

void ide_wait_irq() {
	while (!ide_irq_invoked) {
		asm("hlt");
	}
	ide_irq_invoked = 0;
}

void ide_irq() {
	ide_irq_invoked = 1;
	ide_write(0, ATA_REG_DMA_STATUS, ide_read(0, ATA_REG_DMA_STATUS) | 1 << 2);
	ide_write(0, ATA_REG_DMA_STATUS, ide_read(1, ATA_REG_DMA_STATUS) | 1 << 2);
}

uint8_t ide_access_drive(uint8_t channel, uint8_t drive, uint32_t lba, uint32_t numsects) {
	uint8_t lba_mode; /* 0: CHS, 1:LBA28, 2: LBA48 */
	uint8_t head, sect;
	uint16_t cyl;
	uint8_t lba_io[6];
	if (lba > 0x10000000) { // Sure Drive should support LBA in this case, or you are giving a wrong LBA.
		lba_mode  = 2;
		lba_io[0] = (lba & 0x000000FF) >> 0;
		lba_io[1] = (lba & 0x0000FF00) >> 8;
		lba_io[2] = (lba & 0x00FF0000) >> 16;
		lba_io[3] = (lba & 0xFF000000) >> 24;
		lba_io[4] = 0; // We said that we lba is integer, so 32-bit are enough to access 2TB.
		lba_io[5] = 0; // We said that we lba is integer, so 32-bit are enough to access 2TB.
		head      = 0; // Lower 4-bits of HDDEVSEL are not used here.
	} else if (ide_devices[drive].capabilities & 0x200)  {
		lba_mode = 1;
		lba_io[0] = (lba & 0x00000FF) >> 0;
		lba_io[1] = (lba & 0x000FF00) >> 8;
		lba_io[2] = (lba & 0x0FF0000) >> 16;
		lba_io[3] = 0; // These Registers are not used here.
		lba_io[4] = 0; // These Registers are not used here.
		lba_io[5] = 0; // These Registers are not used here.
		head      = (lba & 0xF000000) >> 24;
	} else {
		lba_mode  = 0;
		sect      = (lba % 63) + 1;
		cyl = (lba + 1  - sect) / (16 * 63);
		lba_io[0] = sect;
		lba_io[1] = (cyl >> 0) & 0xFF;
		lba_io[2] = (cyl >> 8) & 0xFF;
		lba_io[3] = 0;
		lba_io[4] = 0;
		lba_io[5] = 0;
		head      = (lba + 1  - sect) % (16 * 63) / (63); // Head number is written to HDDEVSEL lower 4-bits.
	}
	while (ide_read(channel, ATA_REG_STATUS) & ATA_SR_BSY); // Wait if Busy.

	// (IV) Select Drive from the controller;
	if (lba_mode == 0)
		ide_write(channel, ATA_REG_HDDEVSEL, 0xA0 | (drive << 4) | head); // Select Drive CHS.
	else
		ide_write(channel, ATA_REG_HDDEVSEL, 0xE0 | (drive << 4) | head); // Select Drive LBA.
	// (V) Write Parameters;
	if (lba_mode == 2) {
		ide_write(channel, ATA_REG_SECCOUNT1,   0);
		ide_write(channel, ATA_REG_LBA3,   lba_io[3]);
		ide_write(channel, ATA_REG_LBA4,   lba_io[4]);
		ide_write(channel, ATA_REG_LBA5,   lba_io[5]);
	}
	ide_write(channel, ATA_REG_SECCOUNT0,   numsects);
	ide_write(channel, ATA_REG_LBA0,   lba_io[0]);
	ide_write(channel, ATA_REG_LBA1,   lba_io[1]);
	ide_write(channel, ATA_REG_LBA2,   lba_io[2]);
	return lba_mode;
}



