/*
 * virtio.h
 * Copyright (C) 2021 mac <hzshang15@gmail.com>
 *
 * Distributed under terms of the MIT license.
 */

#ifndef VIRTIO_H
#define VIRTIO_H
#include <pci.h>
#include <virtio_pci.h>
#include <virtio_dev.h>

//Virtual I/O Device (VIRTIO) Version 1.0, Spec 4, section 5.1.3:  Feature bits
#define VIRTIO_CSUM                 0
#define VIRTIO_GUEST_CSUM           1
#define VIRTIO_CTRL_GUEST_OFFLOADS  2
#define VIRTIO_MAC                  5
#define VIRTIO_GUEST_TSO4           7
#define VIRTIO_GUEST_TSO6           8
#define VIRTIO_GUEST_ECN            9
#define VIRTIO_GUEST_UFO            10
#define VIRTIO_HOST_TSO4            11
#define VIRTIO_HOST_TSO6            12
#define VIRTIO_HOST_ECN             13
#define VIRTIO_HOST_UFO             14
#define VIRTIO_MRG_RXBUF            15
#define VIRTIO_STATUS               16
#define VIRTIO_CTRL_VQ              17
#define VIRTIO_CTRL_RX              18
#define VIRTIO_CTRL_VLAN            19
#define VIRTIO_CTRL_RX_EXTRA        20
#define VIRTIO_GUEST_ANNOUNCE       21
#define VIRTIO_MQ                   22
#define VIRTIO_CTRL_MAC_ADDR        23
#define VIRTIO_EVENT_IDX            29
#define VIRTIO_F_NOTIFICATION_DATA  38

#define VIRTIO_F_RING_PACKED        34
#define VIRTIO_F_VERSION_1      32


#define VIRTIO_ACKNOWLEDGE 1
#define VIRTIO_DRIVER 2
#define VIRTIO_FAILED 128
#define VIRTIO_FEATURES_OK 8
#define VIRTIO_DRIVER_OK 4
#define VIRTIO_DEVICE_NEEDS_RESET 64


#define VIRTIO_NET_HDR_F_NEEDS_CSUM    1
#define VIRTIO_NET_HDR_GSO_NONE        0
#define VIRTIO_NET_HDR_GSO_TCPV4       1
#define VIRTIO_NET_HDR_GSO_UDP         3
#define VIRTIO_NET_HDR_GSO_TCPV6       4
#define VIRTIO_NET_HDR_GSO_ECN         0x80

#define PAGE_COUNT(x) ((x+0xFFF)>>12)
#define PAGE_ALIGN(x) ((((uint32_t)(x)+0xfff)>>12)<<12)

#define DISABLE_FEATURE(v,feature) v &= ~(1ULL<<feature)
#define ENABLE_FEATURE(v,feature) v |= (1<<feature)
#define HAS_FEATURE(v,feature) (v & (1<<feature))

#define VIRTQ_USED_F_NO_NOTIFY  1


void virtio_disable_interrupts(virt_queue* vq);
void virtio_enable_interrupts(virt_queue* vq);
void virtio_fill_buffer(virtio_device* vdev, uint16_t queue, virtq_desc* desc_chain, uint32_t count,uint32_t copy);
void setup_virtqueue(virtio_device* vdev,int idx);
void virtio_recv_buffer(virtio_device* vdev, uint16_t queue,uint8_t*,uint32_t*);

#endif /* !VIRTIO_H */
