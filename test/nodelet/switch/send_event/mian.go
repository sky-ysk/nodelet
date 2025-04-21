package main

import (
	"bufio"
	"context"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/client-go/tools/recorder"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
	"net/http"
	"os"
	"time"
)

var scheme = runtime.NewScheme()

const NodeName = "CloudNode1"

var node = &apis.Node{
	ObjectMeta: meta.ObjectMeta{Name: NodeName, Namespace: "test"},
	TypeMeta:   meta.TypeMeta{Kind: "Node", APIVersion: "resources/v1"},
	Spec:       apis.NodeSpec{NodeName: NodeName},
}

func main() {
	logs.Init("testEventModule")

	clientSet := initClientSet(scheme)
	eventclient := clientSet.Core().Events("test")

	postEventForMigrate(eventclient)

	// 暂停让事件协程顺利执行
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
	recorder.Event(node, apis.EventTypeNormal, events.TriggerLocalMigration, fmt.Sprintf("Node Name:\t %s is shortage", node.Name))
	// recorder.Eventf(group, apis.EventTypeNormal, events.ReadyToMigrate, fmt.Sprintf("The task %v is ready for migration", group.Spec.Actions[0].Name))
}

func initClientSet(scheme *runtime.Scheme) *clients.ClientSet {
	apis.AddToScheme(scheme)
	logs.Info(scheme)
	// 创建ClientSet
	c := &rest.Config{
		Host:    "http://127.0.0.1:10000",
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
