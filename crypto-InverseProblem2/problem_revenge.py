import numpy as np
from secret2 import flag

def getQ(n):
    return np.linalg.qr(np.random.random([n,n]))[0]

def pad(x,N=40,k=256):
    return np.hstack([x,np.random.random(N-len(x))*k])

n=len(flag)
N=40
A=np.hstack([getQ(N)[:,:n]@np.diag(np.logspace(n,1,n))@getQ(n),getQ(N)[:,n:]@np.diag([10**n]*(N-n))@getQ(N-n)])
x=pad(list(flag))
b=A@x
np.savetxt('A2.txt',A)
np.savetxt('b2.txt',b)