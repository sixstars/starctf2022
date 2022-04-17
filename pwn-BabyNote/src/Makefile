# apt-get install musl-tools
.PHONY: all clean

CFLAGS = -fPIC -pie -g
# CFLAGS += -ggdb3

all: babyheap

babyheap: babyheap.c
	musl-gcc $(CFLAGS) $< -o $@

clean:
	rm babyheap
