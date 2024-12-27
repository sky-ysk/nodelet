package grpc_client

import (
	"google.golang.org/grpc"
	"hit.edu/framework/pkg/component-base/logs"
	pb "hit.edu/framework/pkg/nodelet/task/interaction/intwithTask/proto"
)

type Client struct {
	ServerIPAndPort string
}

//TaskExporter直接与任务进行通信的client相关的函数接口，被assigner里的函数调用

func NewClient() *Client {
	return &Client{"127.0.0.1:50057"}
}

// 与任务建立连接，返回给taskExporter与任务建立连接的client
func (c *Client) ConnectGrpcServer() (pb.TaskIntentClient, *grpc.ClientConn, bool) {
	conn, err := grpc.Dial(c.ServerIPAndPort, grpc.WithInsecure())
	if err != nil {
		logs.Error("grpc-server error, cannot connect to server")
		return nil, nil, false
	}
	grpcClient := pb.NewTaskIntentClient(conn)
	logs.Info("grpc client created")
	return grpcClient, conn, true
}

// func (c *Client) InitTask() (pb.TaskIntentClient, *grpc.ClientConn, bool) {
// 	//使用ConnectGrpcServer返回的与任务对应的conn，与相对应的grpc连接通信

// }

// func (c *Client) StartTask() (pb.TaskIntentClient, *grpc.ClientConn, bool) {

// }

// func (c *Client) StopTask() (pb.TaskIntentClient, *grpc.ClientConn, bool) {

// }

// func (c *Client) RestartTask() (pb.TaskIntentClient, *grpc.ClientConn, bool) {

// }
