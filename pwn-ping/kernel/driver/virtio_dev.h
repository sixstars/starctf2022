#ifndef VIRTIO_DEV_H
#define VIRTIO_DEV_H

#include <virtio_pci.h>

#define VIRTIO_DEVICE_NET 1
#define VIRTIO_DEVICE_GPU 16


typedef struct {
    int inuse;
    virtio_pci_dev pdev;
    bool modern;
    virt_queue queue[QUEUE_COUNT];
} virtio_device;


virtio_device* alloc_virtdev(Device* dev);
void show_device_status(virtio_device* vdev);
void virtio_dev_install();



extern virtio_device vdevs[0x10];


#endif
