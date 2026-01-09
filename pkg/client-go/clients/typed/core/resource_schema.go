package core

import (
	"context"

	"hit.edu/framework/pkg/apimachinery/types"
	"hit.edu/framework/pkg/apimachinery/watch"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/scheme"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/gentype"
)

const (
	ResourceSchemaResource string = "resourceschemas"
)

type ResourceSchemasGetter interface {
	ResourceSchemas(namespace string) ResourceSchemaInterface
}

type ResourceSchemaInterface interface {
	Create(ctx context.Context, resourceSchema *apis.ResourceSchema, opts metav1.CreateOptions) (*apis.ResourceSchema, error)
	Update(ctx context.Context, resourceSchema *apis.ResourceSchema, opts metav1.UpdateOptions) (*apis.ResourceSchema, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*apis.ResourceSchema, error)
	List(ctx context.Context, opts metav1.ListOptions) (*apis.ResourceSchemaList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *apis.ResourceSchema, err error)
}

type resourceSchemas struct {
	*gentype.ClientWithList[*apis.ResourceSchema, *apis.ResourceSchemaList]
}

// newResourceSchemas returns a ResourceSchemas
func newResourceSchemas(c *CoreClient, namespace string) *resourceSchemas {
	return &resourceSchemas{
		gentype.NewClientWithList[*apis.ResourceSchema, *apis.ResourceSchemaList](
			ResourceSchemaResource,
			c.RESTClient(),
			namespace,
			scheme.ParameterCodec,
			func() *apis.ResourceSchema { return &apis.ResourceSchema{} },
			func() *apis.ResourceSchemaList { return &apis.ResourceSchemaList{} },
		),
	}
}
