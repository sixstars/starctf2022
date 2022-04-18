#include <virtio.h>
#include <pci.h>
#include <physical_page.h>
#include <libcc.h>
#include <picirq.h>
#include <x86.h>
#include <libcc.h>
#include <virtio_pci.h>
#include <virtio_net.h>
#include <virtio_ops.h>

#define QUEUE_LENGTH 0x10
#define VIRTIO_BLK_F_SIZE_MAX 1
#define VIRTIO_BLK_F_SEG_MAX 2
#define VIRTIO_BLK_F_GEOMETRY 4
#define VIRTIO_BLK_F_RO 5
#define VIRTIO_BLK_F_BLK_SIZE 6
#define VIRTIO_BLK_F_FLUSH 9
#define VIRTIO_BLK_F_TOPOLOGY 10
#define VIRTIO_BLK_F_CONFIG_WCE 11
#define VIRTIO_BLK_T_IN           0
#define VIRTIO_BLK_T_OUT          1
#define VIRTIO_BLK_T_FLUSH        4
#define VIRTIO_BLK_S_OK        0
#define VIRTIO_BLK_S_IOERR     1
#define VIRTIO_BLK_S_UNSUPP    2


#define DEVICE_STATUS 0x12
#define DEVICE_FEATURE 0
#define DRIVER_FEATURE 4
#define QUEUE_SELECT 0xe
#define QUEUE_SIZE 0xc
#define QUEUE_ADDR 0x8
#define QUEUE_NOTIFY 0x10
/*
 * 0: Device Features bits 0:31
 * 4: Driver Features bits 0:31
 * 8: Queue Address
 * 12: queue_size
 * 14: queue_select
 * 16: Queue Notify
 * 18: Device Status
 * 19: ISR Status
 */
#define FRAME_SIZE 1526
bool virtio_queue_init(virt_queue* queue,uint16_t port,uint16_t idx){
    outw(port + QUEUE_SELECT,idx);
    uint16_t queue_size = inw(port+QUEUE_SIZE);
    memset(queue,0,sizeof(*queue));
    if(!queue_size || queue_size == 0xffff){
        return false;
    }
    printf("queue size[%d]: %x\n",idx,queue_size);
    uint32_t buffers_size = sizeof(virtq_desc)*queue_size;
    uint32_t available_size = (2 + queue_size )* sizeof(uint16_t);
    uint32_t used_size = sizeof(virtq_ring)*queue_size + 2*sizeof(uint16_t);
    uint32_t page_count = PAGE_COUNT(buffers_size+available_size) + PAGE_COUNT(used_size);
    uint8_t* buf = physical_alloc(page_count<<12,0x1000);

    memset(buf,0,page_count<<12);
    printf("queue buffer addr: %x\n",buf);

    queue->base_addr = buf;
    queue->available = (virtq_avail*)&buf[buffers_size];
    queue->used = (virtq_used*)PAGE_ALIGN(&buf[buffers_size + available_size]);
    outl(port + QUEUE_ADDR,((uint32_t)buf)>>12);
    queue->available->flags = 0;
    queue->inuse = 1;
    queue->queue_size = queue_size;
    return true;
}

void setup_virtqueue(virtio_device* vdev,int idx);

void notify_queue(virtio_device* vdev, uint16_t queue);

void show_device_status(virtio_device* vdev);




void virtio_enable_interrupts(virt_queue* vq)
{
    vq->used->flags = 0;
}

void virtio_disable_interrupts(virt_queue* vq)
{
    vq->used->flags = 1;
}



void setup_virtqueue(virtio_device* vdev,int idx){

    uint16_t queue_size = vdev->pdev.ops->get_queue_num(&vdev->pdev,idx);
    virt_queue* queue = &vdev->queue[idx];
    memset(queue,0,sizeof(*queue));
    if(!queue_size || queue_size == 0xffff){
        return;
    }
    queue->queue_size = queue_size;

    uint32_t buffers_size = sizeof(virtq_desc)*queue_size;
    uint32_t available_size = (2 + queue_size )* sizeof(uint16_t);
    uint32_t used_size = sizeof(virtq_ring)*queue_size + 2*sizeof(uint16_t);
    uint32_t page_count = PAGE_COUNT(buffers_size+available_size) + PAGE_COUNT(used_size);
    uint8_t* buf = physical_alloc(page_count<<12,0x1000);
    memset(buf,0,page_count<<12);
    queue->base_addr = buf;
    queue->available = (virtq_avail*)&buf[buffers_size];
    queue->used = (virtq_used*)PAGE_ALIGN(&buf[buffers_size + available_size]);
    queue->idx = idx;
    vdev->pdev.ops->setup_queue(&vdev->pdev,queue);
    queue->available->flags = 0;
    // queue->used->flags = 0;
    // kprintf(
    //     "queue [%d] size: %x\n"
    //     " desc addr:  %x available addr: %x used addr: %x\n"
    //     " notify addr: %x\n"
    //     ,idx,queue_size,queue->base_addr,queue->available,queue->used,
    //     queue->notify_addr
    //     );
    return;
}


