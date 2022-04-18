ping 


There is a stackoverflow vulnerability when processing icmp packet, it copy origin echo request payload to a small buffer([0x200]) in kernel/driver/virtio_net.c:226.


    ICMPFrame* tmp = (ICMPFrame*)((uint8_t*)buf + 20);
    tmp->type = REPLY;
    tmp->code = icmp->code;
    tmp->csum = icmp->csum + 8;
    memcpy(tmp->data,icmp->data,icmpsize - 4);          <------- stackoverflow

    IPFrame* iphdr = (IPFrame*)buf;
    iphdr->version = 4;
    iphdr->ihl = 5;
    iphdr->flags = 0;
    iphdr->tos = 0;
    iphdr->len = ip->len;
    iphdr->flags = 0;




You need to send a icmp packet, control pc, copy flag at a fixed address to icmp reply packet and send it back. Don't forget calc your ip packet checksum.





