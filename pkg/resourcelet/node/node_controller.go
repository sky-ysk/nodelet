package node

import (
	"context"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	coresinformer "hit.edu/framework/pkg/client-go/informers/cores"
	coreslister "hit.edu/framework/pkg/client-go/listers/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"k8s.io/apimachinery/pkg/util/wait"
	"time"
)

// 每个Node会向API-Server注册自己的Node信息
// 并且定期收集Node信息，上报到API-Server

// Node Controller监控所有的Node的状态及其资源信息
// 对于出现异常情况的节点，尝试恢复节点 TODO：这个部分后面可能放在事件处理
//     错误恢复的措施包括
//          尝试重新连接节点，......
// 对于无法恢复的节点，则将上述节点上的Group标记为Pending状态，由调度器重新调度部署，并将该节点删除
// 对于无法恢复的节点，与节点关联的设备需要迁移到其他节点上（未与节点直接建立连接的设备，如机器人等), 无法迁移的任务需要删除

// 资源的扩缩容
// 扩容：所有Node的节点内的资源应该保持在一定的阈值内，如果可用资源低于一定的阈值时，创建新的VM，以扩展节点的数量
// TODO：+Optional 缩容：当可用资源高于一定的资源，如果有节省资源的需求，则可以关闭一些VM

type NodeController struct {
	//TODO: Node Client
	client clients.Interface

	//TODO: Node Lister
	nodeLister coreslister.NodeLister

	//TODO: NodeInformer
	nodeInformer coresinformer.NodeInformer

	// 存储本地所有已知的Node
	knownNodes map[string]*apis.Node

	// Node监控的周期
	nodeMonitorPeriod time.Duration

	//TODO: Event Recorder

	//TODO: Queue
	// TODO: workqueue

	//TODO: Node健康检查?
}

func NewNodeController(ctx context.Context) (*NodeController, error) {
	// TODO: 参数填充
	// TODO: 创建Event Broadcaster

	// TODO: 创建Informer,以及对应的EventHandler
	// TODO: NodeInformer

	nc := &NodeController{}
	return nc, nil
}

func (nc *NodeController) Run(ctx context.Context) {
	// TODO: 添加各类Worker

	// 定期监控Node的Health情况（通过List Node），监控阈值
	// 如果资源过低，进行资源的扩容
	// 剩余资源较多，按需进行缩容
	go wait.UntilWithContext(ctx,
		func(ctx context.Context) {
			if err := nc.monitorNodes(ctx); err != nil {
				logs.Error("Failed to monitor nodes ", err)
			}
		}, nc.nodeMonitorPeriod)
}

func (nc *NodeController) monitorNodes(ctx context.Context) error {
	start := time.Now()

	// 使用Lister获取所有的Node信息
	// Node信息从本地Cache中获取，会有一定的延迟
	// 如果使用Informer,那么有改动则会通知
	nodes, err := nc.nodeLister.List() // FIXME：目前默认返回所有的Node, 后面需要添加Labels和Namespace
	if err != nil {
		return err
	}

	// 将Nodes进行分类，区分出来是Added, Updated还是Deleted
	// 如果是Added
	// 将Node添加到本地已知队列里面

	// 如果是Deleted
	// 将Node从本地已知队列里面删除

	// TODO: 如果是Updated, 操作待定
	// TODO: 对于产生异常事件的Node

	// 如果是其他，长时间没有更新的资源，则尝试恢复

	// TODO: 设置一个参数，来确定是否开始资源扩缩容的功能，只有云边侧考虑使用该功能
	// TODO: 计算资源整体的阈值
	// TODO: 资源缩扩容
	// TODO: 一次可能扩容多个Node
	// TODO: 扩容完成的Node,需要能够自动添加到集群中（将Nodelet打包到镜像中，打成系统服务，自动注册，添加到集群中)

	return nil
}
