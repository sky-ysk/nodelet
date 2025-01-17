package node

import (
	"context"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/node/collector"
	"strings"
	"sync"
	"time"
)

//TODO 11.14 node-exporter后续需要实现的功能，数据处理并填入nodestatus字段，通过client-go定期写入api-server中 --完成

// TODO: 接口格式调整
type Exporter interface {
	// TODO: 定义接口
	Run(ctx context.Context) error
}

var _ Exporter = &NodeExporter{}

type NodeExporter struct {
	// Register
	// TODO: 扫描节点信息，并注册到API-Server中
	// TODO: 定义组件
	nodeCollector    *collector.NodeCollector
	staticCache      map[string]collector.Metric
	dynamicCache     map[string]collector.Metric
	staticCacheLock  sync.RWMutex
	dynamicCacheLock sync.RWMutex
	nodesClient      core.NodeInterface
	apiServerURL     string //上传目标的API-server的地址
}

func NewNodeExporter(cfg *Config, client core.NodeInterface) (*NodeExporter, error) {
	// TODO：参数配置
	// 创建NodeCollector 读取配置信息
	logs.Info("Init NodeExporter module")
	nc, err := collector.NewNodeCollector(cfg.EnabledCollectors)
	if err != nil {
		return nil, err
	}
	return &NodeExporter{
		nodeCollector: nc,
		staticCache:   make(map[string]collector.Metric),
		dynamicCache:  make(map[string]collector.Metric),
		apiServerURL:  cfg.ResourceAccessMethod,
		nodesClient:   client,
	}, nil
}

// 想改成每隔60秒收集一次静态信息，每隔1s收集一次动态信息
func (n *NodeExporter) Run(ctx context.Context) error {
	// 刚开始启动先收集一次，之后定期Gather一次数据
	err := n.nodeCollector.GatherStaticData(n.processMetri) //注意方法参数后面没有()
	if err != nil {
		return err
	}
	err = n.nodeCollector.GatherDynamicData(n.processMetri)
	if err != nil {
		return err
	}

	staticTicker := time.NewTicker(time.Second * 60)
	defer staticTicker.Stop()
	dynamicTicker := time.NewTicker(time.Second * 5)
	defer dynamicTicker.Stop()
	//TODO
	// 将收集的数据更新到API-Server中
	// 更新Node
	for {
		select {
		case <-staticTicker.C:
			err := n.nodeCollector.GatherStaticData(n.processMetri)
			if err != nil {
				return err
			}
			//这里需要先从api-server当中获取Node结构体指针
			node, getErr := n.nodesClient.Get(context.TODO(), "demo-nodes", metav1.GetOptions{})
			if getErr != nil {
				panic(fmt.Errorf("Failed to get : %v", getErr))
			}
			// 然后将收集到的数据填充到NodeStatus当中，最后返回Node结构体指针给api-server
			n.UploadCache(node, "static")
		case <-dynamicTicker.C:
			err := n.nodeCollector.GatherDynamicData(n.processMetri)
			if err != nil {
				return err
			}
			//这里需要先从api-server当中获取Node结构体指针
			node := &apis.Node{Spec: apis.NodeSpec{NodeName: "1"}}
			// 然后将收集到的数据填充到NodeStatus当中，最后返回Node结构体指针给api-server
			n.UploadCache(node, "dynamic")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
func (n *NodeExporter) UploadCache(node *apis.Node, cacheType string) {
	// TODO: 数据预处理，然后放入Node中，然后写入API-Server中----应该是将信息放入到NodeStatus当中，然后通过client-go，定期写入到api-server当中
	var cache map[string]collector.Metric
	var lock *sync.RWMutex

	if cacheType == "static" {
		cache = n.staticCache
		lock = &n.staticCacheLock
	} else {
		cache = n.dynamicCache
		lock = &n.dynamicCacheLock
	}
	lock.Lock()
	defer lock.Unlock()

	if len(cache) == 0 {
		logs.Info("Cache is empty, nothing to upload")
		return
	}
	// 读取内存cache当中的metric，转换为NodeStatus当中的Resource、Usage
	for _, metric := range cache {
		n.processMetriToNode(node, metric, cacheType)
	}

	//TODO 实现上传逻辑到API-server{  ----先放入到NodeStatus当中，然后通过client-go写入到api-server当中
	_, err := n.nodesClient.Update(context.TODO(), node, metav1.UpdateOptions{})
	if err != nil {
		logs.Errorf("Failed to update Node, err:%v", err)
		return
	}
	// 清空缓存
	// 清空缓存，直接操作结构体中的缓存
	if cacheType == "static" {
		n.staticCache = make(map[string]collector.Metric)
	} else {
		n.dynamicCache = make(map[string]collector.Metric)
	}
	logs.Info("Cache uploaded and cleared")
}

func (n *NodeExporter) processMetri(types string, metric collector.Metric) {
	// TODO：处理Metric
	// TODO: 存入本地Cache或同步到manager中？待定
	// TODO: 将Cache放到nodelet.go中   --应该是放到Nodeexporter当中
	//暂时先实现写入本地Cache
	key := fmt.Sprintf("%s-%v", metric.Item[0].GetName(), metric.Item[0].GetLabels())
	if types == "static" {
		n.staticCacheLock.Lock()
		defer n.staticCacheLock.Unlock()
		n.staticCache[key] = metric // 将 metric 存储到缓存中
	} else {
		n.dynamicCacheLock.Lock()
		defer n.dynamicCacheLock.Unlock()
		n.dynamicCache[key] = metric // 将 metric 存储到缓存中
	}
	logs.Infof("Processed metric:%v-%v successfully", key, metric.ToString())
}

func (n *NodeExporter) processMetriToNode(node *apis.Node, metric collector.Metric, cacheType string) {
	// 这里有个细节，就是metric里面装的是Item，Item可能是静态数据也可能是动态数据，那么我怎么知道是静态数据还是动态数据
	var itemList []apis.Item
	for _, item := range metric.Item {
		apisItem := convertToApisItem(*item)
		itemList = append(itemList, apisItem)
	}
	part := strings.Split(metric.Item[0].GetName(), ".")[1]

	if cacheType == "static" {
		n.addStaticDataToNodeSpec(node, part, itemList)
	} else {
		n.addDynamicDataToNodeStatus(node, part, itemList)
	}
}

func (n *NodeExporter) addStaticDataToNodeSpec(node *apis.Node, part string, itemList []apis.Item) {
	switch part {
	case collector.CpuCollectorName:
		node.Spec.Resource["cpu"] = itemList
	case collector.MemoryCollectorName:
		node.Spec.Resource["memory"] = itemList
	case collector.StorageCollectorName:
		node.Spec.Resource["storage"] = itemList
	}
}

func (n *NodeExporter) addDynamicDataToNodeStatus(node *apis.Node, part string, itemList []apis.Item) {
	switch part {
	case collector.CpuCollectorName:
		node.Status.Usage["cpu"] = itemList
	case collector.MemoryCollectorName:
		node.Status.Usage["memory"] = itemList
	case collector.StorageCollectorName:
		node.Status.Usage["storage"] = itemList
	}
}
func convertToApisItem(item collector.Item) apis.Item {
	return apis.Item{
		Name:   item.GetName(),
		Desc:   item.GetDesc(),
		Labels: item.GetLabels(),
		Values: item.GetValues(),
	}
}
