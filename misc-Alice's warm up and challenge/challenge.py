import os
from PIL import Image
import torch
import torch.nn as nn
import torch.nn.functional as F
from torchvision import transforms

class AliceNet2(nn.Module):
    def __init__(self):
        super(AliceNet2, self).__init__()
        self.conv = \
            nn.Sequential(
            nn.Conv2d(3,12,kernel_size=5,padding=2,stride=2),
            nn.Sigmoid(),
            nn.Conv2d(12,12,kernel_size=5,padding=2, stride=2),
            nn.Sigmoid(),
            nn.Conv2d(12,12,kernel_size=5,padding=2,stride=1),
            nn.Sigmoid(),
            nn.Conv2d(12,12,kernel_size=5,padding=2,stride=1),
            nn.Sigmoid(),
        )
        self.fc = \
            nn.Sequential(
            nn.Linear(768, 200)
        )

    def forward(self, x):
        x = self.conv(x)
        x = x.view(x.size(0), -1)
        x = self.fc(x)
        return x

def getonehot(label, num_classes=200):
    label = torch.unsqueeze(label, 1)
    onehot = torch.zeros(label.size(0), num_classes, device=label.device)
    onehot.scatter_(1, label, 1)
    return onehot

def criterion(pred_y, grand_y):
    # This is the Cross entropy loss function
    tmptensor=torch.mean(
        torch.sum(
            - grand_y * F.log_softmax(pred_y, dim=-1), 1
        ))
    return tmptensor

ts1 = transforms.Compose([transforms.Resize(32),transforms.CenterCrop(32),transforms.ToTensor()])
ts2 = transforms.ToPILImage()

my_device = "cpu"
if torch.cuda.is_available():
    my_device = "cuda"

Net = torch.load('./Net.model').to(my_device)
flagpath='./flag_img/'
imgnames=os.listdir(flagpath)
labellist=[]
outpath='./grad/'

for i in range(len(imgnames)):
    tmpname=imgnames[i]
    tmpnum=int(tmpname.split('_')[0])
    tmppath=flagpath+tmpname
    Bob_data = Image.open(tmppath)
    Bob_data = ts1(Bob_data).to(my_device)
    Bob_data = Bob_data.view(1, *Bob_data.size())
    Bob_label = torch.Tensor([0]).long().to(my_device)
    Bob_label = Bob_label.view(1, )
    Bob_one_hot = getonehot(Bob_label, num_classes=200)
    out = Net(Bob_data)
    y = criterion(out, Bob_one_hot)
    Bob_Grad = torch.autograd.grad(y, Net.parameters())
    torch.save(Bob_Grad, outpath+str(tmpnum)+'.tensor')