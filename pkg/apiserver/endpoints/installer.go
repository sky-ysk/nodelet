package endpoints

import (
	"fmt"
	"github.com/emicklei/go-restful/v3"
	"go.uber.org/zap"
	"hit.edu/framework/pkg/apimachinery/conversion"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apiserver/endpoints/handler"
	negotiation "hit.edu/framework/pkg/apiserver/endpoints/handler/negotitation"
	"hit.edu/framework/pkg/apiserver/endpoints/handler/types"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"k8s.io/apimachinery/pkg/util/sets"
	"unicode"

	"net/http"

	"hit.edu/framework/pkg/apimachinery/runtime"

	"reflect"
	"sort"
	"strings"
	"time"
)

const (
	RouteMetaGVK    = "databus-group-version-kind"
	RouteMetaAction = "databus-action"
)

type APIInstaller struct {
	group             *APIGroupVersion
	prefix            string // Path prefix where API resources are to be registered.
	minRequestTimeout time.Duration
}

// TODO: 命名
// TODO: 增加Namespace支持
type action struct {
	Verb          string               // Verb 标识动词 ("GET", "POST", "WATCH", "PROXY", etc).
	Path          string               // 动词路径
	Namer         handler.ScopeNamer   // 从请求和runtime.object中获取名称
	Params        []*restful.Parameter // 和动词关联的参数
	AllNamespaces bool                 // 如果动词是命名空间的，但适用于所有命名空间的聚合结果，则为 true
}

var verbsMap = map[string]string{
	"DELETE":           "delete",
	"DELETECOLLECTION": "deletecollection",
	"GET":              "get",
	"POST":             "create",
	"PUT":              "update",
	"PROXY":            "proxy",
	"LIST":             "list",
	"PATCH":            "patch",
	"WATCH":            "watch",
	"WATCHLIST":        "watch",
}

type documentable interface {
	SwaggerDoc() map[string]string
}

func (a *APIInstaller) Install() (*restful.WebService, []error) {
	// TODO: 增加存储相关节点

	var apiResources []meta.APIResource
	var errors []error
	// WebService
	ws := a.newWebService()

	paths := make([]string, len(a.group.Storage))
	var i int = 0
	for path := range a.group.Storage {
		paths[i] = path
		i++
	}
	sort.Strings(paths)

	for _, path := range paths {
		logs.Debug("register Resource Handlers", zap.String("resource path", a.prefix+"/"+path))
		apiResource, err := a.registerResourceHandlers(path, a.group.Storage[path], ws)
		if apiResource != nil {
			apiResources = append(apiResources, *apiResource)
		}
		if err != nil {
			logs.Error("register Resource Handlers failed", zap.String("resource path", path), zap.String("error", err.Error()))
			errors = append(errors, fmt.Errorf("error in registering resource: %s, %v", path, err))
		}
	}

	return ws, errors

}

