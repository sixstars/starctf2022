import hashlib
l=2**252+27742317777372353535851937790883648493
Z=Zmod(l)
PK0='F/Vm9CLcafzgoEIKqjE9TOaPpJTtf0gNg4+mgjkZRq4='.decode('base64')
PK1='+tP/02i3K7k7XAZvVC1fBUoMEpUG3xfM3ZttO11xsxc='.decode('base64')

sign0='mpiry4AquSWIwAG10toNQ3B/c8H0QZ6/AC6waa9PUcqeorIBKeFZJ1J1T4XTZU/V/nYea3KK9On9BT/3u1clDw=='.decode('base64')
sign1='1d9UKtBwRvveUmWdD5rXWHFs2q5TKc29ABDSpfoOue5iB2pj/Vdf6qfnuJAeyr9XayfscZKVzZoNODYSDz+PAg=='.decode('base64')

R0=sign0[:32]
R1=sign1[:32]

S0=Z(int(sign0[32:][::-1].encode('hex'),16))
S1=Z(int(sign1[32:][::-1].encode('hex'),16))
M='a'*1024
sha512=hashlib.sha512()
sha512.update(R0)
sha512.update(PK0)
sha512.update(M)
H0=Z(int(sha512.hexdigest().decode('hex')[::-1].encode('hex'),16))
sha512=hashlib.sha512()
sha512.update(R1)
sha512.update(PK1)
sha512.update(M)
H1=Z(int(sha512.hexdigest().decode('hex')[::-1].encode('hex'),16))

sha512=hashlib.sha512()
sha512.update(M)
sha512.update('\x00'*32)
print(sha512.hexdigest().decode('hex').encode('base64'))
Dr0=Z(int(sha512.hexdigest().decode('hex')[::-1].encode('hex'),16))
sha512=hashlib.sha512()
sha512.update(M)
sha512.update('\x01'+'\x00'*31)
print(sha512.hexdigest().decode('hex').encode('base64'))
Dr1=Z(int(sha512.hexdigest().decode('hex')[::-1].encode('hex'),16))
a=(S0-S1+H1-Dr0+Dr1)/(H0-H1)
print(hex(int(a))[2:-1].decode('hex')[::-1])


