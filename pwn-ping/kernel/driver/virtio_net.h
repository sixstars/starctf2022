#ifndef VIRTIO_NET_H
#define VIRTIO_NET_H
#define FRAME_SIZE 1526 // including the net_header
#include <trap.h>



typedef struct {
#define VIRTIO_NET_HDR_F_NEEDS_CSUM    1
    uint8_t flags;
#define VIRTIO_NET_HDR_GSO_NONE        0
#define VIRTIO_NET_HDR_GSO_TCPV4       1
#define VIRTIO_NET_HDR_GSO_UDP         3
#define VIRTIO_NET_HDR_GSO_TCPV6       4
#define VIRTIO_NET_HDR_GSO_ECN      0x80
    uint8_t gso_type;
    uint16_t hdr_len;
    uint16_t gso_size;
    uint16_t csum_start;
    uint16_t csum_offset;
    uint16_t num_buffers;
} virtio_net_hdr;

typedef struct {
    uint8_t mac[6];
    uint16_t status;
    uint16_t max_virtqueue_pairs;
    uint16_t mtu;
}__attribute__((packed)) virtio_net_config;
void network_card_setup(virtio_device* vdev);
void virtionet_handler(struct trapframe* trap);
int network_card_init(virtio_device* vdev);

#endif
