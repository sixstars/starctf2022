#include <virtio.h>
#include <pci.h>
#include <physical_page.h>
#include <libcc.h>
#include <picirq.h>
#include <x86.h>
#include <virtio_pci.h>
#include <virtio_net.h>
#include <virtio_ops.h>
#include <virtio.h>
#include <virtio_dev.h>

#include "ip.h"
virtio_device* network_card;
void network_card_setup(virtio_device* vdev){

    virt_queue* rx = &vdev->queue[0]; // Receive
    virt_queue* tx = &vdev->queue[1]; // Send
    rx->chunk_size = FRAME_SIZE;
    rx->available->index = 0;

    virtq_desc buffer;
    buffer.length = FRAME_SIZE;
    buffer.flags = VIRTIO_DESC_FLAG_WRITE_ONLY;
    buffer.address = 0;
    for(int i=0;i<rx->queue_size;i++){
        virtio_fill_buffer(vdev, 0, &buffer, 1,1);
    }
    tx->available->index = 0;
    tx->chunk_size = FRAME_SIZE;

    vdev->pdev.ops->notify_queue(&vdev->pdev,tx);

    // PCI enable
    void virtionet_handler(struct trapframe* trap);
    virtio_enable_interrupts(rx);
    rx->used->flags = 1;
    rx->available->flags = 0;
    tx->used->flags = 0;
    tx->available->flags = 0;
    network_card = vdev;
    register_intr_handler(vdev->pdev.irq + IRQ_OFFSET,virtionet_handler);
    pic_enable(vdev->pdev.irq);
    PCI_enableBusmaster(vdev->pdev.pci);
}


int network_card_init(virtio_device* vdev){
    virtio_pci_dev* pdev = &vdev->pdev;
    // reset
    pdev->ops->set_status(pdev,0);

    uint8_t c = VIRTIO_ACKNOWLEDGE;
    pdev->ops->set_status(pdev,c);
    c |= VIRTIO_DRIVER;
    pdev->ops->set_status(&vdev->pdev,c);
    uint64_t device_feature = pdev->ops->get_features(pdev);

    DISABLE_FEATURE(device_feature,VIRTIO_CTRL_VQ);
    DISABLE_FEATURE(device_feature,VIRTIO_GUEST_TSO4);
    DISABLE_FEATURE(device_feature,VIRTIO_GUEST_TSO6);
    DISABLE_FEATURE(device_feature,VIRTIO_GUEST_UFO);
    DISABLE_FEATURE(device_feature,VIRTIO_EVENT_IDX);
    DISABLE_FEATURE(device_feature,VIRTIO_MRG_RXBUF);

    ENABLE_FEATURE(device_feature,VIRTIO_CSUM);

    pdev->ops->set_features(pdev,device_feature);
    pdev->guest_feature = device_feature;
    // debug("current device feature: 0x%08x\n",device_feature);
    c |= VIRTIO_FEATURES_OK;
    pdev->ops->set_status(pdev,c);

    uint8_t virtio_status = pdev->ops->get_status(pdev);
    if((virtio_status&VIRTIO_FEATURES_OK) == 0){
        printf("feature is not ok\n");
        return 0;
    }
    for(int i=0;i<QUEUE_COUNT;i++){
        setup_virtqueue(vdev,i);
    }
    c |= VIRTIO_DRIVER_OK;
    pdev->ops->set_status(pdev,c);
    virtio_status = pdev->ops->get_status(pdev);
    if (virtio_status & VIRTIO_FAILED){
        printf("virtio init failed\n");
        return 0;
    }
    return 1;
}

uint8_t outnetbuf[FRAME_SIZE];
uint32_t outnetsize;
uint8_t innetbuf[FRAME_SIZE];
uint32_t innetsize;

void virtionet_handler(struct trapframe* trap){
    virtio_device* vdev = network_card;
    if(vdev == NULL){
        printf("network card is null\n");
        return;
    }
    uint8_t isr = vdev->pdev.ops->get_isr(&vdev->pdev);
    if(isr != 1)
        return;

    virt_queue* vq = &vdev->queue[0];
    virtio_disable_interrupts(vq);
    memset(innetbuf,0,sizeof(innetbuf));
    virtio_recv_buffer(vdev,0,innetbuf,&innetsize);
//    network_main(innetbuf + 12,innetsize - 12,outnetbuf,&outnetsize);

    virtq_desc buffer;
    buffer.length = FRAME_SIZE;
    buffer.flags = VIRTIO_DESC_FLAG_WRITE_ONLY;
    buffer.address = 0;
    virtio_fill_buffer(vdev, 0, &buffer, 1,1);
    virtio_enable_interrupts(vq);

    return;
}


void network_send_packet(uint8_t* pkt,size_t length){
    virtio_device* vdev = network_card;
    virtq_desc desc[2];
    virtio_net_hdr hr;
    memset(&hr,0,sizeof(hr));

    desc[0].address = (uint64_t)(uint32_t)&hr;
    desc[0].length = sizeof(hr);
    desc[0].flags = 0;
    desc[1].address = (uint64_t)(uint32_t)pkt;
    desc[1].length = length;
    desc[1].flags = 0;
    virtio_fill_buffer(vdev, 1, desc, 2,1);
    virtio_enable_interrupts(&network_card->queue[0]);
}

