import torch
from torch import nn
import string
import numpy as np

np.random.seed(4061314)

class AliceNet1(nn.Module):

    def __init__(self):
        super(AliceNet1,self).__init__()
        self.fc=nn.Sequential(
            nn.Linear(47,47),
            nn.Linear(47,10),
            nn.Linear(10,1)
        )

    def forward(self,x):
        x=self.fc(x)
        return x

def char2num(ch):
    tmpset = string.printable[0:36]+'*CTF{ALIZE}'
    tmplen=len(tmpset)
    for i in range(tmplen):
        if(ch==tmpset[i]):
            return i

def get_0_1_array(x,y,rate=0.2):
    #得到一个全1矩阵，按照rate=0.5的比率生成新矩阵
    array = np.ones(x*y)
    array = array.reshape(x,y)
    zeros_num = int(array.size * rate)#根据0的比率来得到 0的个数
    new_array = np.ones(array.size,dtype=int)#生成与原来模板相同的矩阵，全为1
    new_array[:zeros_num] = 0 #将一部分换为0
    np.random.shuffle(new_array)#将0和1的顺序打乱
    re_array = new_array.reshape(array.shape)#重新定义矩阵的维度，与模板相同
    return re_array


flagset=string.printable[0:36]+'*CTF{ALIZE}'
set_len=len(flagset)
flag="*CTF{qx1jukznmr}"
inter_flag='qx1jukznmr'
mymat=[[0]*set_len for i in range(set_len)]
for i in range(1,len(flag)):
    tmpi,tmpj=char2num(flag[i-1]),char2num(flag[i])
    #print(flag[i-1],flag[i])
    mymat[tmpi][tmpj]=1
notflag=''
for i in range(len(flagset)):
    if(flagset[i] not in flag):
        notflag+=flagset[i]
layer1='ab0pcg'
for i in range(len(layer1)):
    tmpi,tmpj=char2num('{'),char2num(layer1[i])
    mymat[tmpi][tmpj]=1
layer2='oilw'
for i in range(len(layer2)):
    tmpi,tmpj=char2num('{'),char2num(layer2[i])
    mymat[tmpi][tmpj]=1
    tmpi,tmpj=char2num(layer2[i]),char2num(flag[0-(i+1)])
    mymat[tmpi][tmpj]=1
notlayer=''
for i in range(len(notflag)):
    if(notflag[i] not in layer1 and notflag[i] not in layer2):
        notlayer+=notflag[i]
for i in range(len(layer1)):
    for j in range(len(notlayer)):
        tmpi,tmpj=char2num(layer1[i]),char2num(notlayer[j])
        tmpflag=get_0_1_array(3,3,0.8).tolist()[0][0]
        mymat[tmpi][tmpj]=tmpflag
randmat1=get_0_1_array(len(notlayer),len(notlayer),rate=0.8).tolist()
for i in range(len(notlayer)):
    for j in range(len(notlayer)):
        if(i==j):
            continue
        tmpi,tmpj=char2num(notlayer[i]),char2num(notlayer[j])
        if(randmat1[i][j]==1):
            mymat[tmpi][tmpj]=1

sum=0
cnt=0
for i in range(len(mymat)):
    print(mymat[i])
    for j in range(len(mymat)):
        sum+=mymat[i][j]
        cnt+=1
print('edge num: ',sum,'\nnot num: ',cnt-sum)

net=AliceNet1()
for name in net.state_dict():
    print(name)




print(net.state_dict()['fc.0.weight'])
tmpweight=net.state_dict()
tmpweight['fc.0.weight']=torch.tensor(mymat)
#print(tmpweight['fc.0.weight'])
net.load_state_dict(tmpweight)
print(net.state_dict()['fc.0.weight'])
savepath='./hP76VYJ3ih.zip'
torch.save(net,savepath)
tmpnet=torch.load(savepath)
print(tmpnet.state_dict()['fc.0.weight'])

for name in tmpnet.state_dict():
    print(name)
    print(net.state_dict()[name])