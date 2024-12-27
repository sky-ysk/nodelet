package node

import (
	"context"
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/node/collector"
	"sync"
	"time"
)

//TODO 11.14 node-exporter后续需要实现的功能，数据处理并填入nodestatus字段，通过client-go定期写入api-server中

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
	nodeCollector *collector.NodeCollector
	cache         map[string]collector.Metric
	cacheLock     sync.RWMutex
	apiServerURL  string //上传目标的API-server的地址
}

func NewNodeExporter(cfg *Config) (*NodeExporter, error) {
	// TODO：参数配置
	// 创建NodeCollector 读取配置信息
	logs.Info("init NodeExporter-------")
	nc, err := collector.NewNodeCollector(cfg.EnabledCollectors)
	if err != nil {
		return nil, err
	}
	ne := &NodeExporter{
		nodeCollector: nc,
		cache:         make(map[string]collector.Metric),
		apiServerURL:  cfg.ResourceAccessMethod,
	}
	return ne, nil
}

// 想改成每隔60秒收集一次静态信息，每隔1s收集一次动态信息
func (n *NodeExporter) Run(ctx context.Context) error {
	// TODO: 开始运行

	// 刚开始启动先收集一次，之后定期Gather一次数据
	err := n.nodeCollector.GatherStaticData(n.processMetri) //注意方法参数后面没有()
	if err != nil {
		return err
	}

	err = n.nodeCollector.GatherDynamicData(n.processMetri)
	if err != nil {
		return err
	}

	staticTicker := time.NewTicker(time.Second * 30)
	defer staticTicker.Stop()
	dynamicTicker := time.NewTicker(time.Second * 5)
	defer dynamicTicker.Stop()
	upLoadTicker := time.NewTicker(time.Second * 30)
	defer upLoadTicker.Stop()

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
		case <-dynamicTicker.C:
			err := n.nodeCollector.GatherDynamicData(n.processMetri)
			if err != nil {
				return err
			}
		case <-upLoadTicker.C:
			n.UploadCache()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (n *NodeExporter) UploadCache() {
	// TODO: 数据预处理，然后放入Node中，然后写入API-Server中----应该是将信息放入到NodeStatus当中，然后通过client-go，定期写入到api-server当中
	n.cacheLock.Lock()
	defer n.cacheLock.Unlock()

	if len(n.cache) == 0 {
		logs.Info("cache is empty, nothing to upload")
		return
	}
	//TODO 实现上传逻辑到API-server{  ----先放入到NodeStatus当中，然后通过client-go写入到api-server当中
	for key, metric := range n.cache {
		logs.Infof("uploading metric to API-Server: %s -> %v", key, metric.ToString())
	}
	// 清空缓存
	n.cache = make(map[string]collector.Metric)
	logs.Info("cache uploaded and cleared")
}

func (n *NodeExporter) processMetri(metric collector.Metric) {
	// TODO：处理Metric
	// TODO: 存入本地Cache或同步到manager中？待定

	//暂时先实现写入本地Cache
	key := fmt.Sprintf("%s-%v", metric.Item.GetName(), metric.Item.GetLbels())
	n.cacheLock.Lock()
	defer n.cacheLock.Unlock()
	n.cache[key] = metric // 将 metric 存储到缓存中
	logs.Info("successfully processed metric: ", key, metric.ToString())
	// TODO: 将Cache放到nodelet.go中   --应该是放到Nodeexporter当中
}