func (a *APIInstaller) registerResourceHandlers(path string, storage rest.Storage, ws *restful.WebService) (*meta.APIResource, error) {
	optionsExternalVersion := a.group.GroupVersion
	if a.group.MetaGroupVersion != nil {
		optionsExternalVersion = *a.group.MetaGroupVersion
	}
	//获取资源、子资源（资源的status等子资源）、组和版本名称
	resource, subresource, err := splitSubresource(path)
	if err != nil {
		logs.Error("splitSubresource failed,path:"+path, zap.Error(err))
		return nil, nil
	}
	isSubresource := len(subresource) > 0
	group, version := a.group.GroupVersion.Group, a.group.GroupVersion.Version
	//
	//创建特定Kind类型的对象实例
	fqKindToRegister, err := GetResourceKind(a.group.GroupVersion, storage, a.group.Typer)
	if err != nil {
		logs.Error("get resource kind failed", zap.Error(err))
		return nil, err
	}
	versionedPtr, err := a.group.Creater.New(fqKindToRegister)
	if err != nil {
		logs.Error("create new versionedPtr failed", zap.Error(err))
		return nil, err
	}
	defaultVersionedObject := indirectArbitraryPointer(versionedPtr)
	kind := fqKindToRegister.Kind

	var apiResource meta.APIResource
	// 如果存在子资源，则命名空间范围由父资源定义
	var namespaceScoped bool

	if isSubresource {
		parentStorage, ok := a.group.Storage[resource]
		if !ok {
			return nil, fmt.Errorf("missing parent storage: %q", resource)
		}
		scoper, ok := parentStorage.(rest.NamespaceScopedStrategy)
		if !ok {
			return nil, fmt.Errorf("%q must implement scoper", resource)
		}
		namespaceScoped = scoper.NamespaceScoped()
	} else {
		scoper, ok := storage.(rest.NamespaceScopedStrategy)
		if !ok {
			return nil, fmt.Errorf("%q must implement scoper", resource)
		}
		namespaceScoped = scoper.NamespaceScoped()
	}

	//判断资源Storage实现了哪些操作接口，用来判断path路径支持哪些动词
	creater, isCreater := storage.(rest.Creater)
	namedCreater, isNamedCreater := storage.(rest.NamedCreater)
	getter, isGetter := storage.(rest.Getter)
	getterWithOptions, isGetterWithOptions := storage.(rest.GetterWithOptions)
	lister, isLister := storage.(rest.Lister)
	updater, isUpdater := storage.(rest.Updater)
	watcher, isWatcher := storage.(rest.Watcher)
	deleter, isDeleter := storage.(rest.GracefulDeleter)
	collectionDeleter, isCollectionDeleter := storage.(rest.CollectionDeleter)
	patcher, isPatcher := storage.(rest.Patcher)
	if isNamedCreater {
		isCreater = true
	}

	var versionedList interface{}
	if isLister {
		list := lister.NewList()
		listGVKs, _, err := a.group.Typer.ObjectKinds(list)
		if err != nil {
			logs.Error("get resource kind failed", zap.Error(err))
			return nil, err
		}
		versionedListPtr, err := a.group.Creater.New(a.group.GroupVersion.WithKind(listGVKs[0].Kind))
		if err != nil {
			logs.Error("create new versionedPtr failed", zap.Error(err))
			return nil, err
		}
		versionedList = indirectArbitraryPointer(versionedListPtr)
	}

	versionedListOptions, err := a.group.Creater.New(optionsExternalVersion.WithKind("ListOptions"))
	if err != nil {
		logs.Error("create new ListOptions failed", zap.Error(err))
		return nil, err
	}
	versionedCreateOptions, err := a.group.Creater.New(optionsExternalVersion.WithKind("CreateOptions"))
	if err != nil {
		logs.Error("create new CreateOptions failed", zap.Error(err))
		return nil, err
	}
	versionedPatchOptions, err := a.group.Creater.New(optionsExternalVersion.WithKind("PatchOptions"))
	if err != nil {
		logs.Error("create new PatchOptions failed", zap.Error(err))
		return nil, err
	}
	versionedUpdateOptions, err := a.group.Creater.New(optionsExternalVersion.WithKind("UpdateOptions"))
	if err != nil {
		logs.Error("create new UpdateOptions failed", zap.Error(err))
		return nil, err
	}

	var versionedDeleteOptions runtime.Object
	var versionedDeleterObject interface{}
	if isDeleter {
		versionedDeleteOptions, err = a.group.Creater.New(optionsExternalVersion.WithKind("DeleteOptions"))
		if err != nil {
			logs.Error("create new DeleteOptions failed", zap.Error(err))
			return nil, err
		}
		versionedDeleterObject = indirectArbitraryPointer(versionedDeleteOptions)
	}

	var getSubpath bool

	if isGetterWithOptions {
		_, getSubpath, _ = getterWithOptions.NewGetOptions()
		isGetter = true
	}

	versionedStatusPtr, err := a.group.Creater.New(optionsExternalVersion.WithKind("Status"))
	if err != nil {
		logs.Error("create new Status failed", zap.Error(err))
		return nil, err
	}
	versionedStatus := indirectArbitraryPointer(versionedStatusPtr)

	//为API资源定义路径参数和处理动作
	nameParam := ws.PathParameter("name", "name of the "+kind).DataType("string")
	pathParam := ws.PathParameter("path", "path of the "+kind).DataType("string")
	params := []*restful.Parameter{}
	actions := []action{}

	switch {
	case !namespaceScoped:
		//获取支持的action列表
		resourcePath := resource
		resourceParams := params
		itemPath := resourcePath + "/{name}"
		nameParams := append(params, nameParam)
		proxyParams := append(params, pathParam)
		suffix := ""
		if isSubresource {
			suffix = "/" + subresource
			itemPath = itemPath + suffix
			resourcePath = itemPath
			resourceParams = nameParams
		}

		apiResource.Namespaced = false
		namer := handler.ContextBasedNaming{Namer: a.group.Namer, ClusterScoped: true}

		//标准REST动词（GET、PUT、POST和DELETE）的处理
		//在资源路径"resources/v1/resource"下添加动词
		actions = appendIf(actions, action{"POST", resourcePath, namer, resourceParams, false}, isCreater)
		actions = appendIf(actions, action{"LIST", resourcePath, namer, resourceParams, false}, isLister)
		actions = appendIf(actions, action{"DELETECOLLECTION", resourcePath, namer, resourceParams, false}, isDeleter)
		//在资源路径"resources/v1/resource/{name}"下添加动词
		actions = appendIf(actions, action{"GET", itemPath, namer, nameParams, false}, isGetter)
		if getSubpath {
			actions = appendIf(actions, action{"GET", itemPath + "/{path:*}", namer, proxyParams, false}, isGetter)
		}
		actions = appendIf(actions, action{"PUT", itemPath, namer, nameParams, false}, isUpdater)
		actions = appendIf(actions, action{"DELETE", itemPath, namer, nameParams, false}, isDeleter)
		actions = appendIf(actions, action{"PATCH", itemPath, namer, nameParams, false}, isPatcher)
	default:
		namespaceParamName := "namespaces"
		namespaceParam := ws.PathParameter("namespace", "object name and auth scope, such as for teams and projects").DataType("string")
		namespacedPath := namespaceParamName + "/{namespace}/" + resource
		namespaceParams := []*restful.Parameter{namespaceParam}

		resourcePath := namespacedPath
		resourceParams := namespaceParams
		itemPath := namespacedPath + "/{name}"
		nameParams := append(namespaceParams, nameParam)
		proxyParams := append(nameParams, pathParam)
		itemPathSuffix := ""
		if isSubresource {
			itemPathSuffix = "/" + subresource
			itemPath = itemPath + itemPathSuffix
			resourcePath = itemPath
			resourceParams = nameParams
		}

		apiResource.Namespaced = true
		namer := handler.ContextBasedNaming{Namer: a.group.Namer, ClusterScoped: false}

		//标准REST动词（GET、PUT、POST和DELETE）的处理
		//在资源路径"resources/v1/resource"下添加动词
		actions = appendIf(actions, action{"POST", resourcePath, namer, resourceParams, false}, isCreater)
		actions = appendIf(actions, action{"LIST", resourcePath, namer, resourceParams, false}, isLister)
		actions = appendIf(actions, action{"DELETECOLLECTION", resourcePath, namer, resourceParams, false}, isDeleter)
		//在资源路径"resources/v1/resource/{name}"下添加动词
		actions = appendIf(actions, action{"GET", itemPath, namer, nameParams, false}, isGetter)
		if getSubpath {
			actions = appendIf(actions, action{"GET", itemPath + "/{path:*}", namer, proxyParams, false}, isGetter)
		}
		actions = appendIf(actions, action{"PUT", itemPath, namer, nameParams, false}, isUpdater)
		actions = appendIf(actions, action{"DELETE", itemPath, namer, nameParams, false}, isDeleter)
		actions = appendIf(actions, action{"PATCH", itemPath, namer, nameParams, false}, isPatcher)

		// list or post across namespace.
		// For ex: LIST all pods in all namespaces by sending a LIST request at /api/apiVersion/pods.
		// TODO: more strongly type whether a resource allows these actions on "all namespaces" (bulk delete)
		if !isSubresource {
			actions = appendIf(actions, action{"LIST", resource, namer, params, true}, isLister)
		}
	}
	//为每个动词创建路由
	//为ws设置支持的媒体类型
	//for _, s := range a.group.Serializer.SupportedMediaTypes() {
	//	if len(s.MediaTypeSubType) == 0 || len(s.MediaTypeType) == 0 {
	//		return nil, fmt.Errorf("all serializers in the group Serializer must have MediaTypeType and MediaTypeSubType set: %s", s.MediaType)
	//	}
	//}
	mediaTypes, streamMediaTypes := negotiation.MediaTypesForSerializer(a.group.Serializer)
	allMediaTypes := append(mediaTypes, streamMediaTypes...)
	ws.Produces(allMediaTypes...)

	//构造公共字段放入reqScope中
	verbs := map[string]struct{}{}
	reqScope := handler.RequestScope{
		Serializer:     a.group.Serializer,
		ParameterCodec: a.group.ParameterCodec,
		Creater:        a.group.Creater,
		Convertor:      a.group.Convertor,
		Typer:          a.group.Typer,

		Resource:    a.group.GroupVersion.WithResource(resource),
		Subresource: subresource,
		Kind:        fqKindToRegister,

		MetaGroupVersion: meta.SchemeGroupVersion,
		HubGroupVersion:  schema.GroupVersion{Group: fqKindToRegister.Group, Version: fqKindToRegister.Version},

		MaxRequestBodyBytes: a.group.MaxRequestBodyBytes,
	}

	//路由注册
	for _, action := range actions {
		var producedObject interface{}
		producedObject = defaultVersionedObject
		reqScope.Namer = action.Namer
		//requestScope := ""
		//if strings.Contains(action.Path, "/{name}") || action.Verb == "POST" {
		//	requestScope = "resource"
		//}
		//requestScope := "cluster"
		var namespaced string
		var operationSuffix string
		if apiResource.Namespaced {
			//requestScope = "namespace"
			namespaced = "Namespaced"
		}
		if strings.HasSuffix(action.Path, "/{path:*}") {
			//requestScope = "resource"
			operationSuffix = operationSuffix + "WithPath"
		}
		//if strings.Contains(action.Path, "/{name}") || action.Verb == "POST" {
		//	requestScope = "resource"
		//}
		if action.AllNamespaces {
			//requestScope = "cluster"
			operationSuffix = operationSuffix + "ForAllNamespaces"
			namespaced = ""
		}

		routes := []*restful.RouteBuilder{}

		if verb, found := verbsMap[action.Verb]; found {
			if len(verb) != 0 {
				verbs[verb] = struct{}{}
			}
		} else {
			return nil, fmt.Errorf("unknown action verb for discovery: %s", action.Verb)
		}

		if isSubresource {
			parentStorage, ok := a.group.Storage[resource]
			if !ok {
				return nil, fmt.Errorf("missing parent storage: %q", resource)
			}

			fqParentKind, err := GetResourceKind(a.group.GroupVersion, parentStorage, a.group.Typer)
			if err != nil {
				return nil, err
			}
			kind = fqParentKind.Kind
		}

		var handler restful.RouteFunction
		switch action.Verb {
		case "LIST": //列出所有资源
			handler = restfulListResource(lister, watcher, reqScope, a.minRequestTimeout)

			if isSubresource {
				continue
			}
			doc := "list objects of kind " + kind
			if isSubresource {
				doc = "list " + subresource + " of objects of kind " + kind
			}
			route := ws.GET(action.Path).To(handler).
				Doc(doc).
				Operation("list"+namespaced+kind+strings.Title(subresource)+operationSuffix).
				Produces(mediaTypes...).
				Returns(http.StatusOK, "OK", versionedList).
				Writes(versionedList)
			if err := AddObjectParams(ws, route, versionedListOptions); err != nil {
				return nil, err
			}
			switch {
			case isLister && isWatcher:
				doc := "list or watch objects of kind " + kind
				if isSubresource {
					doc = "list or watch " + subresource + " of objects of kind " + kind
				}
				route.Doc(doc)
				verbs["watchlist"] = struct{}{}
				verbs["watch"] = struct{}{}
			case isWatcher:
				doc := "watch objects of kind " + kind
				if isSubresource {
					doc = "watch " + subresource + "of objects of kind " + kind
				}
				route.Doc(doc)
				verbs["watchlist"] = struct{}{}
				verbs["watch"] = struct{}{}
			}
			addParams(route, action.Params)
			routes = append(routes, route)
		case "GET": //查找一个资源
			handler = restfulGetResource(getter, reqScope)

			doc := "read the specified " + kind
			if isSubresource {
				doc = "read " + subresource + " of the specified " + kind
			}
			route := ws.GET(action.Path).To(handler).
				Doc(doc).
				Operation("read"+namespaced+kind+strings.Title(subresource)+operationSuffix).
				Produces(mediaTypes...).
				Returns(http.StatusOK, "OK", producedObject).
				Writes(producedObject)
			addParams(route, action.Params)
			routes = append(routes, route)
		case "POST": //创建一个资源
			if isNamedCreater {
				handler = restfulCreateNamedResource(namedCreater, reqScope)
			} else {
				handler = restfulCreateResource(creater, reqScope)
			}
			article := GetArticleForNoun(kind, " ")
			doc := "create" + article + kind
			if isSubresource {
				doc = "create " + subresource + " of " + kind
			}
			route := ws.POST(action.Path).To(handler).
				Doc(doc).
				Operation("create"+namespaced+kind+strings.Title(subresource)+operationSuffix).
				Produces(mediaTypes...).
				Returns(http.StatusOK, "OK", producedObject).
				Returns(http.StatusCreated, "Created", producedObject).
				Reads(producedObject).
				Writes(producedObject)
			if err := AddObjectParams(ws, route, versionedCreateOptions); err != nil {
				return nil, err
			}
			addParams(route, action.Params)
			routes = append(routes, route)
		case "PUT": //修改一个资源
			handler = restfulUpdateResource(updater, reqScope)

			doc := "replace the specified " + kind
			if isSubresource {
				doc = "replace " + subresource + " of the specified " + kind
			}
			route := ws.PUT(action.Path).To(handler).
				Doc(doc).
				Operation("replace"+namespaced+kind+strings.Title(subresource)+operationSuffix).
				Produces(mediaTypes...).
				Returns(http.StatusOK, "OK", producedObject).
				Returns(http.StatusCreated, "Created", producedObject).
				Reads(producedObject).
				Writes(producedObject)
			if err := AddObjectParams(ws, route, versionedUpdateOptions); err != nil {
				return nil, err
			}
			addParams(route, action.Params)
			routes = append(routes, route)
		case "DELETE": //删除一个资源
			handler = restfulDeleteResource(deleter, isDeleter, reqScope)
			article := GetArticleForNoun(kind, " ")
			doc := "delete" + article + kind
			if isSubresource {
				doc = "delete " + subresource + " of " + kind
			}

			route := ws.DELETE(action.Path).To(handler).
				Doc(doc).
				Operation("delete"+namespaced+kind+strings.Title(subresource)+operationSuffix).
				Produces(mediaTypes...).
				Returns(http.StatusOK, "OK", producedObject).
				Writes(producedObject)
			if isDeleter {
				route.Reads(versionedDeleterObject)
				route.ParameterNamed("body").Required(false)
				if err := AddObjectParams(ws, route, versionedDeleteOptions); err != nil {
					return nil, err
				}
			}
			addParams(route, action.Params)
			routes = append(routes, route)
		case "DELETECOLLECTION":
			handler := restfulDeleteCollection(collectionDeleter, isCollectionDeleter, reqScope)
			doc := "delete collection of " + kind
			if isSubresource {
				doc = "delete collection of " + subresource + " of a " + kind
			}
			route := ws.DELETE(action.Path).To(handler).
				Doc(doc).
				Operation("deletecollection"+namespaced+kind+strings.Title(subresource)+operationSuffix).
				Produces(mediaTypes...).
				Writes(producedObject).
				Returns(http.StatusOK, "OK", versionedStatus)
			if isCollectionDeleter {
				route.Reads(versionedDeleterObject)
				route.ParameterNamed("body").Required(false)
				if err := AddObjectParams(ws, route, versionedDeleteOptions); err != nil {
					return nil, err
				}
			}
			if err := AddObjectParams(ws, route, versionedListOptions, "watch", "allowWatchBookmarks"); err != nil {
				return nil, err
			}
			addParams(route, action.Params)
			routes = append(routes, route)
		case "PATCH":
			supportedTypes := []string{
				string(types.JSONPatchType),
				string(types.MergePatchType),
			}
			handler := restfulPatchResource(patcher, reqScope, supportedTypes)
			doc := "partially update the specified " + kind
			if isSubresource {
				doc = "partially update " + subresource + " of the specified " + kind
			}
			route := ws.PATCH(action.Path).To(handler).
				Doc(doc).
				Operation("patch"+namespaced+kind+strings.Title(subresource)+operationSuffix).
				Consumes(supportedTypes...).
				Produces(mediaTypes...).
				Returns(http.StatusOK, "OK", producedObject).
				Writes(producedObject)
			if err := AddObjectParams(ws, route, versionedPatchOptions); err != nil {
				return nil, err
			}
			addParams(route, action.Params)
			routes = append(routes, route)
		default:
			return nil, fmt.Errorf("unrecognized action verb: %s", action.Verb)
		}
		for _, route := range routes {
			route.Metadata(RouteMetaGVK, meta.GroupVersionKind{
				Group:   reqScope.Kind.Group,
				Version: reqScope.Kind.Version,
				Kind:    reqScope.Kind.Kind,
			})
			route.Metadata(RouteMetaAction, strings.ToLower(action.Verb))
			ws.Route(route)
		}
	}

	//生成apiResource并返回

	apiResource.Name = path
	apiResource.Group = group
	apiResource.Version = version
	apiResource.Name = path
	apiResource.Kind = kind
	apiResource.Verbs = make([]string, 0, len(verbs))
	for verb := range verbs {
		apiResource.Verbs = append(apiResource.Verbs, verb)
	}
	sort.Strings(apiResource.Verbs)
	//if shortNamesProvider, ok := storage.(rest.ShortNamesProvider); ok {
	//	apiResource.ShortName = shortNamesProvider.ShortNames()
	//}
	logs.Debug("install restfulAPI for resource " + apiResource.Name + " done,supported verbs:" + apiResource.Verbs.String())
	return &apiResource, nil
}

