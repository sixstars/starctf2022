#ifndef __VIRTIO_GPU_H
#define __VIRTIO_GPU_H

#include <x86.h>

#define VIRTIO_GPU_EVENT_DISPLAY (1 << 0) 
 
struct virtio_gpu_config { 
        uint32_t events_read; 
        uint32_t events_clear; 
        uint32_t num_scanouts; 
        uint32_t reserved; 
};
enum virtio_gpu_ctrl_type { 
 
        /* 2d commands */ 
        VIRTIO_GPU_CMD_GET_DISPLAY_INFO = 0x0100, 
        VIRTIO_GPU_CMD_RESOURCE_CREATE_2D, 
        VIRTIO_GPU_CMD_RESOURCE_UNREF, 
        VIRTIO_GPU_CMD_SET_SCANOUT, 
        VIRTIO_GPU_CMD_RESOURCE_FLUSH, 
        VIRTIO_GPU_CMD_TRANSFER_TO_HOST_2D, 
        VIRTIO_GPU_CMD_RESOURCE_ATTACH_BACKING, 
        VIRTIO_GPU_CMD_RESOURCE_DETACH_BACKING, 
        VIRTIO_GPU_CMD_GET_CAPSET_INFO, 
        VIRTIO_GPU_CMD_GET_CAPSET, 
        VIRTIO_GPU_CMD_GET_EDID, 
 
        /* cursor commands */ 
        VIRTIO_GPU_CMD_UPDATE_CURSOR = 0x0300, 
        VIRTIO_GPU_CMD_MOVE_CURSOR, 
 
        /* success responses */ 
        VIRTIO_GPU_RESP_OK_NODATA = 0x1100, 
        VIRTIO_GPU_RESP_OK_DISPLAY_INFO, 
        VIRTIO_GPU_RESP_OK_CAPSET_INFO, 
        VIRTIO_GPU_RESP_OK_CAPSET, 
        VIRTIO_GPU_RESP_OK_EDID, 
 
        /* error responses */ 
        VIRTIO_GPU_RESP_ERR_UNSPEC = 0x1200, 
        VIRTIO_GPU_RESP_ERR_OUT_OF_MEMORY, 
        VIRTIO_GPU_RESP_ERR_INVALID_SCANOUT_ID, 
        VIRTIO_GPU_RESP_ERR_INVALID_RESOURCE_ID, 
        VIRTIO_GPU_RESP_ERR_INVALID_CONTEXT_ID, 
        VIRTIO_GPU_RESP_ERR_INVALID_PARAMETER, 
}; 
#define VIRTIO_GPU_FLAG_FENCE (1 << 0) 
struct virtio_gpu_ctrl_hdr { 
        uint32_t type; 
        uint32_t flags; 
        uint64_t fence_id; 
        uint32_t ctx_id; 
        uint32_t padding;
};
#define VIRTIO_GPU_MAX_SCANOUTS 16 
struct virtio_gpu_rect { 
        uint32_t x; 
        uint32_t y; 
        uint32_t width; 
        uint32_t height; 
};
struct virtio_gpu_resp_display_info { 
        struct virtio_gpu_ctrl_hdr hdr; 
        struct virtio_gpu_display_one { 
                struct virtio_gpu_rect r; 
                uint32_t enabled; 
                uint32_t flags; 
        } pmodes[VIRTIO_GPU_MAX_SCANOUTS]; 
};
struct virtio_gpu_cursor_pos { 
        uint32_t scanout_id; 
        uint32_t x; 
        uint32_t y; 
        uint32_t padding; 
}; 
struct virtio_gpu_update_cursor { 
        struct virtio_gpu_ctrl_hdr hdr; 
        struct virtio_gpu_cursor_pos pos; 
        uint32_t resource_id; 
        uint32_t hot_x; 
        uint32_t hot_y; 
        uint32_t padding; 
};
enum virtio_input_config_select { 
  VIRTIO_INPUT_CFG_UNSET      = 0x00, 
  VIRTIO_INPUT_CFG_ID_NAME    = 0x01, 
  VIRTIO_INPUT_CFG_ID_SERIAL  = 0x02, 
  VIRTIO_INPUT_CFG_ID_DEVIDS  = 0x03, 
  VIRTIO_INPUT_CFG_PROP_BITS  = 0x10, 
  VIRTIO_INPUT_CFG_EV_BITS    = 0x11, 
  VIRTIO_INPUT_CFG_ABS_INFO   = 0x12, 
}; 

struct virtio_input_absinfo { 
  uint32_t  min; 
  uint32_t  max; 
  uint32_t  fuzz; 
  uint32_t  flat; 
  uint32_t  res; 
}; 
 
struct virtio_input_devids { 
  uint16_t  bustype; 
  uint16_t  vendor; 
  uint16_t  product; 
  uint16_t  version; 
}; 
 
struct virtio_input_config { 
  uint8_t    select; 
  uint8_t    subsel; 
  uint8_t    size; 
  uint8_t    reserved[5]; 
  union { 
    char string[128]; 
    uint8_t   bitmap[128]; 
    struct virtio_input_absinfo abs; 
    struct virtio_input_devids ids; 
  } u; 
};


int virtio_gpu_init(virtio_device* vdev);

/*
virtiogpu use two queue

0 controlq - queue for sending control commands
1 cursorq - queue for sending cursor updates

*/


#endif
