package cores

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/listers"
	"hit.edu/framework/pkg/client-go/tools/cache"
)

type NodeLister interface {
	// 列出所有Node
	List() (ret []*apis.Node, err error)

	// TODO: 处理Namespace
	// TODO: 处理Labels
}

type nodeLister struct {
	listers.ResourceIndexer[*apis.Node]
}

func NewNodeLister(indexer cache.Indexer) NodeLister {
	return &nodeLister{listers.New[*apis.Node](indexer, apis.Resource("node"))}
}