// 创建Webservice
// Director重定向服务到Webservice
func (a *APIInstaller) newWebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.Path(a.prefix)
	ws.Doc("API at " + a.prefix)
	ws.Consumes("*/*")

	mediaTypes, streamMediaTypes := negotiation.MediaTypesForSerializer(a.group.Serializer)
	ws.Produces(append(mediaTypes, streamMediaTypes...)...)
	ws.ApiVersion(a.group.GroupVersion.String())
	return ws
}

// splitSubresource 从路径中获取资源和子资源（若有）名称
func splitSubresource(path string) (string, string, error) {
	var resource, subresource string
	switch parts := strings.Split(path, "/"); len(parts) {
	case 2:
		resource, subresource = parts[0], parts[1]
	case 1:
		resource = parts[0]
	default:
		return "", "", fmt.Errorf("api_installer allows only one or two segment paths (resource or resource/subresource)")
	}
	return resource, subresource, nil
}

// GetResourceKind 返回资源的GroupVersionKind
func GetResourceKind(groupVersion schema.GroupVersion, storage rest.Storage, typer runtime.ObjectTyper) (schema.GroupVersionKind, error) {
	object := storage.New()
	fqKinds, _, err := typer.ObjectKinds(object)
	if err != nil {
		return schema.GroupVersionKind{}, err
	}

	//一个给定的go类型可以有多个潜在的完全限定类型。找到与该组对应的那个
	fqKindToRegister := schema.GroupVersionKind{}
	for _, fqKind := range fqKinds {
		if fqKind.Group == groupVersion.Group {
			fqKindToRegister = groupVersion.WithKind(fqKind.Kind)
			break
		}
	}
	if fqKindToRegister.Empty() {
		return schema.GroupVersionKind{}, fmt.Errorf("unable to locate fully qualified kind for %v: found %v when registering for %v", reflect.TypeOf(object), fqKinds, groupVersion)
	}

	return fqKindToRegister, nil
}

