from pwn import *

context.log_level='debug'


#c = process("../task")
c = remote('127.0.0.1', 10001)

k0 = 'U"\x8eJsx\xfc\xd2\x9b\x87\xb8ME\x80\xc0\x0c)\xe1?ic\xf1\xe5-S\xffR\xfe\xc2\xdb$\x88'
k0 += '\x00'*0x20
k0 += "\x40"
k0 += "\x31"
c.send(k0.ljust(0x101))
#[245, 121, 67, 242, 248, 143, 232, 17, 196, 162, 251, 139, 64, 200, 32, 88, 106, 11, 255, 73, 51, 136, 153, 225, 1, 48, 105, 6, 140, 15, 121, 112]
#c.recv()

p1 = "\x00"
p1 += "\x41"
c.send(p1.ljust(0x101))

r = c.recvrepeat(timeout=1)

t = process("task2")
t.send(r)
res = t.recvrepeat(timeout=1)
t.close()
p = res.find('flag')
p2 = res.find("\x81\x82\x83\x84", p+1)
fname = res[p:p2]
print fname

t = process("task2")
t.send(fname)
res = t.recvrepeat(timeout=1)[:len(fname)]
t.close()
for i in range(len(res)):
    assert ord(res[i]) != i

p3 = res[:]
p3 += chr(len(res))
p3 += "\x42"

c.send(k0.ljust(0x101))
c.send(p3.ljust(0x101))

r = c.recvrepeat(timeout=1)

t = process("task2")
t.send(res)
t.recvrepeat(timeout=1)
t.send(r)
res = t.recvrepeat(timeout=1)
t.close()
print res

c.interactive()
