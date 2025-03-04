package recorder

import (
	"context"
	"net/http"
	"testing"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
)

var table = []apis.Event{
	{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test-event",
			Namespace: "test",
		},
		TypeMeta: meta.TypeMeta{
			Kind:       "Event",
			APIVersion: "resources/v1",
		},
		InvolvedObject: apis.ObjectReference{
			Kind:       "Group",
			Name:       "mygroup",
			Namespace:  "test",
			UID:        "default",
			APIVersion: "resources/v1",
			FieldPath:  "spec.containers{mycon}",
		},
		Reason:  "Started",
		Message: "Successfully Started ComandRuntime",
		Source:  apis.EventSource{Component: "test-controller", Host: "127.0.0.1"},
		Count:   1,
		Type:    apis.EventTypeNormal,
	},
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

func TestForBroadcaster(t *testing.T) {
	moduleName := "TestForBroadcaster"
	logs.Init(moduleName)
	ctx := context.Background()
	scheme := runtime.NewScheme()

	// 1. 创建eventsClient
	eventsClient := initClientSet(scheme).Core().Events("test")

	// 2. 创建eventBroadcaster
	eventBroadcaster := NewBroadcaster(WithContext(ctx))
	defer eventBroadcaster.Shutdown()

	// 3.1 启动事件的 API Server 记录功能, StartRecordingToSink()定义了将事件上传至api server的事件处理方式
	// 配置事件接收器，需要绑定一个eventsClient
	eventBroadcaster.StartRecordingToSink(ctx, &core.EventSinkImpl{Interface: eventsClient})

	// 3.2 启动日志记录功能
	eventBroadcaster.StartLogging(ctx, logs.Infof)

	// 4. 创建事件记录器EventRecorder, 用于记录事件
	recorder := eventBroadcaster.NewRecorder(scheme, "test-controller")

	// 5. 模拟一个资源对象（如Pod、Task）的引用，因为事件通常需要与具体的资源相关联
	var testGroup = &apis.Group{
		ObjectMeta: meta.ObjectMeta{
			Name:      "group-test",
			Namespace: "test",
			UID:       "123",
		},
	}

	for _, item := range table {
		recorder.Eventf(testGroup, item.Type, item.Reason, item.Message)
	}

	time.Sleep(2 * time.Second)
}

func TestForEventClient(t *testing.T) {
	moduleName := "TestForEventClient"
	logs.Init(moduleName)
	scheme := runtime.NewScheme()

	// 获得eventsClient
	eventsClient := initClientSet(scheme).Core().Events("test")

	// 先删除冗余事件
	clearEvents(eventsClient)

	// Create一个event
	logs.Info("event creating")
	for _, item := range table {
		result, err := eventsClient.Create(context.TODO(), &item, meta.CreateOptions{})
		if err != nil {
			logs.Errorf("Failed to create event: %v", err)
			panic(err)
		}
		logs.Info("Created event : ", result.Name)
	}
}

func TestForRef(t *testing.T) {

}

func clearEvents(client core.EventInterface) {
	logs.Info("event deleting")
	for _, item := range table {
		err := client.Delete(context.TODO(), item.Name, meta.DeleteOptions{})
		if err != nil {
			logs.Errorf("Failed to delete event: %v", err)
			// panic(err)
		}
	}

}

// objRef := &apis.ObjectReference{
// 	Kind:       "Pod",
// 	Name:       "mypod",
// 	Namespace:  "test",
// 	UID:        "default",
// 	APIVersion: "resources/v1",
// 	FieldPath:  "spec.containers{mycon}",
// }
// testRef, err := reference.GetPartialReference(scheme, testGroup, "spec.actions[2]")