// indirectArbitraryPointer 对于任意指针返回解引用后的值
func indirectArbitraryPointer(ptrToObject interface{}) interface{} {
	return reflect.Indirect(reflect.ValueOf(ptrToObject)).Interface()
}

func GetArticleForNoun(noun string, padding string) string {
	if !strings.HasSuffix(noun, "ss") && strings.HasSuffix(noun, "s") {
		// Plurals don't have an article.
		// Don't catch words like class
		return fmt.Sprintf("%v", padding)
	}

	article := "a"
	if isVowel(rune(noun[0])) {
		article = "an"
	}

	return fmt.Sprintf("%s%s%s", padding, article, padding)
}
func isVowel(c rune) bool {
	vowels := []rune{'a', 'e', 'i', 'o', 'u'}
	for _, value := range vowels {
		if value == unicode.ToLower(c) {
			return true
		}
	}
	return false
}

func AddObjectParams(ws *restful.WebService, route *restful.RouteBuilder, obj interface{}, excludedNames ...string) error {
	sv, err := conversion.EnforcePtr(obj)
	if err != nil {
		return err
	}
	st := sv.Type()
	excludedNameSet := sets.NewString(excludedNames...)
	switch st.Kind() {
	case reflect.Struct:
		for i := 0; i < st.NumField(); i++ {
			name := st.Field(i).Name
			sf, ok := st.FieldByName(name)
			if !ok {
				continue
			}
			switch sf.Type.Kind() {
			case reflect.Interface, reflect.Struct:
			case reflect.Pointer:
				if (sf.Type.Elem().Kind() == reflect.Interface || sf.Type.Elem().Kind() == reflect.Struct) && strings.TrimPrefix(sf.Type.String(), "*") != "meta.Time" {
					continue
				}
				fallthrough
			default:
				jsonTag := sf.Tag.Get("json")
				if len(jsonTag) == 0 {
					continue
				}
				jsonName := strings.SplitN(jsonTag, ",", 2)[0]
				if len(jsonName) == 0 {
					continue
				}
				if excludedNameSet.Has(jsonName) {
					continue
				}
				var desc string
				if docable, ok := obj.(documentable); ok {
					desc = docable.SwaggerDoc()[jsonName]
				}
				route.Param(ws.QueryParameter(jsonName, desc).DataType(typeToJSON(sf.Type.String())))
			}
		}
	}
	return nil
}

