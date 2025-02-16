package grpc_client

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"hit.edu/framework/pkg/component-base/logs"
	pb "hit.edu/framework/pkg/nodelet/task/interaction/intwithRuntime/proto"
	"time"
)

type RuntimeClient struct {
	RuntimeID       string
	ServerIPAndPort string
	conn            *grpc.ClientConn
	grpcClient      pb.RuntimeIntentClient
}

func NewRuntimeClient(port string, runtimeID string) *RuntimeClient {
	client := &RuntimeClient{RuntimeID: runtimeID, ServerIPAndPort: "127.0.0.1:" + port}
	for {
		success := client.checkConnection()
		if success {
			break
		}
		logs.Debug("try to connect grpc server")
		time.Sleep(time.Millisecond * 500)
	}
	return client
}

func (r *RuntimeClient) checkConnection() bool {
	connState := true
	if r.conn == nil {
		conn, err := grpc.Dial(r.ServerIPAndPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logs.Debugf("failed to connect to grpc server:%v", err)
			connState = false
		} else {
			r.conn = conn
			r.grpcClient = pb.NewRuntimeIntentClient(conn)
		}
	}
	return connState
}

// rpc远程调用服务端启动应用
func (c *RuntimeClient) RunAppStart() (result *pb.Result, err error) {
	if !c.checkConnection() {
		return &pb.Result{}, errors.New("runAppStart: grpc connection failed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err = c.grpcClient.Start(ctx, &pb.StartIntent{})
	for err != nil {
		logs.Debug("retry to runAppStart")
		result, err = c.grpcClient.Start(ctx, &pb.StartIntent{})
	}
	logs.Infof("runAppStart: result:%v", result)
	return result, nil
}

// rpc远程调用服务端保存应用状态
func (c *RuntimeClient) RunAppStore() (result *pb.Result, err error) {
	if !c.checkConnection() {
		return &pb.Result{}, errors.New("runAppStore: grpc connection failed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err = c.grpcClient.Store(ctx, &pb.StoreIntent{})
	for err != nil {
		logs.Debug("retry to runAppStore")
		result, err = c.grpcClient.Store(ctx, &pb.StoreIntent{})
	}
	logs.Infof("runAppStore: result:%v", result)
	return result, nil
}

// rpc远程调用服务端保存应用状态
func (c *RuntimeClient) RunAppRestore(keyStatus string) (result *pb.Result, err error) {
	if !c.checkConnection() {
		return &pb.Result{}, errors.New("runAppRestore: grpc connection failed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	data := &pb.Data{
		Name: "restoreData",
		Data: keyStatus,
	}
	result, err = c.grpcClient.Restore(ctx, &pb.RestoreIntent{Data: []*pb.Data{data}})
	for err != nil {
		logs.Debug("retry to runAppRestore")
		result, err = c.grpcClient.Restore(ctx, &pb.RestoreIntent{Data: []*pb.Data{data}})
	}
	logs.Infof("runAppRestore: result:%v", result)
	return result, nil
}

func (c *RuntimeClient) close() {
	defer c.conn.Close()
}
