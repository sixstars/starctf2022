#include <virtio_pci.h>
#include <virtio_ops.h>
#define VIRTIO_PCI_HOST_FEATURES  0  /* host's supported features (32bit, RO)*/
#define VIRTIO_PCI_GUEST_FEATURES 4  /* guest's supported features (32, RW) */
#define VIRTIO_PCI_QUEUE_PFN      8  /* physical address of VQ (32, RW) */
#define VIRTIO_PCI_QUEUE_NUM      12 /* number of ring entries (16, RO) */
#define VIRTIO_PCI_QUEUE_SEL      14 /* current VQ selection (16, RW) */
#define VIRTIO_PCI_QUEUE_NOTIFY   16 /* notify host regarding VQ (16, RW) */
#define VIRTIO_PCI_STATUS         18 /* device status register (8, RW) */
#define VIRTIO_PCI_ISR		      19 /* interrupt status register, reading */

void legacy_read_dev_config(virtio_pci_dev *pdev, uint32_t offset, void *dst, int len){
	uint16_t offset_base = pdev->pci->iobase + VIRTIO_PCI_CONFIG(pdev);
	while(len > 0){
		if(len >= 4){
			*(uint32_t *)dst = inl(offset_base + offset);
			dst += 4;
			offset += 4;
			len -= 4;
		}else if(len >= 2){
			*(uint16_t *)dst = inw(offset_base + offset);
			dst += 2;
			offset += 2;
			len -= 2;
		}else{
			*(uint8_t*)dst = inb(offset_base + offset);
			dst += 1;
			offset += 1;
			len -= 1;
		}
	}
}

void legacy_write_dev_config(virtio_pci_dev *pdev, uint32_t offset, const void *src, int len){
	uint16_t offset_base = pdev->pci->iobase + VIRTIO_PCI_CONFIG(pdev);
	while(len > 0){
		if(len >= 4){
			outl(offset_base + offset, *(uint32_t *)src);
			src += 4;
			offset += 4;
			len -= 4;
		}else if(len >= 2){
			outw(offset_base + offset, *(uint16_t *)src);
			src += 2;
			offset += 2;
			len -= 2;
		}else{
			outb(offset_base + offset, *(uint8_t*)src);
			src += 1;
			offset += 1;
			len -= 1;
		}
	}
}
uint8_t legacy_get_status(virtio_pci_dev *pdev){
	return inb(pdev->pci->iobase + VIRTIO_PCI_STATUS);
}
void legacy_set_status(virtio_pci_dev *pdev, uint8_t status){
	outb(pdev->pci->iobase + VIRTIO_PCI_STATUS,status);
}

uint64_t legacy_get_features(virtio_pci_dev *pdev){
	return (uint64_t)inl(pdev->pci->iobase + VIRTIO_PCI_HOST_FEATURES);
}

void legacy_set_features(virtio_pci_dev *pdev, uint64_t features){
	outl(pdev->pci->iobase + VIRTIO_PCI_GUEST_FEATURES,(uint32_t)features);
}

int legacy_features_ok(virtio_pci_dev *pdev){
	return 0;
}
uint8_t legacy_get_isr(virtio_pci_dev *pdev){
	return inb(pdev->pci->iobase + VIRTIO_PCI_ISR);
}


uint16_t legacy_set_config_irq(virtio_pci_dev *pdev, uint16_t vec){
	return 0;
}
uint16_t legacy_set_queue_irq(virtio_pci_dev *pdev, uint16_t vq, uint16_t vec){
	return 0;
}
uint16_t legacy_get_queue_num(virtio_pci_dev *pdev, uint16_t queue_id){
	outw(pdev->pci->iobase + VIRTIO_PCI_QUEUE_SEL,queue_id);
	return inw(pdev->pci->iobase + VIRTIO_PCI_QUEUE_NUM);
}

int legacy_setup_queue(virtio_pci_dev *pdev, virt_queue* queue){
	outw(pdev->pci->iobase + VIRTIO_PCI_QUEUE_SEL,queue->idx);
	outl(pdev->pci->iobase + VIRTIO_PCI_QUEUE_PFN,(uint32_t)queue->base_addr >> 12);
	return 0;
}

void legacy_del_queue(virtio_pci_dev *pdev, virt_queue* queue){

}

void legacy_notify_queue(virtio_pci_dev *pdev, virt_queue* queue){
	outw(pdev->pci->iobase + VIRTIO_PCI_QUEUE_NOTIFY, queue->idx);
}

void legacy_intr_detect(virtio_pci_dev *pdev){

}
int legacy_dev_close(virtio_pci_dev *pdev){
	return 0;
}


void modern_read_dev_config(virtio_pci_dev *pdev, uint32_t offset, void *dst, int len){

}

