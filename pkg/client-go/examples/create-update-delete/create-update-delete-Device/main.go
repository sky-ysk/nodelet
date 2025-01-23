package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/types"
	"hit.edu/framework/pkg/apimachinery/watch"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"os"
	"time"
)

// 创建一个Rest Client
// 验证xxx动词
// 与API Server通信，并执行基础操作

func main() {
	logs.Init("main")
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	//参数配置
	// TODO: 填写参数
	//部分参数之后可以在core_client等 编写setConfigDefaults函数进行填充
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
			MaxIdleConns:        10000,            // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 1000 * time.Second,
	}

	//创建ClientSet
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	// 资源定义在 pkg/apis/xxx/type.go 下
	// 这里以访问资源Device为例，
	// 获取访问Device的客户端
	// 默认访问的Namespace是 ""

	devicesClient := clientSet.Core().Devices("test")

	device := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-devices",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "resources/v1",
		},
		Spec: apis.DeviceSpec{
			Name: "demo-device",
		},
	}
	device2 := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-device2",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "dev",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "resources/v1",
		},
		Spec: apis.DeviceSpec{
			Name: "demo-device",
		},
	}
	device3 := &apis.Device{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "demo-device3",
			Namespace: "test",
			Labels: map[string]string{
				"environment": "qa",
			},
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Device",
			APIVersion: "resources/v1",
		},
		Spec: apis.DeviceSpec{
			Name: "demo-device",
		},
	}

	patchDevice, err := json.Marshal(map[string]interface{}{
		"objectMeta": map[string]interface{}{
			"namespace": "test",
		},
		"spec": map[string]interface{}{
			"deviceName": "patch-device-name",
			"hostName":   "master",
		},
	})

	//监听事件并打印  监听resources/v1/devices
	go func() {
		logs.Trace("watching")
		var timeoutSeconds int64 = 20
		watchOptions := metav1.ListOptions{
			TimeoutSeconds: &timeoutSeconds,
		}

		watcher, err := devicesClient.Watch(context.TODO(), watchOptions)
		if err != nil {
			panic(err)
		}
		defer watcher.Stop() // 确保 watcher 被停止

		// 获取事件通道
		watchChan := watcher.ResultChan()

		for {
			select {
			case event, ok := <-watchChan:
				if !ok {
					logs.Tracef("watchChan closed")
					return
				}

				// 打印事件类型和对象的相关信息
				logs.Tracef("接收到事件类型:", event.Type)
				switch event.Type {
				case watch.Added:
					logs.Tracef("资源被添加: ", event.Object)
				case watch.Modified:
					logs.Tracef("资源被修改: ", event.Object)
				case watch.Deleted:
					logs.Tracef("资源被删除: ", event.Object)
				case watch.Error:
					logs.Tracef("发生错误: ", event.Object)
				case watch.Bookmark:
					logs.Tracef("收到Bookmark", event.Object)

				default:
					logs.Tracef("未识别的事件类型: ", event.Type)
				}
			}
		}
	}()

	// Create三个Device
	logs.Trace("creating")
	result, err := devicesClient.Create(context.TODO(), device, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create device: %v", err)
	} else {
		logs.Trace("created device", result)
	}
	_, err = devicesClient.Create(context.TODO(), device2, metav1.CreateOptions{})
	_, err = devicesClient.Create(context.TODO(), device3, metav1.CreateOptions{})
	prompt()

	//Update一个Device

	logs.Trace("updating")
	// 部分更改一个参数
	// 先Get一个Device ,更改Device的参数, UpdateDevice

	result, getErr := devicesClient.Get(context.TODO(), "demo-devices", metav1.GetOptions{})
	if getErr != nil {
		logs.Info(fmt.Errorf("Failed to get : %v", getErr))
	}

	logs.Tracef("get result", result)
	logs.Tracef("修改前的result.Spec.DeviceName：", result.Spec.Name)

	result.Spec.Name = "updatedDeviceName"
	_, updateErr := devicesClient.Update(context.TODO(), result, metav1.UpdateOptions{})
	if updateErr != nil {
		logs.Error(fmt.Errorf("Update failed: %v", updateErr))
	}

	logs.Tracef("修改后的result.Spec.DeviceName：", result.Spec.Name)
	logs.Tracef("Updated device...")
	prompt()

	// List 所有Device
	logs.Tracef("listing 筛选的device")
	lstOpts := metav1.ListOptions{
		LabelSelector: "environment",
	}
	list, err := devicesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}

	logs.Trace("listing done")
	prompt()

	//Patch 一个Device
	logs.Trace("patching")
	patchResult, err := devicesClient.Patch(context.TODO(), "demo-devices", types.StrategicMergePatchType, patchDevice, metav1.PatchOptions{})
	logs.Trace("patchResult: ", patchResult)
	logs.Trace("patch Done")

	// List 所有Device
	logs.Trace("listing")
	lstOpts = metav1.ListOptions{}
	list, err = devicesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}

	logs.Trace("listing done")
	prompt()

	// Delete一个Device
	logs.Trace("deleting")
	err = devicesClient.Delete(context.TODO(), "demo-devices", metav1.DeleteOptions{})
	if err != nil {
		panic(err)
	}
	logs.Trace("Deleted device...")
	prompt()

	// Delete 之后再次 List所有Device
	logs.Trace("listing")
	lstOpts = metav1.ListOptions{}
	list, err = devicesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}

	logs.Tracef("listing done")

	//DeleteCollection 删除所有Spec.DeviceName=demo-device的Device
	logs.Tracef("deleting collection")
	lstOpts = metav1.ListOptions{
		FieldSelector: "Spec.DeviceName=demo-device",
	}
	err = devicesClient.DeleteCollection(context.TODO(), metav1.DeleteOptions{}, lstOpts)
	if err != nil {
		logs.Error(err)
	}
	logs.Tracef("Deleted collection...")
	prompt()

	// DeleteCollection 之后再次 List所有Device
	logs.Trace("listing")
	lstOpts = metav1.ListOptions{}
	list, err = devicesClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err)
	}
	for _, d := range list.Items {
		logs.Trace(d)
	}
	logs.Trace("listing done")

	select {}
}

// From K8s
func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
