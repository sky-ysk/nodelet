package meta

import metav1 "hit.edu/framework/pkg/apis/meta"

type ListMetaAccessor interface {
	GetListMeta() List
}

type List metav1.ListMeta
