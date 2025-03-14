import torch
import torch.nn.functional as F
from torchvision import datasets
from torchvision.transforms import transforms
from torch.utils.data import DataLoader

import task_pb2
import task_pb2_grpc
from concurrent import futures
import grpc

from module import ConvNet

BATCH_SIZE = 256


class TaskIntentServicer(task_pb2_grpc.TaskIntentServicer):
    def init(self, request, context):
        return

    def start(self, request, context):
        print("task start running")
        run()
        # return
        return task_pb2.Result(msg="success", stateCode=0)
    def restart(self, request, context):
        print("restart running")
        return task_pb2.Result(msg="restart", stateCode=0)
    def stop():return
    def getStatus():return
    def store():return
    def restore():return

def train(model, device, train_loader, optimizer, epoch):
    model.train()
    for batch_idx, (data, target) in enumerate(train_loader):
        data, target = data.to(device), target.to(device)
        optimizer.zero_grad()
        output = model(data)
        loss = F.nll_loss(output, target)
        loss.backward()
        optimizer.step()
        if (batch_idx + 1) % 30 == 0:
            print('Train Epoch: {} [{}/{} ({:.0f}%)]\tLoss: {:.6f} '.format(
                epoch, batch_idx * len(data), len(train_loader.dataset), 100. * batch_idx / len(train_loader),
                loss.item()))


def test(model, device, test_loader):
    model.eval()
    test_loss = 0
    correct = 0
    with torch.no_grad():
        for data, target in test_loader:
            data, target = data.to(device), target.to(device)
            output = model(data)
            test_loss += F.nll_loss(output, target, reduction='sum').item()  # 将一批的损失相加
            pred = output.max(1, keepdim=True)[1]  # 找到概率最大的下标
            correct += pred.eq(target.view_as(pred)).sum().item()

    test_loss /= len(test_loader.dataset)
    print('\nTest set: Average loss: {:.4f},Accuracy: {}/{} ({:.0f}%)\n '.format(
        test_loss, correct, len(test_loader.dataset),
        100. * correct / len(test_loader.dataset)))


def run():

    _device = 'cuda' if torch.cuda.is_available() else 'cpu'
    # _device = 'cuda'
    _module = ConvNet().to(_device)
    _optimizer = torch.optim.SGD(params=_module.parameters(), lr=0.01, momentum=0.9)

    _train_loader = torch.utils.data.DataLoader(
        datasets.MNIST('data', train=True, download=True,
                       transform=transforms.Compose([
                           transforms.ToTensor(),
                           transforms.Normalize((0.1307,), (0.3081,))])),
        batch_size=BATCH_SIZE, shuffle=True)
    _test_loader = torch.utils.data.DataLoader(
        datasets.MNIST('data', train=False, transform=transforms.Compose([
            transforms.ToTensor(),
            transforms.Normalize((0.1307,), (0.3081,))
        ])),
        batch_size=BATCH_SIZE, shuffle=True)

    epoches = 1
    for _epoch in range(epoches):
        train(_module, _device, _train_loader, _optimizer, _epoch)
        test(_module, _device, _test_loader)

    #保存训练好的参数文件，后续可能需要改成每多少轮保存一次，当做保存点
    torch.save(_module.state_dict(), 'model.pth')
    print('Saved PyTorch LeNet5 State to model.pth')


def serve():
    print('start grpc server====>')
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    task_pb2_grpc.add_TaskIntentServicer_to_server(TaskIntentServicer(), server)
    server.add_insecure_port('[::]:50056')
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    #serve建立grpc连接
    # serve()
    
    #直接运行单独测试启动可以只使用run()
    run()