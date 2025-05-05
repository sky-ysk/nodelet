package utils

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
)

func TestAddAction(t *testing.T) {
	logs.Init("main")
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	fmt.Println(scheme)
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
	actionsClient := clientSet.Core().Actions("test")
	action := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testmeta",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Action",
			APIVersion: "resources/v1",
		},
		Spec: apis.ActionSpec{
			Name: "testspec",
		},
	}
	fmt.Println("creating")
	_, err = actionsClient.Create(context.TODO(), action, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create action: %v", err)
		panic(err)
	}
	//git config  credential.helper store
	result, getErr := actionsClient.Get(context.TODO(), "testspec", metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get : %v", getErr))
	}
	fmt.Println("get result ", result)

}

func TestGetValue(t *testing.T) {
	logs.Init("main")
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	fmt.Println(scheme)
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
	actionsClient := clientSet.Core().Actions("test")
	action := &apis.Action{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testmeta",
			Namespace: "test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Action",
			APIVersion: "resources/v1",
		},
		Spec: apis.ActionSpec{
			Name: "testspec",
		},
	}
	fmt.Println("creating")
	_, err = actionsClient.Create(context.TODO(), action, metav1.CreateOptions{})
	if err != nil {
		logs.Errorf("Failed to create action: %v", err)
		panic(err)
	}
	//git config  credential.helper store
	result, getErr := actionsClient.Get(context.TODO(), "testspec", metav1.GetOptions{})
	if getErr != nil {
		panic(fmt.Errorf("Failed to get : %v", getErr))
	}
	fmt.Println("get result ", result)

}

// TODO 测试NodeDependency
func TestNodeDependency(t *testing.T) {

	//参数配置
}

// TODO 测试DataDependency
func TestDataDependency(t *testing.T) {
	//参数配置
}

// TODO 测试ResourceDependency
func TestResourceDependency(t *testing.T) {
	//参数配置
}

// TODO 测试ProcessDependency
func TestProcessDependency(t *testing.T) {
	//参数配置
}
