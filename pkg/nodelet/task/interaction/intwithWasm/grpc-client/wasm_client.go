package wasm_client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"hit.edu/framework/pkg/component-base/logs"
	wasm_interface "hit.edu/framework/pkg/nodelet/task/interaction/intwithWasm/proto"
)

type WasmClient struct {
	serverIPAndPort string
	app             string
	client          wasm_interface.WasmInterfaceClient
	conn            *grpc.ClientConn
	ctx             context.Context
}

func NewClient(ctx context.Context, port string, app string) *WasmClient {
	// app := "wasm-test-demo"
	return &WasmClient{serverIPAndPort: "127.0.0.1:" + port, app: app, ctx: ctx}
}

// // 与任务建立连接
// func (c *WasmClient) Connect() error {
// 	conn, err := grpc.Dial(c.serverIPAndPort, grpc.WithInsecure())
// 	if err != nil {
// 		logs.Error("wasm client:与任务建立连接失败")
// 		return err
// 	}
// 	grpcClient := wasm_interface.NewWasmInterfaceClient(conn)
// 	logs.Info("wasm client created")
// 	c.client = grpcClient
// 	// c.conn = conn
// 	return nil
// }

// 检查并重建与任务的rpc连接
func (c *WasmClient) checkConnection() bool {
	connState := true
	if c.conn == nil {
		// conn, err := grpc.NewClient(c.serverIPAndPort)
		conn, err := grpc.Dial(c.serverIPAndPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			// logs.Debugf("failed to connect to Wasm grpc server:%v", err)
			logs.Errorf("failed to connect to Wasm grpc server:%v", err)
			connState = false
		} else {
			c.conn = conn
			c.client = wasm_interface.NewWasmInterfaceClient(conn)
		}
	}
	return connState
}

// 部署-grpc接口
func (c *WasmClient) Deploy(wasm_file string) (*wasm_interface.Result, error) {
	if !c.checkConnection() {
		return &wasm_interface.Result{}, errors.New("Deploy: grpc connection failed")
	}
	// 目前的部署rpc接口是把wasm程序数据作为参数发送给运行时
	// wasm code
	// wasm_file_1 := "test/wasm/wasm_task/printf.wasm"
	wasm_file_1 := wasm_file
	f, err := os.ReadFile(wasm_file_1)
	if err != nil {
		logs.Errorf("read fail %v", err)
		return nil, err
	}
	// fmt.Println(f)
	data := wasm_interface.Data{Type: wasm_interface.DataType_DATA, Data: f}
	wasm_task := wasm_interface.WasmFile{Name: wasm_file_1, Data: &data}
	// runtime config
	// runtime_config := wasm_interface.RuntimeConfig{}
	// deploy
	deploy_data := wasm_interface.Deploy{App: c.app, RuntimeConfig: nil, Wasm: []*wasm_interface.WasmFile{&wasm_task}, Resource: []*wasm_interface.Resource{}}
	// deploy intent
	deploy_intent := &wasm_interface.DeployIntent{Identity: "identity", Data: []*wasm_interface.Deploy{&deploy_data}}

	// ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := c.client.Deploy(ctx, deploy_intent)
	for err != nil {
		logs.Errorf("%v : Deploy() failed, retry to rpc Deploy() , %v", c.app, err)
		result, err = c.client.Deploy(ctx, deploy_intent)
		time.Sleep(500 * time.Millisecond)
	}
	out := fmt.Sprintf("wasm deploy result: code: %d _ msg:  %s ", result.GetStateCode(), result.GetMsg())
	logs.Info(out)

	return result, err
}

// init-grpc接口
func (c *WasmClient) Init() (*wasm_interface.Result, error) {
	if !c.checkConnection() {
		return &wasm_interface.Result{}, errors.New("Init: grpc connection failed")
	}
	// init_intent_data
	init_intent := &wasm_interface.InitIntent{App: c.app}

	// ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := c.client.Init(ctx, init_intent)
	for err != nil {
		logs.Error("%v : Init() failed, retry to rpc Init() , %v", c.app, err)
		result, err = c.client.Init(ctx, init_intent)
		time.Sleep(500 * time.Millisecond)
	}
	out := fmt.Sprintf("wasm init result: code: %d _ msg:  %s ", result.GetStateCode(), result.GetMsg())
	logs.Info(out)
	return result, err
}

// start-grpc接口
func (c *WasmClient) Start() (*wasm_interface.Result, error) {
	if !c.checkConnection() {
		return &wasm_interface.Result{}, errors.New("Start: grpc connection failed")
	}

	// init_intent_data
	start_intent := &wasm_interface.StartIntent{App: c.app}

	// ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := c.client.Start(ctx, start_intent)
	for err != nil {
		logs.Error("%v : Start() failed, retry to rpc Start() , %v", c.app, err)
		result, err = c.client.Start(ctx, start_intent)
		time.Sleep(500 * time.Millisecond)
	}
	out := fmt.Sprintf("wasm start result: code: %d _ msg:  %s ", result.GetStateCode(), result.GetMsg())
	logs.Info(out)
	return result, err
}

// destroy-grpc接口
func (c *WasmClient) Destory() (*wasm_interface.Result, error) {
	if !c.checkConnection() {
		return &wasm_interface.Result{}, errors.New("Destory: grpc connection failed")
	}

	// init_intent_data
	destroy_intent := &wasm_interface.DestroyIntent{App: c.app}

	// ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result, err := c.client.Destroy(ctx, destroy_intent)
	for err != nil {
		logs.Error("%v : Destory() failed, retry to rpc Destory() , %v", c.app, err)
		result, err = c.client.Destroy(ctx, destroy_intent)
		time.Sleep(500 * time.Millisecond)
	}
	out := fmt.Sprintf("wasm destroy result: code: %d _ msg:  %s ", result.GetStateCode(), result.GetMsg())
	logs.Info(out)
	return result, err
}
