package plugins

import (
	"context"
	"encoding/json"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	"hit.edu/framework/pkg/apimachinery/types"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/apis/config"
	"hit.edu/framework/pkg/scheduler/framework"
	"math/rand"
	"net/http"
	"time"
)

type DefaultFilter struct {
}

func (fl *DefaultFilter) Name() string {
	return "DefaultFilter"
}

// 插件接口 字段
// TODO @linbohai GO 单元测试
// 过滤插件 返回Group是否能在对应的设备节点上运行的信息
// 可以运行返回success
func (fl *DefaultFilter) Filter(ctx context.Context, group *apis.Group, node *config.NodeInfo) *framework.Status {
	status := framework.NewStatus(framework.Success, "default success")
	return status
}

type DefaultScorePlugin struct {
	r *rand.Rand
}

func (sp *DefaultScorePlugin) Name() string {
	return "DefaultScorePlugin"
}

// 默认打分 随机生成一个0-10的整数
func (sp *DefaultScorePlugin) Score(ctx context.Context, group *apis.Group, nodeName string) (int64, *framework.Status) {
	status := framework.NewStatus(framework.Success, "default success")
	return int64(sp.r.Intn(11)), status
}

func NewDefaultFilterPlugin(ctx context.Context, f framework.Handle) (framework.Plugin, error) {
	return &DefaultFilter{}, nil
}

func NewDefaultScorePlugin(ctx context.Context, f framework.Handle) (framework.Plugin, error) {
	return &DefaultScorePlugin{
		r: rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

type DefaultBindPlugin struct {
	//TODO 需要资源层对象
	groupClient core.GroupInterface
}

func (bp *DefaultBindPlugin) Name() string {
	return "DefaultBindPlugin"
}

func (bp *DefaultBindPlugin) Bind(ctx context.Context, state *framework.CycleState, group *apis.Group, nodeName string) (status *framework.Status) {
	//group.Spec.Phase = binding..

	//from hezhangyi
	//bindMethod(nodeName, group)
	//check feedback
	logs.Info("binding", group.ObjectMeta.Name, nodeName)
	patchGroup, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"phase": apis.ReadyToDeploy,
			"node":  nodeName,
		},
	})
	if err != nil {
		logs.Error(err.Error())
	}
	patchResult, err := bp.groupClient.Patch(context.TODO(), group.ObjectMeta.Name, types.StrategicMergePatchType, patchGroup, metav1.PatchOptions{})
	if err != nil {
		logs.Error(err.Error())
	}
	logs.Info(patchResult)
	status = framework.NewStatus(framework.Success, "bind success")
	return status
}

func NewDefaultBindPlugin(ctx context.Context, f framework.Handle) (framework.Plugin, error) {
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	//fmt.Println(scheme)
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
		Timeout: 3600 * time.Second,
	}

	//创建ClientSet
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		panic(err)
	}
	groupsClient := clientSet.Core().Groups("")
	return &DefaultBindPlugin{
		groupClient: groupsClient,
	}, nil
}
