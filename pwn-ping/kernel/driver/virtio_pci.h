#ifndef __VIRTIO_PCI_H
#define __VIRTIO_PCI_H
#include <pci.h>
#include <virtio_queue.h>
/* This is the PCI capability header: */
struct virtio_pci_cap {
    uint8_t cap_vndr;       /* Generic PCI field: PCI_CAP_ID_VNDR */
    uint8_t cap_next;       /* Generic PCI field: next ptr. */
    uint8_t cap_len;        /* Generic PCI field: capability length */
    uint8_t cfg_type;       /* Identifies the structure. */
    uint8_t bar;            /* Where to find it. */
    uint8_t padding[3];     /* Pad to full dword. */
    uint32_t offset;        /* Offset within bar. */
    uint32_t length;        /* Length of the structure, in bytes. */
};


struct virtio_pci_common_cfg {
    /* About the whole device. */
    uint32_t device_feature_select;     /* read-write , 0*/
    uint32_t device_feature;            /* read-only for driver , 4*/
    uint32_t guest_feature_select;     /* read-write , 8*/
    uint32_t guest_feature;            /* read-write , 12*/
    uint16_t msix_config;               /* read-write , 20*/
    uint16_t num_queues;                /* read-only for driver , 22*/
    uint8_t device_status;               /* read-write , 24*/
    uint8_t config_generation;           /* read-only for driver , 25*/

    /* About a specific virtqueue. */
    uint16_t queue_select;              /* read-write , 26*/
    uint16_t queue_size;                /* read-write, power of 2, or 0. , 28*/
    uint16_t queue_msix_vector;         /* read-write , 30*/
    uint16_t queue_enable;              /* read-write , 32*/
    uint16_t queue_notify_off;          /* read-only for driver , 34*/
    uint64_t queue_desc;                /* read-write , 36*/
    uint64_t queue_avail;               /* read-write , 42*/
    uint64_t queue_used;                /* read-write , 50*/
} __attribute__((packed));

#define QUEUE_COUNT 4
enum virtio_msix_status {
    VIRTIO_MSIX_NONE = 0,
    VIRTIO_MSIX_DISABLED = 1,
    VIRTIO_MSIX_ENABLED = 2
};


typedef struct {
    uint16_t iobase;
    struct virtio_pci_common_cfg* common_cfg;
    uint8_t* device_cfg;
    uint8_t irq;
    uint16_t* notify_base;
    uint32_t notify_off_multiplier;
    uint8_t* isr;
    uint8_t macaddr[6];
    enum virtio_msix_status msix_status;
    Device* pci;
    const struct virtio_ops * ops;
    uint32_t guest_feature;
} virtio_pci_dev;

struct virtio_ops {
    void (*read_dev_cfg)(virtio_pci_dev *hw, uint32_t offset, void *dst, int len);
    void (*write_dev_cfg)(virtio_pci_dev *hw, uint32_t offset, const void *src, int len);
    uint8_t (*get_status)(virtio_pci_dev *hw);
    void (*set_status)(virtio_pci_dev *hw, uint8_t status);
    uint64_t (*get_features)(virtio_pci_dev *hw);
    void (*set_features)(virtio_pci_dev *hw, uint64_t features);
    int (*features_ok)(virtio_pci_dev *hw);
    uint8_t (*get_isr)(virtio_pci_dev *hw);
    uint16_t (*set_config_irq)(virtio_pci_dev *hw, uint16_t vec);
    uint16_t (*set_queue_irq)(virtio_pci_dev *hw, uint16_t vq, uint16_t vec);
    uint16_t (*get_queue_num)(virtio_pci_dev *hw, uint16_t queue_id);
    int (*setup_queue)(virtio_pci_dev *hw, virt_queue* vq);
    void (*del_queue)(virtio_pci_dev *hw, virt_queue* vq);
    void (*notify_queue)(virtio_pci_dev *hw, virt_queue* vq);
    void (*intr_detect)(virtio_pci_dev *hw);
    int (*dev_close)(virtio_pci_dev *hw);
};



#define VIRTIO_PCI_CONFIG(dev) \
        (((dev)->msix_status == VIRTIO_MSIX_ENABLED) ? 24 : 20)

int virtio_read_caps(virtio_pci_dev* pdev);

/* Common configuration */
#define VIRTIO_PCI_CAP_COMMON_CFG        1
/* Notifications */
#define VIRTIO_PCI_CAP_NOTIFY_CFG        2
/* ISR Status */
#define VIRTIO_PCI_CAP_ISR_CFG           3
/* Device specific configuration */
#define VIRTIO_PCI_CAP_DEVICE_CFG        4
/* PCI configuration access */
#define VIRTIO_PCI_CAP_PCI_CFG           5

#define PCI_CAPABILITY_LIST 0x34
#define PCI_CAP_ID_VNDR     0x09
#define PCI_CAP_ID_MSIX     0x11

#endif