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
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/events"
)

var scheme = runtime.NewScheme()

func main() {
	logs.Init("testEventModule")
	clientSet := initClientSet(scheme)
	client := clientSet.Core().Events("test")

	// 监听event
	go eventListener(context.Background(), client)

	// 获知触发迁移的事件源进行迁移处理
	prompt()

}

func eventListener(ctx context.Context, client core.EventInterface) {
	logs.Infof("event watching")

	// 筛选出 事件触发原因是 ReadyToMigrate 的事件
	fieldSelector := fmt.Sprintf("reason=%v", events.TriggerMigration)
	// 设置长超时时间
	var timeout int64 = 3600
	watchOptions := meta.ListOptions{
		TimeoutSeconds: &timeout,
		FieldSelector:  fieldSelector,
	}

	watcher, err := client.Watch(ctx, watchOptions)
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
				// logs.Infof("资源被添加: ", event.Object)
				event, ok := event.Object.(*apis.Event)
				if !ok {
					return
				}
				logs.Infof("*****************now time:%v,event time:%v", time.Now(), event.EventTime.Time)
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
		Timeout: 3600 * time.Second,
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
