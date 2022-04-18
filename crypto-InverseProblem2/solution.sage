import numpy as np

AB=np.loadtxt('A.txt')
b=np.loadtxt('b.txt')
N=50
n=int(np.log10(np.linalg.svd(AB)[1][0]).round())
A=matrix(AB[:,n:].tolist())
B=matrix(AB[:,:n].tolist())
y=vector(b.tolist())
E=(A.T*A)**(-1)
K=A*E*A.T-identity_matrix(N)
M=block_matrix(QQ,[[(K*B).T,zero_matrix(n,1)],[matrix(K*y),matrix([1e20])]])
L=M.LLL()
v=L[0]/M
flag=bytes(-v[:-1]/v[-1]).decode()
print(flag)