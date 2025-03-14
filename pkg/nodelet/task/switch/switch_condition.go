package _switch

//import (
//	"context"
//	apis "hit.edu/framework/pkg/apis/cores"
//	metav1 "hit.edu/framework/pkg/apis/meta"
//	"hit.edu/framework/pkg/client-go/clients/typed/core"
//	"hit.edu/framework/pkg/component-base/logs"
//	"hit.edu/framework/pkg/nodelet/node"
//	"strconv"
//	"time"
//)
//
//const (
//	thresholdbattery float64 = 5
//	thresholdCPU     float64 = 70
//	thresholdGPU     float64 = 90
//	thresholdMemory  float64 = 80
//	thresholdNetWork float64 = 50
//	thresholdStorage float64 = 60
//)
//
//type SwitchCheck struct {
//	ClusterCategory string
//	//Client-go
//	nodesClient     core.NodeInterface //需要查node信息
//	SwitchCondition switchCondition
//}
//type switchCondition interface {
//	CheckCondition(node *apis.Node) bool
//}
//
//func NewSwitchCheck(nodeClient core.NodeInterface) *SwitchCheck {
//	var switchCondition switchCondition
//	node, err := nodeClient.Get(context.TODO(), node.NodeName, metav1.GetOptions{})
//	if err != nil {
//		logs.Errorf("Get node:%s from etcd err: %v", node.Name, err)
//	}
//	logs.Infof("node.Spec.ClusterCategory:%v", node.Spec.ClusterCategory)
//	clusterCategoty := node.Spec.ClusterCategory
//	switch clusterCategoty {
//	case "Cloud":
//		switchCondition = &CloudNodeSwitchCondition{} //因为此处为(c *CloudNodeSwitchCondition) CheckCondition
//	case "Edge":
//		switchCondition = &EdgeNodeSwitchCondition{}
//	case "End":
//		switchCondition = &EndNodeSwitchCondition{}
//	default:
//		logs.Errorf("Unknown cluster categoty:%s", clusterCategoty)
//		return nil
//	}
//	return &SwitchCheck{
//		ClusterCategory: clusterCategoty,
//		nodesClient:     nodeClient,
//		SwitchCondition: switchCondition,
//	}
//}
//
//type CloudNodeSwitchCondition struct {
//}
//
//func (c *CloudNodeSwitchCondition) CheckCondition(node *apis.Node) bool {
//	cpuAveUtil := getFloatValue(node.Status.Usage["cpu"][0].Values["AveUtil"])
//	memoryUsage := getFloatValue(node.Status.Usage["memory"][0].Values["Usage"])
//	storageUsage := getFloatValue(node.Status.Usage["storage"][0].Values["Usage"])
//	logs.Infof("检查任务状态----CPU利用率：%v,内存利用率：%v，存储利用率：%v", cpuAveUtil, memoryUsage, storageUsage)
//
//	// 下面两行代码为了测试迁移流程
//	time.Sleep(10 * time.Second)
//	return true
//	//return cpuAveUtil > thresholdCPU || memoryUsage > thresholdMemory || storageUsage > thresholdStorage
//}
//
//type EdgeNodeSwitchCondition struct {
//}
//
//func (e *EdgeNodeSwitchCondition) CheckCondition(node *apis.Node) bool {
//	cpuAveUtil := getFloatValue(node.Status.Usage["cpu"][0].Values["AveUtil"])
//	memoryUsage := getFloatValue(node.Status.Usage["memory"][0].Values["Usage"])
//	storageUsage := getFloatValue(node.Status.Usage["storage"][0].Values["Usage"])
//	//logs.Infof("检查任务状态----CPU利用率：%v,内存利用率：%v，存储利用率：%v", cpuAveUtil, memoryUsage, storageUsage)
//	return cpuAveUtil > thresholdCPU || memoryUsage > thresholdMemory || storageUsage > thresholdStorage
//}
//
//type EndNodeSwitchCondition struct {
//}
//
//func (e *EndNodeSwitchCondition) CheckCondition(node *apis.Node) bool {
//	cpuAveUtil := getFloatValue(node.Status.Usage["cpu"][0].Values["AveUtil"])
//	memoryUsage := getFloatValue(node.Status.Usage["memory"][0].Values["Usage"])
//	storageUsage := getFloatValue(node.Status.Usage["storage"][0].Values["Usage"])
//
//	return cpuAveUtil > thresholdCPU || memoryUsage > thresholdMemory || storageUsage > thresholdStorage
//}
//
//func getFloatValue(s string) float64 {
//	value, err := strconv.ParseFloat(s, 64)
//	if err != nil {
//		logs.Errorf("Convert value failed, err:%v", err)
//		return 0
//	}
//	return value
//}
