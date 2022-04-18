#include <virtio.h>
#include <pci.h>
#include <libcc.h>
#include <picirq.h>
#include <x86.h>
#include <virtio_dev.h>
#include <virtio.h>
#include <virtio_gpu.h>

virtio_device* gpu_dev;
int virtio_gpu_init(virtio_device* vdev){
    virtio_pci_dev* pdev = &vdev->pdev;
    pdev->ops->set_status(pdev,0);
	uint8_t c = VIRTIO_ACKNOWLEDGE;
    pdev->ops->set_status(pdev,c);
    c |= VIRTIO_DRIVER;
    pdev->ops->set_status(&vdev->pdev,c);
    uint64_t device_feature = pdev->ops->get_features(pdev);
    // kprintf("gpu device feature: %x\n",device_feature);

    pdev->ops->set_features(pdev,device_feature);
    pdev->guest_feature = device_feature;
    c |= VIRTIO_FEATURES_OK;
    pdev->ops->set_status(pdev,c);
    uint8_t virtio_status = pdev->ops->get_status(pdev);
    if((virtio_status&VIRTIO_FEATURES_OK) == 0){
        printf("gpu feature is not ok\n");
        return 0;
    }
    for(int i=0;i<QUEUE_COUNT;i++){
        setup_virtqueue(vdev,i);
    }
    c |= VIRTIO_DRIVER_OK;
    pdev->ops->set_status(pdev,c);
    virtio_status = pdev->ops->get_status(pdev);
    if (virtio_status & VIRTIO_FAILED){
        printf("virtio gpu init failed\n");
        return 0;
    }
    // show_device_status(vdev);
    gpu_dev = vdev;
    return 1;
}












