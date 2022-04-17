from pwn import *
from data import dic,questions
context.log_level='debug'

R.<x>=PolynomialRing(GF(2))
g=x^8+x^7+x^6+x^4+1

def n2p(a):          
     a=bin(a)[2:]              
     p=0                    
     for i in range(len(a)):                                        
         if a[len(a)-i-1]=='1':
             p+=x^i     
     return p

def enc(i):
     m=n2p(i)
     c=(x^8)*m+(((x^8)*m)%g)
     return "".join([str(int(i in c.exponents())) for i in range(15)])[::-1]

dic=[]
for i in range(2**7):
	dic.append(enc(i))

part="( C0 == {} and C1 == {} and C2 == {} and C3 == {} and C4 == {} and C5 == {} and C6 == {} ) "
questions=[]
for i in range(15):
	question=""
	nums=[j[:7] for j in dic if j[i]=='1']
	assert(len(nums)==64)
	for j in range(64):
		n=nums[j]
		question+=part.format(n[0],n[1],n[2],n[3],n[4],n[5],n[6])+"or "
	questions.append(question[:-4])

def dif(a,b):
	cnt=0
	for i in range(len(a)):
		if a[i]!=b[i]:
			cnt+=1
	return cnt

r=process(['python3','patches.py'])
for i in range(50):
	msg=''
	for j in range(15):
		r.sendlineafter("es:",questions[j])
		r.recvuntil("answers:")
		if 'True' in r.recvuntil("!"):
			msg+='1'
		else:
			msg+='0'
	for j in range(128):
		if(dif(msg,dic[j])<=2):
			break
	ans=""
	for k in dic[j][:7]:
		ans+=k+" "
	r.sendlineafter("chests:",ans)
r.interactive()
		