uint8_t mymac[6] = {1,2,3,4,5,6};
uint8_t dstmac[6];
uint32_t my_ip = 0x0a0a0a0a;
uint32_t dst_ip;

void network_main(uint8_t* ipbuf,int size,uint8_t* outbuf,uint32_t* outsize_ptr){
    if (size < 14){
        *outsize_ptr = 0;
        return;
    }
    EthFrame* frame = (EthFrame*) ipbuf;
    uint8_t buffer[0x200];
    uint32_t tmp_size;
    switch(frame->type){
        case ARP:
            arpmain(&frame->arp,size - 14,buffer,&tmp_size);
            break;
        case IP:
            ipmain(&frame->ip,size - 14,buffer,&tmp_size);
            break;
        default:
            tmp_size = 0;
            break;
    }
    if(tmp_size){
        EthFrame* out = (EthFrame*)outbuf;
        memcpy(out->dst_mac,dstmac,6);
        memcpy(out->src_mac,mymac,6);
        out->type = frame->type;
        memcpy(&out->ip,buffer,tmp_size);
        *outsize_ptr = tmp_size + 14;
    }else{
        *outsize_ptr = 0;
    }
    return;
}

void arpmain(ARPFrame* arp,int size,uint8_t* buf,uint32_t* size_ptr){
    if(arp->ar_hrd != htons(1) || arp->ar_pro != htons(0x800) || arp->ar_op != htons(1)) {
        *size_ptr = 0;
        return;
    }
    //handler ARP request
    if(arp->ar_tip != my_ip){
        *size_ptr = 0;
        return;
    }
    dst_ip = arp->ar_sip;
    memcpy(dstmac,arp->ar_sha,6);

    ARPFrame* tmp = (ARPFrame*)buf;
    tmp->ar_hrd = htons(1);
    tmp->ar_pro = htons(0x800);
    tmp->ar_hln = 6;
    tmp->ar_pln = 4;
    tmp->ar_op = htons(2);
    memcpy(tmp->ar_sha,mymac,6);
    tmp->ar_sip =  my_ip;
    memcpy(tmp->ar_tha,dstmac,6);
    tmp->ar_tip = dst_ip;
    *size_ptr = sizeof(ARPFrame);
    return;
}

void ipmain(IPFrame* ip,int size,uint8_t* buf,uint32_t* size_ptr){
    uint16_t len = htons(ip->len);
    if(size <= 20 ||
            ip->proto != 1 ||
            ip->daddr != my_ip ||
            ip->saddr != dst_ip ||
            ip->version != 4 ||
            len <= ip->ihl*4){
        *size_ptr = 0;
        return;
    }
    uint16_t icmpsize = len - ip->ihl * 4;
    ICMPFrame* icmp = (ICMPFrame*)(ip->ihl * 4 + (uint8_t*)ip);
    if(icmpsize < 4 || icmp->type != ECHO){
        *size_ptr = 0;
        return;
    }

    ICMPFrame* tmp = (ICMPFrame*)((uint8_t*)buf + 20);
    tmp->type = REPLY;
    tmp->code = icmp->code;
    tmp->csum = icmp->csum + 8;
    memcpy(tmp->data,icmp->data,icmpsize - 4);

    IPFrame* iphdr = (IPFrame*)buf;
    iphdr->version = 4;
    iphdr->ihl = 5;
    iphdr->flags = 0;
    iphdr->tos = 0;
    iphdr->len = ip->len;
    iphdr->flags = 0;
    iphdr->ttl = ip->ttl;
    iphdr->proto = 1;
    iphdr->saddr = my_ip;
    iphdr->daddr = dst_ip;
    iphdr->csum = 0;
    *size_ptr = len;

    void compute_ip_checksum(IPFrame* iphdrp);
    compute_ip_checksum(iphdr);
}




/* Compute checksum for count bytes starting at addr, using one's complement of one's complement sum*/
uint16_t compute_checksum(unsigned short *addr, unsigned int count) {
    register unsigned long sum = 0;
    while (count > 1) {
        sum += * addr++;
        count -= 2;
    }
    //if any bytes left, pad the bytes and add
    if(count > 0) {
        sum += ((*addr)&htons(0xFF00));
    }
    //Fold sum to 16 bits: add carrier to result
    while (sum>>16) {
        sum = (sum & 0xffff) + (sum >> 16);
    }
    //one's complement
    sum = ~sum;
    return ((unsigned short)sum);
}

/* set ip checksum of a given ip header*/
void compute_ip_checksum(IPFrame* iphdrp){
    iphdrp->csum = 0;
    iphdrp->csum = compute_checksum((unsigned short*)iphdrp, iphdrp->ihl * 4);
}


