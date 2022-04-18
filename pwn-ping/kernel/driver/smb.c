

#include <smb.h>
#include <pci.h>
#include <libcc.h>

Device* smb_dev;
void smb_init(){
	Device* dev = 0;
	for(int i=0;i<device_num;i++){
		if(devices[i].vendor == 0x8086 && devices[i].device == 0x283e){
			dev = &devices[i];
			break;
		}
	}
	if(!dev){
		// debug("can't find smb bus\n");
		return;
	}
	smb_dev = dev;
	PCI_loadbars(dev);
	debug("smb io port: %x\n",dev->iobase);
}
