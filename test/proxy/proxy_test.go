package proxy

import (
	"bufio"
	"context"
	"fmt"
	"github.com/google/uuid"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/scheme"
	"hit.edu/framework/pkg/client-go/rest"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestTimestamp(t *testing.T) {
	fmt.Println(time.Now().Format("2006-01-02-15-04-05"))
	u := uuid.New().String()
	fmt.Println(u)
}

func TestCreateTask(t *testing.T) {
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
			NegotiatedSerializer: serializer.NewCodecFactory(scheme.Scheme),
		},
		UserAgent: "defaultUserAgent",
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 10 * time.Second,
	}

	//创建ClientSet
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	groupName := "demo-groups"
	// 创建Group
	group1 := apis.Group{
		Spec: apis.GroupSpec{
			Name:    groupName,
			Actions: []apis.Action{},
		},
	}

	group2 := apis.Group{
		Spec: apis.GroupSpec{
			Name:    groupName,
			Actions: []apis.Action{},
		},
	}

	// 资源定义在 pkg/apis/xxx/type.go 下
	// 这里以访问资源Task为例，
	// 获取访问Task的客户端
	// 默认访问的Namespace是""
	tasksClient := clientSet.Core().Tasks(apis.NamespaceAll)
	taskName := "demo-tasks"
	task := &apis.Task{
		ObjectMeta: metav1.ObjectMeta{
			Name: taskName,
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "Task",
			APIVersion: "resources/v1",
		},
		Spec: apis.TaskSpec{
			Name: taskName,
			Groups: []apis.Group{
				group1,
				group2,
			},
		},
	}

	// Create一个Task
	fmt.Println("creating")
	results, err := tasksClient.Create(context.TODO(), task, metav1.CreateOptions{})

	if err != nil {
		panic(err)
	}
	fmt.Println("Created task ", results)
	prompt()

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
	fmt.Println()
}
