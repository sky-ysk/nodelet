package core

import (
	context "context"

	"hit.edu/framework/pkg/apimachinery/types"
	"hit.edu/framework/pkg/apimachinery/watch"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apis/meta/internalversion/scheme"
	"hit.edu/framework/pkg/client-go/gentype"
)

const (
	EventResource string = "events"
)

type EventsGetter interface {
	Events(namespace string) EventInterface
}

// 该接口包含操作Event资源的一系列方法，events为实现类
type EventInterface interface {
	Create(ctx context.Context, event *apis.Event, opts meta.CreateOptions) (*apis.Event, error)
	Update(ctx context.Context, event *apis.Event, opts meta.UpdateOptions) (*apis.Event, error)
	Delete(ctx context.Context, name string, opts meta.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts meta.DeleteOptions, listOpts meta.ListOptions) error
	Get(ctx context.Context, name string, opts meta.GetOptions) (*apis.Event, error)
	List(ctx context.Context, opts meta.ListOptions) (*apis.EventList, error)
	Watch(ctx context.Context, opts meta.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts meta.PatchOptions, subresources ...string) (result *apis.Event, err error)

	// Watch(ctx context.Context, opts runtime.ListOptions) (watch.Interface, error)
	EventExpansion
}

// 实现EventInterface
type events struct {
	*gentype.ClientWithList[*apis.Event, *apis.EventList]
}

// returns a Events
func newEvents(c *CoreClient, namespace string) *events {
	return &events{
		gentype.NewClientWithList[*apis.Event, *apis.EventList](
			EventResource,
			c.RESTClient(),
			namespace,
			scheme.ParameterCodec,
			func() *apis.Event { return &apis.Event{} },
			func() *apis.EventList { return &apis.EventList{} },
		),
	}
}
