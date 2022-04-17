/*
 * demo.c
 * Copyright (C) 2022 hal <hal@server20>
 *
 * Distributed under terms of the MIT license.
 */

#include <linux/fs.h>
#include <linux/init.h>
#include <linux/file.h>
#include <linux/errno.h>
#include <linux/module.h>
#include <linux/miscdevice.h>
#include <linux/kernel.h>
#include <linux/slab.h>
#include <linux/tty.h>
#include <linux/userfaultfd.h>
#include <linux/cred.h>
#include <linux/proc_fs.h>

#include <asm/string.h>

MODULE_AUTHOR("albanis");
MODULE_LICENSE("GPL");
MODULE_DESCRIPTION("Pwn me :)");

char demo_buf[0x1000];

static int device_open(struct inode *inode, struct file *filp)
{
	// printk(KERN_ALERT "Device opened.\n");
  	return 0;
}

static int device_release(struct inode *inode, struct file *filp)
{
	// printk(KERN_ALERT "Device closed.\n");
  	return 0;
}

static ssize_t device_read(struct file *filp, char *buffer, size_t length, loff_t *offset)
{
    int tmp[32];
    if (length > 0x1000) {
        printk(KERN_WARNING "Buffer overflow detected (%d < %lu)!!\n", 0x1000, length);
        return -EINVAL;
    }

    __memcpy(demo_buf, tmp, length);
    if (copy_to_user(buffer, demo_buf, length))
        return -EINVAL;

	// printk(KERN_ALERT "Device write.\n");
  	return length;
}

static ssize_t device_write(struct file *filp, const char *buf, size_t len, loff_t *off)
{
    int tmp[32];
    if (len > 0x1000) {
        printk(KERN_WARNING "Buffer overflow detected (%d < %lu)!\n", 0x1000, len);
        return -EINVAL;
    }

    if (copy_from_user(demo_buf, buf, len))
        return -EINVAL;

    __memcpy(tmp, demo_buf, len);

	// printk(KERN_ALERT "Device write.\n");
  	return len;
}

static struct proc_ops fops = {
  	.proc_read = device_read,
  	.proc_write = device_write,
  	.proc_open = device_open,
  	.proc_release = device_release
};

struct proc_dir_entry *proc_entry = NULL;

int init_module(void)
{
	proc_entry = proc_create("demo", 0666, NULL, &fops);
  	return 0;
}

void cleanup_module(void)
{
	if (proc_entry) proc_remove(proc_entry);
}

