package main

import (
	"context"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/scheduler/utils"
)

func main() {
	cs, err := utils.CreateClientSetWithTimeOut(3600)
	if err != nil {
		panic(err)
	}
	nc := cs.Core().Nodes(apis.NamespaceAll)
	lstOpts := metav1.ListOptions{}
	list, err := nc.List(context.TODO(), lstOpts)
	if err != nil {
		panic(err)
	}
	nodes := make([]apis.Node, 0)
	for _, n := range list.Items {
		//info := config.NewNodeInfo(&n)
		nodes = append(nodes, n)
		fmt.Println(n.Name)
	}
	for _, nn := range nodes {
		get, err := nc.Get(context.TODO(), nn.Name, metav1.GetOptions{})
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(get.Name)

	}
}
