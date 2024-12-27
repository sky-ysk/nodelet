package cores

import "hit.edu/framework/pkg/client-go/tools/cache"

type NodeInformer interface {
	Informer() cache.SharedInformer
}