void modern_write_dev_config(virtio_pci_dev *pdev, uint32_t offset, const void *src, int len){

}

void modern_set_status(virtio_pci_dev *pdev, uint8_t status){
	pdev->common_cfg->device_status = status;
}
uint8_t modern_get_status(virtio_pci_dev *pdev){
	return pdev->common_cfg->device_status;
}

uint64_t modern_get_features(virtio_pci_dev *pdev){
	pdev->common_cfg->guest_feature_select = 0;
	uint64_t low = pdev->common_cfg->device_feature;
	pdev->common_cfg->guest_feature_select = 0;
	uint64_t high = pdev->common_cfg->device_feature;
	return (high << 32) | low;
}

void modern_set_features(virtio_pci_dev *pdev, uint64_t features){
	pdev->common_cfg->guest_feature_select = 0;
	pdev->common_cfg->guest_feature = features&0xffffffff;
	pdev->common_cfg->guest_feature_select = 1;
	pdev->common_cfg->guest_feature = features>>32;
	
}
int modern_features_ok(virtio_pci_dev *pdev){
	return 0;
}
uint8_t modern_get_isr(virtio_pci_dev *pdev){
	return *pdev->isr;
}

uint16_t modern_set_config_irq(virtio_pci_dev *pdev, uint16_t vec){
	return 0;
}
uint16_t modern_set_queue_irq(virtio_pci_dev *pdev, uint16_t vq, uint16_t vec){
	return 0;
}
uint16_t modern_get_queue_num(virtio_pci_dev *pdev, uint16_t queue_id){
	pdev->common_cfg->queue_select = queue_id;
	return pdev->common_cfg->queue_size;
}
int modern_setup_queue(virtio_pci_dev *pdev, virt_queue* vq){
	pdev->common_cfg->queue_select = vq->idx;
	pdev->common_cfg->queue_desc = (uint64_t)(uint32_t)vq->base_addr;
	pdev->common_cfg->queue_avail = (uint64_t)(uint32_t)vq->available;
	pdev->common_cfg->queue_used = (uint64_t)(uint32_t)vq->used;
	uint16_t notify_off = pdev->common_cfg->queue_notify_off;
	vq->notify_addr = (uint16_t*)((uint8_t*)pdev->notify_base + notify_off*pdev->notify_off_multiplier);
	pdev->common_cfg->queue_enable = 1;
	return 0;
}
void modern_del_queue(virtio_pci_dev *pdev, virt_queue* vq){
	
}
static inline int virtio_with_feature(virtio_pci_dev* pdev,uint16_t bit){
	return (pdev->guest_feature & (1<<bit)) != 0;
}
void modern_notify_queue(virtio_pci_dev *pdev, virt_queue* vq){
	if (!virtio_with_feature(pdev, VIRTIO_F_NOTIFICATION_DATA)) {
		*vq->notify_addr = vq->idx;
		return;
	}
	// abort("wtf");
}
void modern_intr_detect(virtio_pci_dev *pdev){
	
}
int modern_dev_close(virtio_pci_dev *pdev){
	return 0;
}

// use IO port
const struct virtio_ops legacy_ops = {
	.read_dev_cfg	= legacy_read_dev_config,
	.write_dev_cfg	= legacy_write_dev_config,
	.get_status	= legacy_get_status,
	.set_status	= legacy_set_status,
	.get_features	= legacy_get_features,
	.set_features	= legacy_set_features,
	.get_isr	= legacy_get_isr,
	.set_config_irq	= legacy_set_config_irq,
	.set_queue_irq  = legacy_set_queue_irq,
	.get_queue_num	= legacy_get_queue_num,
	.setup_queue	= legacy_setup_queue,
	.del_queue	= legacy_del_queue,
	.notify_queue	= legacy_notify_queue,
	.intr_detect	= legacy_intr_detect,
	.dev_close	= legacy_dev_close,
};
// use mmio
const struct virtio_ops modern_ops = {
	.read_dev_cfg	= modern_read_dev_config,
	.write_dev_cfg	= modern_write_dev_config,
	.get_status	= modern_get_status,
	.set_status	= modern_set_status,
	.get_features	= modern_get_features,
	.set_features	= modern_set_features,
	// .features_ok	= modern_features_ok,
	.get_isr	= modern_get_isr,
	.set_config_irq	= modern_set_config_irq,
	.set_queue_irq  = modern_set_queue_irq,
	.get_queue_num	= modern_get_queue_num,
	.setup_queue	= modern_setup_queue,
	.del_queue	= modern_del_queue,
	.notify_queue	= modern_notify_queue,
	.intr_detect	= modern_intr_detect,
	.dev_close	= modern_dev_close,
};

