package plugins

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

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
	if node == nil || group == nil || node.Node() == nil {
		logs.Error("filter plugins", "node or group is nil")
		return framework.NewStatus(framework.Error, "node or group is nil")
	}
	if len(group.Spec.AffinityNodes) != 0 {
		for _, n := range group.Spec.AffinityNodes {
			if node.Node().Name == n {
				return framework.NewStatus(framework.Success, "")
			}
		}
		return framework.NewStatus(framework.Unschedulable, "node is not in affinity list")
	}
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
	logs.Infof("use default stategy to generate the score on node %s ", nodeName)
	//return sp.r.Int63n(10) + 1, status
	return int64(2), status
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
	taskClient  core.TaskInterface
}

func (bp *DefaultBindPlugin) Name() string {
	return "DefaultBindPlugin"
}

func (bp *DefaultBindPlugin) Bind(ctx context.Context, state *framework.CycleState, group *apis.Group, nodeName string) (status *framework.Status) {
	logs.Info("binding", group.ObjectMeta.Name, nodeName)
	//patch group
	patchGroup, err := json.Marshal(map[string]interface{}{
		"status": map[string]interface{}{
			"phase": apis.ReadyToDeploy,
			"node":  nodeName,
		},
	})
	if err != nil {
		logs.Error(err.Error())
		return framework.NewStatus(framework.Error, err.Error())
	}
	_, err = bp.groupClient.Patch(context.TODO(), group.ObjectMeta.Name, types.StrategicMergePatchType, patchGroup, metav1.PatchOptions{})
	if err != nil {
		logs.Error(err.Error())
	}
	//logs.Info(patchResult)

	//get belonged task
	//belongedTaskID := group.Status.Belongs.TaskID
	//task, err := bp.getTaskByID(ctx, belongedTaskID)
	//if err != nil {
	//	return framework.NewStatus(framework.Error, err.Error())
	//}
	//logs.Info("get belonged task ", task)
	//
	////patch task phase
	//patchTaskStatus, err := json.Marshal(map[string]interface{}{
	//	"status": map[string]interface{}{
	//		//TODO 后续修改状态
	//		"phase": apis.ReadyToDeploy,
	//	},
	//})
	//patchTaskResult, err := bp.taskClient.Patch(context.TODO(), task.ObjectMeta.Name, types.StrategicMergePatchType,
	//	patchTaskStatus, metav1.PatchOptions{})
	//if err != nil {
	//	logs.Error(err.Error())
	//	return framework.NewStatus(framework.Error, err.Error())
	//}
	//logs.Info(patchTaskResult)
	////patch task groups
	//for i := range task.Spec.Groups {
	//	if task.Spec.Groups[i].ObjectMeta.Name == group.ObjectMeta.Name {
	//		if task.Spec.Groups[i].Status.Phase == apis.Pending {
	//			task.Spec.Groups[i].Status.Phase = apis.ReadyToDeploy
	//			task.Spec.Groups[i].Status.Node = nodeName
	//		}
	//	}
	//}
	//groupsJson, err := json.Marshal(task.Spec.Groups)
	//if err != nil {
	//	logs.Error(err.Error())
	//	return framework.NewStatus(framework.Error, err.Error())
	//}
	//patchTaskGroups, err := json.Marshal(map[string]interface{}{
	//	"spec": map[string]interface{}{
	//		"groups": groupsJson,
	//	},
	//})
	//patchTaskGroupsResult, err := bp.taskClient.Patch(context.TODO(), task.ObjectMeta.Name, types.StrategicMergePatchType,
	//	patchTaskGroups, metav1.PatchOptions{})
	//if err != nil {
	//	logs.Error(err.Error())
	//	return framework.NewStatus(framework.Error, err.Error())
	//}
	//logs.Info(patchTaskGroupsResult)
	status = framework.NewStatus(framework.Success, "bind success")
	return status
}

func (bp *DefaultBindPlugin) getTaskByID(ctx context.Context, taskID string) (*apis.Task, error) {
	selector := fmt.Sprintf("Status.TaskID=%s", taskID)
	lstOpts := metav1.ListOptions{
		FieldSelector: selector,
	}
	list, err := bp.taskClient.List(context.TODO(), lstOpts)
	if err != nil {
		logs.Error(err.Error())
		return nil, err
	}
	if len(list.Items) != 1 {
		logs.Error("list the specific task fail", taskID)
		return nil, errors.New("list the specific task fail")
	}
	return &list.Items[0], nil
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
	groupsClient := clientSet.Core().Groups("test")
	tasksClient := clientSet.Core().Tasks("test")
	return &DefaultBindPlugin{
		groupClient: groupsClient,
		taskClient:  tasksClient,
	}, nil
}
