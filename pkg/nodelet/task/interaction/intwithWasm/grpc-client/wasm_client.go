package wasm_client

import (
	"context"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
	"hit.edu/framework/pkg/component-base/logs"
	wasm_interface "hit.edu/framework/pkg/nodelet/task/interaction/intwithWasm/proto"
)

type WasmClient struct {
	serverIPAndPort string
	app             string
	client          wasm_interface.WasmInterfaceClient
	// conn            *grpc.ClientConn
	ctx context.Context
}

//TaskExporter直接与任务进行通信的client相关的函数接口，被assigner里的函数调用

func NewClient(ctx context.Context, serverIPAndPort string) *WasmClient {
	app := "wasm-test-demo"
	return &WasmClient{serverIPAndPort: serverIPAndPort, app: app, ctx: ctx}
}

// 与任务建立连接
func (c *WasmClient) Connect() error {
	conn, err := grpc.Dial(c.serverIPAndPort, grpc.WithInsecure())
	if err != nil {
		logs.V2().Error("wasm client:与任务建立连接失败")
		return err
	}
	grpcClient := wasm_interface.NewWasmInterfaceClient(conn)
	logs.V1().Info("wasm client created")
	c.client = grpcClient
	// c.conn = conn
	return nil
}

// 部署-grpc接口
func (c *WasmClient) Deploy(wasm_file string) (*wasm_interface.Result, error) {
	// 目前的部署rpc接口是把wasm程序数据作为参数发送给运行时
	// wasm code
	// wasm_file_1 := "test/wasm/wasm_task/printf.wasm"
	wasm_file_1 := wasm_file
	f, err := os.ReadFile(wasm_file_1)
	if err != nil {
		logs.V2().Error("read fail", err)
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	result, err := c.client.Deploy(ctx, deploy_intent)
	if err != nil {
		logs.V2().Error(c.app+" : deploy fail,", err)
	}
	out := fmt.Sprintf("deploy result: code: %d _ msg:  %s ", result.GetStateCode(), result.GetMsg())
	logs.V1().Info(out)
	return result, err
}

// init-grpc接口
func (c *WasmClient) Init() (*wasm_interface.Result, error) {
	// init_intent_data
	init_intent := &wasm_interface.InitIntent{App: c.app}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := c.client.Init(ctx, init_intent)
	if err != nil {
		logs.V2().Error(c.app+" : init fail", err)
	}
	out := fmt.Sprintf("init result: code: %d _ msg:  %s ", result.GetStateCode(), result.GetMsg())
	logs.V1().Info(out)
	return result, err
}

// start-grpc接口
func (c *WasmClient) Start() (*wasm_interface.Result, error) {
	// init_intent_data
	start_intent := &wasm_interface.StartIntent{App: c.app}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := c.client.Start(ctx, start_intent)
	if err != nil {
		logs.V2().Error(c.app+" : start fail", err)
	}
	out := fmt.Sprintf("start result: code: %d _ msg:  %s ", result.GetStateCode(), result.GetMsg())
	logs.V1().Info(out)
	return result, err
}

// destroy-grpc接口
func (c *WasmClient) Destory() (*wasm_interface.Result, error) {
	// init_intent_data
	destroy_intent := &wasm_interface.DestroyIntent{App: c.app}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := c.client.Destroy(ctx, destroy_intent)
	if err != nil {
		logs.V2().Error(c.app+" : destroy fail", err)
	}
	out := fmt.Sprintf("destroy result: code: %d _ msg:  %s ", result.GetStateCode(), result.GetMsg())
	logs.V1().Info(out)
	return result, err
}
