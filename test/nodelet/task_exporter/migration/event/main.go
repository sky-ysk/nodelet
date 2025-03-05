package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/watch"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/client-go/tools/recorder"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
)

var scheme = runtime.NewScheme()

// 模拟一个资源对象（如Pod、Task）的引用，因为事件通常需要与具体的资源相关联
var group = &apis.Group{
	ObjectMeta: meta.ObjectMeta{Name: "TestGroup-cmd", Namespace: "test"},
	TypeMeta:   meta.TypeMeta{Kind: "Group", APIVersion: "resources/v1"},
	Spec: apis.GroupSpec{
		Name:    "TestGroup-cmd",
		Parents: make([]string, 0),
		Actions: []apis.Action{
			{
				ObjectMeta: meta.ObjectMeta{Name: "cmd_action"},
				Spec: apis.ActionSpec{
					Name: "cmd_action",
					Runtimes: []apis.Runtime{
						{
							Name:                         "yolo-cmd",
							Type:                         apis.ByCommand,
							Command:                      []string{"python3"},
							Args:                         []string{"/home/kcm/workplace/migration-demo-0116/yolo-runner.py"},
							EnableFineGrainedControl:     true,
							EnableFineGrainedControlPort: "5123",
						},
					},
				},
			},
		},
	},
}

func main() {
	logs.Init("testEventModule")
	clientSet := initClientSet(scheme)
	client := clientSet.Core().Events("test")
	// 模拟提交一个事件，事件Reason为 events.ReadyToMigrate
	postEventForMigrate(client)
	// 监听event
	go eventListener(client)
	// 获知触发迁移的事件源进行迁移处理
	prompt()

}

func postEventForMigrate(client core.EventInterface) {
	// 这些配置实际在组件初始化时就已经完成
	ctx := context.Background()
	eventBroadcaster := recorder.NewBroadcaster(recorder.WithContext(ctx))
	defer eventBroadcaster.Shutdown()
	eventBroadcaster.StartRecordingToSink(ctx, &core.EventSinkImpl{Interface: client})
	recorder := eventBroadcaster.NewRecorder(scheme, "test-controller")

	// 通过 recorder.Event或 recorder.Eventf可以生成事件
	recorder.Eventf(group, apis.EventTypeNormal, events.ReadyToMigrate, fmt.Sprintf("The task %v is ready for migration", group.Spec.Actions[0].Name))
}

func eventListener(client core.EventInterface) {
	logs.Infof("event watching")

	// 筛选出 事件触发原因是 ReadyToMigrate 的事件
	fieldSelector := fmt.Sprintf("reason=%v", events.ReadyToMigrate)
	watchOptions := meta.ListOptions{
		FieldSelector: fieldSelector,
	}

	watcher, err := client.Watch(context.TODO(), watchOptions)
	if err != nil {
		panic(err)
	}
	defer watcher.Stop() // 确保 watcher 被停止
	watchChan := watcher.ResultChan()

	for {
		select {
		case event, ok := <-watchChan:
			if !ok {
				logs.Infof("watchChan closed")
				return
			}
			// 打印事件类型和对象的相关信息
			logs.Infof("接收到事件类型: %v\n", event.Type)
			switch event.Type {
			case watch.Added:
				logs.Infof("资源被添加: ", event.Object)
			case watch.Modified:
				logs.Infof("资源被修改: ", event.Object)
			case watch.Deleted:
				logs.Infof("资源被删除: ", event.Object)
			case watch.Error:
				logs.Infof("发生错误: ", event.Object)
			default:
				logs.Infof("未识别的事件类型: ", event.Type)
			}
		}
	}
}

func initClientSet(scheme *runtime.Scheme) *clients.ClientSet {
	apis.AddToScheme(scheme)
	logs.Info(scheme)
	// 创建ClientSet
	c := &rest.Config{
		Host:    "http://localhost:10000",
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8", //text/plain; charset=UTF-8
			ContentType:        "application/json; charset=UTF-8", //application/json; charset=UTF-8
			GroupVersion: &schema.GroupVersion{
				Group:   "resources",
				Version: "v1",
			},
			NegotiatedSerializer: serializer.NewCodecFactory(scheme),
		},
		UserAgent: "defaultUserAgent",
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 10 * time.Second,
	}
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	return clientSet
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	logs.Info()
}
