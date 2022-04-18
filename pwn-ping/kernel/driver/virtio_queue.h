#ifndef VIRTIO_QUEUE_H
#define VIRTIO_QUEUE_H

typedef struct{
    uint64_t address;
    uint32_t length;
    uint16_t flags;
    uint16_t next;
#define VIRTIO_DESC_FLAG_NEXT           1
#define VIRTIO_DESC_FLAG_WRITE_ONLY     2
#define VIRTIO_DESC_FLAG_INDIRECT       4
} virtq_desc;

typedef struct{
    uint16_t flags;
    uint16_t index;
    uint16_t ring[];
} virtq_avail;

typedef struct {
    uint32_t index;
    uint32_t length;
} virtq_ring;

typedef struct{
    uint16_t flags;
    uint16_t index;
    virtq_ring ring[];
} virtq_used;

typedef struct {
    int queue_size;
    int inuse;
    uint32_t idx;
    union{
        uint8_t* base_addr;
        virtq_desc* buffers; // for the Descriptor Area
    };
    virtq_avail* available; // for the Driver Area
    virtq_used* used;// for the Device Area
    uint16_t* notify_addr;
    uint16_t chunk_size;
    uint16_t last_used_index;
	uint16_t next_buffer;
    uint8_t arena[0x100*1526];
}virt_queue;

#endif