void virtio_fill_buffer(virtio_device* vdev, uint16_t queue, virtq_desc* desc_chain, uint32_t count,uint32_t copy){
    virt_queue* vq = &vdev->queue[queue];
    uint16_t idx = vq->available->index % vq->queue_size;
    uint16_t buf_idx = vq->next_buffer;
    uint16_t next_buf;
    uint8_t* buf = (uint8_t *)(&vq->arena[vq->chunk_size * buf_idx]);
    vq->available->ring[idx] = buf_idx;
    for(int i=0;i<count;i++){
        next_buf = (buf_idx + 1) % vq->queue_size;
        vq->buffers[buf_idx].flags = desc_chain[i].flags;
        if (i != count -1) {
            vq->buffers[buf_idx].flags |= VIRTIO_DESC_FLAG_NEXT;
        }
        vq->buffers[buf_idx].next = next_buf;
        vq->buffers[buf_idx].length = desc_chain[i].length;
        if(copy){
            vq->buffers[buf_idx].address = (uint64_t)(uint32_t)buf;
            if(desc_chain[i].address)
                memcpy(buf, (const void*)(uint32_t)desc_chain[i].address, desc_chain[i].length);
            buf += desc_chain[i].length;
        }else{
            vq->buffers[buf_idx].address = (uint64_t)(uint32_t)desc_chain[i].address;
        }
        buf_idx = next_buf;
    }
    vq->next_buffer = next_buf;
    vq->available->index++;
    vdev->pdev.ops->notify_queue(&vdev->pdev,vq);
}

void virtio_recv_onepkt(virt_queue* vq,uint16_t head,uint8_t **dst){
    uint8_t* dst_buf = *dst;
    uint32_t i = vq->used->ring[head].index;
    uint32_t length = vq->used->ring[head].length;
    // memcpy(dst_buf,vq->buffers[i].address,vq->used->ring[head].length);
    // dst_buf += vq->used->ring[head].length;
    while(1){
        uint32_t current_read = min(vq->buffers[i].length,length);
        memcpy(dst_buf,(const void*)(uint32_t)vq->buffers[i].address,current_read);
        length -= current_read;
        dst_buf += current_read;
        if(length == 0){
            if(vq->buffers[i].flags&VIRTIO_DESC_FLAG_NEXT){
            }
            break;
        }
        if(!(vq->buffers[i].flags&VIRTIO_DESC_FLAG_NEXT)){
            break;
        }
        i = vq->buffers[i].next;
    }
    *dst = dst_buf;
}

void virtio_recv_buffer(virtio_device* vdev, uint16_t queue,uint8_t* output,uint32_t* size){
    virt_queue* vq = &vdev->queue[queue];
    if (vq->last_used_index == vq->used->index){
        return;
    }
    *size = 0;
    while(vq->last_used_index != vq->used->index){
        uint16_t last_idx = vq->last_used_index &(vq->queue_size-1);
        uint8_t* iter = output + *size;
        virtio_recv_onepkt(vq,last_idx,&iter);
        *size += iter - output;
        vq->last_used_index++;
    }
    return;
}


void virtionet_send(void* driver, void *packet, uint16_t length){
    virtio_device* vdev = (virtio_device*)driver;
//    uint32_t virt_size = length + sizeof(virtio_net_hdr);
    virtq_desc desc[2];
    virtio_net_hdr net;
    memset(&net,0,sizeof(net));

    net.flags = VIRTIO_NET_HDR_F_NEEDS_CSUM;
    net.gso_type = 0;
    net.csum_start = 0;
    net.csum_offset = length;
    
    desc[0].flags = 0;
    desc[0].address = (uint32_t)&net;
    desc[0].length = sizeof(virtio_net_hdr);

    desc[1].flags = 0;
    desc[1].address = (uint64_t)(uint32_t)packet;
    desc[1].length = length;
    virtio_fill_buffer(vdev,1,desc,2,1);
}





