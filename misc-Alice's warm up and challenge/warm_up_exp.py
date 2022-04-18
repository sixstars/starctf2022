import torch
from torch import nn
import string
class AliceNet1(nn.Module):
    pass
def char2num(ch):
    tmpset = string.printable[0:36]+'*CTF{ALIZE}'
    tmplen=len(tmpset)
    for i in range(tmplen):
        if(ch==tmpset[i]):
            return i
def dfs(ch,depth,ans):
    ans+=ch
    if(len(ans)==flaglen and ans[-1]=='}'):
        print('flag is:',ans)
    elif(len(ans)==flaglen):
        return
    else:
        tmpi=char2num(ch)
        for i in range(setlen):
            if(flagset[i]==ch):
                continue
            tmpj=char2num(flagset[i])
            if(mymat[tmpi][tmpj]==1.0 and used[tmpj]==False):
                used[tmpj]=True
                dfs(flagset[i],depth+1,ans)
                used[tmpj]=False
savepath='./0bdb74e42cdf4a42923ccf40d2a66313.zip'
net=torch.load(savepath)
print(net)
mymat=net.state_dict()['fc.0.weight'].tolist()
flagset=string.printable[0:36]+'*CTF{ALIZE}'
setlen=len(flagset)
flaglen=10+6
used=[False]*setlen
flag=''
used[char2num('*')]=True
dfs('*',0,flag)