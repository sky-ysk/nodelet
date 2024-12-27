from torch import nn
import torch.nn.functional as F
from torch import tanh


class ConvNet(nn.Module):
    def __init__(self):
        super().__init__()
        self.conv1 = nn.Conv2d(1, 6, 5, padding=2)  # 28x28
        self.pool1 = nn.MaxPool2d(2, 2)  # 14x14
        self.conv2 = nn.Conv2d(6, 16, 5)  # 10x10
        self.pool2 = nn.MaxPool2d(2, 2)  # 5x5
        self.conv3 = nn.Conv2d(16, 120, 5)
        self.fc1 = nn.Linear(120, 84)
        self.fc2 = nn.Linear(84, 10)

    def forward(self, x):
        in_size = x.size(0)
        out = self.conv1(x)  # 24
        out = tanh(out)
        out = self.pool1(out)  # 12
        out = self.conv2(out)  # 10
        out = tanh(out)
        out = self.pool2(out)
        out = self.conv3(out)
        out = out.view(in_size, -1)
        out = self.fc1(out)
        out = tanh(out)
        out = self.fc2(out)
        out = F.log_softmax(out, dim=1)
        return out
