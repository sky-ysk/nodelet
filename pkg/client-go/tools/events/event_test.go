package events

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
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

func TestForEventClient(t *testing.T) {
	moduleName := "TestForEventClient"
	logs.Info("---", moduleName, "---")
	logs.Init(moduleName)
	scheme := runtime.NewScheme()
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

	// c := &rest.Config{}
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	// 获得eventsClient
	eventsClient := clientSet.Core().Events(apis.NamespaceAll)

	table := []apis.Event{
		{
			ObjectMeta: meta.ObjectMeta{
				Name:      "test-event",
				Namespace: "",
			},
			TypeMeta: meta.TypeMeta{
				Kind:       "Event",
				APIVersion: "resources/v1",
			},
			ObjectReference: apis.ObjectReference{
				Kind:       "Node",
				Name:       "CloudNode1",
				Namespace:  "",
				UID:        "default",
				APIVersion: "resources/v1",
			},
			Reason:  "Started",
			Message: "some verbose message: 1",
			Source:  apis.EventSource{Component: "eventTest", Host: "127.0.0.1"},
			Count:   1,
			Type:    apis.EventTypeNormal,
		},
	}

	// //如果已经存在，先删掉
	// err = eventsClient.Delete(context.TODO(), "test-event", meta.DeleteOptions{})
	// Create一个Task
	logs.Info("event creating")
	for _, item := range table {
		result, err := eventsClient.Create(context.TODO(), &item, meta.CreateOptions{})
		if err != nil {
			logs.Errorf("Failed to create event: %v", err)
			panic(err)
		}
		logs.Info("Created event : ", result)
	}

	prompt()
}

func TestForBroadcaster(t *testing.T) {
	moduleName := "TestForBroadcaster"
	logs.Info("---", moduleName, "---")
	logs.Init(moduleName)
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	logs.Info(scheme)

	// 初始化eventBroadcaster
	ctx := context.Background()
	eventBroadcaster := NewBroadcaster(WithSleepDuration(0), WithContext(ctx))
	defer eventBroadcaster.Shutdown()

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

	// c := &rest.Config{}
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	// 获得eventsClient
	eventsClient := clientSet.Core().Events(apis.NamespaceAll)

	// StartRecordingToSink()绑定了将事件上传至APIserver的handler
	// EventSinkImpl实现了上报事件的Create/patch/update方式，需要绑定一个eventsclient
	eventBroadcaster.StartRecordingToSink(ctx, &core.EventSinkImpl{Interface: eventsClient})

	table := []apis.Event{
		{
			ObjectMeta: meta.ObjectMeta{
				Name:      "test-event",
				Namespace: "default",
			},
			ObjectReference: apis.ObjectReference{
				Kind:       "Pod",
				Name:       "test-pod",
				Namespace:  "default",
				UID:        "bar",
				APIVersion: "v1",
			},
			Reason:  "Started",
			Message: "some verbose message: 1",
			Source:  apis.EventSource{Component: "eventTest", Host: "127.0.0.1"},
			Count:   1,
			Type:    apis.EventTypeNormal,
		},
	}
	// 生成并提交一个event
	recorder := eventBroadcaster.NewRecorder(schema.NewSchema(), apis.EventSource{Component: "eventTest"})
	for _, item := range table {
		recorder.Eventf(item.DeepCopyObject(), item.Type, item.Reason, item.Message)
	}

	time.Sleep(3 * time.Second)
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
