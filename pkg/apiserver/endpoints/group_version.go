package endpoints

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"go.uber.org/zap"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/component-base/logs"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"path"
	"strings"
	"time"
)

// 这里设置API访问路径
// API Group配置
// example: /apis/{group}/{version}/{resources}/

type GroupVersion struct {
	Group   string
	Version string
}

// WithKind 基于方法接收者的GroupVersion和传递的Kind创建一个GroupVersionKind。
func (gv GroupVersion) WithKind(kind string) GroupVersionKind {
	return GroupVersionKind{Group: gv.Group, Version: gv.Version, Kind: kind}
}

// WithResource 基于方法接收者的GroupVersion和传递的Resource创建一个GroupVersionResource。
func (gv GroupVersion) WithResource(resource string) GroupVersionResource {
	return GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: resource}
}

func (gv GroupVersion) String() string {
	if len(gv.Group) > 0 {
		return gv.Group + "/" + gv.Version
	}
	return gv.Version
}

// GroupResource 指定组和资源
type GroupResource struct {
	Group    string
	Resource string
}

func (gr GroupResource) WithVersion(version string) GroupVersionResource {
	return GroupVersionResource{Group: gr.Group, Version: version, Resource: gr.Resource}
}

func (gr GroupResource) Empty() bool {
	return len(gr.Group) == 0 && len(gr.Resource) == 0
}

func (gr GroupResource) String() string {
	if len(gr.Group) == 0 {
		return gr.Resource
	}
	return gr.Resource + "." + gr.Group
}

// GroupVersionResource 指定组版本资源
type GroupVersionResource struct {
	Group    string
	Version  string
	Resource string
}

func (gvr GroupVersionResource) Empty() bool {
	return len(gvr.Group) == 0 && len(gvr.Version) == 0 && len(gvr.Resource) == 0
}

func (gvr GroupVersionResource) GroupResource() GroupResource {
	return GroupResource{Group: gvr.Group, Resource: gvr.Resource}
}

func (gvr GroupVersionResource) GroupVersion() GroupVersion {
	return GroupVersion{Group: gvr.Group, Version: gvr.Version}
}

func (gvr GroupVersionResource) String() string {
	return strings.Join([]string{gvr.Group, "/", gvr.Version, ", Resource=", gvr.Resource}, "")
}

// GroupKind 指定组和类型
type GroupKind struct {
	Group string
	Kind  string
}

func (gk GroupKind) Empty() bool {
	return len(gk.Group) == 0 && len(gk.Kind) == 0
}

func (gk GroupKind) WithVersion(version string) GroupVersionKind {
	return GroupVersionKind{Group: gk.Group, Version: version, Kind: gk.Kind}
}

func (gk GroupKind) String() string {
	if len(gk.Group) == 0 {
		return gk.Kind
	}
	return gk.Kind + "." + gk.Group
}

// GroupVersionKind 指定组版本类型
type GroupVersionKind struct {
	Group   string
	Version string
	Kind    string
}

// Empty returns true if group, version, and kind are empty
func (gvk GroupVersionKind) Empty() bool {
	return len(gvk.Group) == 0 && len(gvk.Version) == 0 && len(gvk.Kind) == 0
}

func (gvk GroupVersionKind) GroupKind() GroupKind {
	return GroupKind{Group: gvk.Group, Kind: gvk.Kind}
}

func (gvk GroupVersionKind) GroupVersion() GroupVersion {
	return GroupVersion{Group: gvk.Group, Version: gvk.Version}
}

func (gvk GroupVersionKind) String() string {
	return gvk.Group + "/" + gvk.Version + ", Kind=" + gvk.Kind
}

type APIGroupVersion struct {
	Storage map[string]rest.Storage

	Root string

	GroupVersion schema.GroupVersion

	MetaGroupVersion *schema.GroupVersion

	ParameterCodec runtime.ParameterCodec
	Serializer     runtime.NegotiatedSerializer
	Typer          runtime.ObjectTyper
	Creater        runtime.ObjectCreater
	Namer          runtime.Namer
	Convertor      runtime.ObjectConvertor
	Defaulter      runtime.ObjectDefaulter

	MinRequestTimeout time.Duration

	MaxRequestBodyBytes int64
}

// InstallREST 安装所有REST节点
func (g *APIGroupVersion) InstallREST(container *restful.Container) error {
	prefix := path.Join(g.Root, g.GroupVersion.Group, g.GroupVersion.Version)
	installer := &APIInstaller{
		group:             g,
		prefix:            prefix,
		minRequestTimeout: g.MinRequestTimeout,
	}
	logs.Debug("APIInstaller created", zap.String("prefix", installer.prefix))

	// 注册所有资源及其Handler
	ws, errs := installer.Install()
	if errs != nil {
		logs.Error("install APIInstaller failed", zap.String("error", utilerrors.NewAggregate(errs).Error()))
		return utilerrors.NewAggregate(errs)
	}

	// 将Webservice注册到HTTP服务器中
	container.Add(ws)

	// 生成swagger文档
	config := restfulspec.Config{
		WebServices:                   container.RegisteredWebServices(),
		APIPath:                       "/apidocs.json",
		PostBuildSwaggerObjectHandler: enrichSwaggerObject}
	container.Add(restfulspec.NewOpenAPIService(config))

	return utilerrors.NewAggregate(errs)
}
