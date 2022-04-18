#!/bin/sh
# arg img bootloader kernel

dd if=/dev/zero of=$1 bs=512 count=200 
mkfs.fat $1 >>/dev/null 2>&1
(
echo n # Add a new partition
echo p # Primary partition
echo 2 # Partition number
echo 5 # First sector (Accept default: 1)
echo 100 # Last sector (Accept default: varies)
echo a
echo w # Write changes
) | fdisk  $1 >>/dev/null 2>&1

dd if=$2 of=$1  conv=notrunc 
dd if=$3 of=$1 seek=5 bs=512 conv=notrunc