func typeToJSON(typeName string) string {
	switch typeName {
	case "bool", "*bool":
		return "boolean"
	case "uint8", "*uint8", "int", "*int", "int32", "*int32", "int64", "*int64", "uint32", "*uint32", "uint64", "*uint64":
		return "integer"
	case "float64", "*float64", "float32", "*float32":
		return "number"
	case "meta.Time", "*meta.Time":
		return "string"
	case "byte", "*byte":
		return "string"
	case "meta.DeletionPropagation", "*meta.DeletionPropagation":
		return "string"
	case "meta.ResourceVersionMatch", "*meta.ResourceVersionMatch":
		return "string"
	case "meta.IncludeObjectPolicy", "*meta.IncludeObjectPolicy":
		return "string"

	// TODO: Fix these when go-restful supports a way to specify an array query param:
	// https://github.com/emicklei/go-restful/issues/225
	case "[]string", "[]*string":
		return "string"
	case "[]int32", "[]*int32":
		return "integer"

	default:
		return typeName
	}
}
func addParams(route *restful.RouteBuilder, params []*restful.Parameter) {
	for _, param := range params {
		route.Param(param)
	}
}

func appendIf(actions []action, a action, shouldAppend bool) []action {
	if shouldAppend {
		actions = append(actions, a)
	}
	return actions
}

