from PIL import Image
import matplotlib.pyplot as plt
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
outpath='./grad/'

torch.manual_seed(0)

for i in range(25):
    original_dy_dx=dy_dx=torch.load(outpath+str(i)+'.tensor')
    dummy_data = torch.randn(1,3,32,32).to(my_device).requires_grad_(True)
    dummy_label = torch.randn(1,200).to(my_device).requires_grad_(True)
    optimizer = torch.optim.LBFGS([dummy_data, dummy_label])
    history = []
    for iters in range(300):
        def closure():
            optimizer.zero_grad()
            pred = Net(dummy_data)
            dummy_onehot_label = F.softmax(dummy_label, dim=-1)
            dummy_loss = criterion(pred,
                                   dummy_onehot_label)
            dummy_dy_dx = torch.autograd.grad(dummy_loss, Net.parameters(), create_graph=True)
            grad_diff = 0
            grad_count = 0
            for gx, gy in zip(dummy_dy_dx, original_dy_dx):
                grad_diff += ((gx - gy) ** 2).sum()
                grad_count += gx.nelement()
            grad_diff.backward()
            return grad_diff

        optimizer.step(closure)
        if iters % 10 == 0:
            current_loss = closure()
            print(iters, "%.4f" % current_loss.item())
        history.append(ts2(dummy_data[0].cpu()))

    plt.figure(figsize=(12, 8))
    for i in range(30):
        plt.subplot(3, 10, i + 1)
        plt.imshow(history[i * 10])
        plt.title("iter=%d" % (i * 10))
        plt.axis('off')
    print("Dummy label is %d." % torch.argmax(dummy_label, dim=-1).item())
    plt.show()