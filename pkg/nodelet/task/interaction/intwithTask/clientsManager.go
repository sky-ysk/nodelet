package intwithTask

import (
	"sync"

	"google.golang.org/grpc"
	"hit.edu/framework/pkg/component-base/logs"
	pb "hit.edu/framework/pkg/nodelet/task/interaction/intwithTask/proto"
)

//管理Taskexporter与所有任务的grpc连接  （grpc客户端）

type GrpcClient struct {
	grpcClient pb.TaskIntentClient
	conn       *grpc.ClientConn
}

type ClientsManager struct {
	grpcClients map[string]GrpcClient
	mu          sync.RWMutex
}

var (
	instance *ClientsManager
	once     sync.Once
)

func GetInstance() *ClientsManager {
	once.Do(func() {
		instance = &ClientsManager{
			grpcClients: make(map[string]GrpcClient),
		}
	})
	return instance
}

func (cm *ClientsManager) MakeGrpcClient(grpcClient pb.TaskIntentClient, conn *grpc.ClientConn) GrpcClient {
	return GrpcClient{grpcClient, conn}
}

func (cm *ClientsManager) AddGrpcClient(taskId string, grpcClient GrpcClient) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.grpcClients[taskId] = grpcClient
}
func (cm *ClientsManager) RemoveGrpcClient(taskId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, ok := cm.grpcClients[taskId]; ok {
		delete(cm.grpcClients, taskId)
	} else {
		logs.Error("grpc client not exist")
	}
}
func (cm *ClientsManager) GetGrpcClient(taskId string) pb.TaskIntentClient {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if _, ok := cm.grpcClients[taskId]; ok {
		return cm.grpcClients[taskId].grpcClient
	}
	logs.Error("grpc client not exist")
	return nil
}