func restfulGetResource(r rest.Getter, scope handler.RequestScope) restful.RouteFunction {
	return func(req *restful.Request, res *restful.Response) {
		handler.GetResource(r, &scope)(res.ResponseWriter, req.Request)
	}
}

func restfulCreateResource(r rest.Creater, scope handler.RequestScope) restful.RouteFunction {
	return func(req *restful.Request, res *restful.Response) {
		handler.CreateResource(r, &scope)(res.ResponseWriter, req.Request)
	}
}

func restfulCreateNamedResource(r rest.NamedCreater, scope handler.RequestScope) restful.RouteFunction {
	return func(req *restful.Request, res *restful.Response) {
		handler.CreateNamedResource(r, &scope)(res.ResponseWriter, req.Request)
	}
}

func restfulUpdateResource(r rest.Updater, scope handler.RequestScope) restful.RouteFunction {
	return func(req *restful.Request, res *restful.Response) {
		handler.UpdateResource(r, &scope)(res.ResponseWriter, req.Request)
	}
}

func restfulDeleteResource(r rest.GracefulDeleter, allowsOptions bool, scope handler.RequestScope) restful.RouteFunction {
	return func(req *restful.Request, res *restful.Response) {
		handler.DeleteResource(r, allowsOptions, &scope)(res.ResponseWriter, req.Request)
	}
}

func restfulListResource(r rest.Lister, rw rest.Watcher, scope handler.RequestScope, minRequestTimeout time.Duration) restful.RouteFunction {
	return func(req *restful.Request, res *restful.Response) {
		handler.ListResource(r, rw, &scope, minRequestTimeout)(res.ResponseWriter, req.Request)
	}
}

func restfulDeleteCollection(r rest.CollectionDeleter, checkBody bool, scope handler.RequestScope) restful.RouteFunction {
	return func(req *restful.Request, res *restful.Response) {
		handler.DeleteCollection(r, checkBody, &scope)(res.ResponseWriter, req.Request)
	}
}

func restfulPatchResource(r rest.Patcher, scope handler.RequestScope, supportedTypes []string) restful.RouteFunction {
	return func(req *restful.Request, res *restful.Response) {
		handler.PatchResource(r, &scope, supportedTypes)(res.ResponseWriter, req.Request)
	}
}
