
#include <virtio.h>
#include <virtio_dev.h>
#include <virtio_net.h>
#include <virtio_gpu.h>
#include <virtio_ops.h>
#include <libcc.h>

virtio_device vdevs[0x10];

virtio_device* alloc_virtdev(Device* dev){
    int i = 0;
    for(;i<sizeof(vdevs)/sizeof(vdevs[0]);i++){
        if(vdevs[i].inuse == 0)
            break;
    }
    if(i == sizeof(vdevs)/sizeof(vdevs[0]))
        return NULL;
    virtio_device* vdev = &vdevs[i];
    memset(vdev,0,sizeof(*vdev));
    vdev->inuse = 1;
    vdev->pdev.pci = dev;
    vdev->pdev.iobase = dev->iobase;
    vdev->pdev.irq = dev->irq;

    if(virtio_read_caps(&vdev->pdev) == 0){
        // modern mode
        vdev->modern = true;
        vdev->pdev.ops = &modern_ops;
        debug("use modern mode\n");

    }else{
        vdev->modern = false;
        vdev->pdev.ops = &legacy_ops;
        debug("use legacy mode\n");
    }
    return vdev;
}

void virtio_dev_install(){
    for(int i=0;i<device_num;i++){
        if(devices[i].vendor == 0x1af4
                && devices[i].device >= 0x1000
                && devices[i].device <= 0x107f)
        {
            PCI_loadbars(&devices[i]);
            virtio_device* vdev = alloc_virtdev(&devices[i]);
        	switch(devices[i].subsystem_id){
        		case VIRTIO_DEVICE_NET:
                    printf("find network card\n");
		            if(!network_card_init(vdev)){
		                vdev->inuse = 0;
		            }else{
			            network_card_setup(vdev);
		            }
	        		break;
	        	case VIRTIO_DEVICE_GPU:
                     printf("find gpu card\n");
	        		if(!virtio_gpu_init(vdev)){
	        			vdev->inuse = 0;
	        		}else{

	        		}
		        	break;
		        default:
                    debug("unknown virtio device: %d\n",devices[i].subsystem_id);
			        break;
        	}
        }
    }
}
void show_device_status(virtio_device* vdev){

    uint16_t status = vdev->pdev.ops->get_status(&vdev->pdev);
    printf("device status: ");
    if(status & VIRTIO_FAILED){
        printf("VIRTIO_FAILED");
    }
    if(status & VIRTIO_ACKNOWLEDGE){
        printf("VIRTIO_ACKNOWLEDGE ");
    }
    if(status & VIRTIO_DRIVER){
        printf("VIRTIO_DRIVER ");
    }
    if(status & VIRTIO_DRIVER_OK){
        printf("VIRTIO_DRIVER_OK ");
    }
    if(status & VIRTIO_FEATURES_OK){
        printf("VIRTIO_FEATURES_OK ");
    }
    if(status & VIRTIO_DEVICE_NEEDS_RESET){
        printf("VIRTIO_DEVICE_NEEDS_RESET ");
    }
    printf("\n");
}
