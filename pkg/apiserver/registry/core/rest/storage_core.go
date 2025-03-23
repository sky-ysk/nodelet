package rest

import (
	"hit.edu/framework/pkg/apis/legacyscheme"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apiserver/registry/generic"
	"hit.edu/framework/pkg/apiserver/server"

	//genericapiserver "k8s.io/apiserver/pkg/server"
	actionstore "hit.edu/framework/pkg/apiserver/registry/core/action"
	datastore "hit.edu/framework/pkg/apiserver/registry/core/data"
	devicestore "hit.edu/framework/pkg/apiserver/registry/core/device"
	eventstore "hit.edu/framework/pkg/apiserver/registry/core/event"
	groupstore "hit.edu/framework/pkg/apiserver/registry/core/group"
	nodestore "hit.edu/framework/pkg/apiserver/registry/core/node"
	resourcestore "hit.edu/framework/pkg/apiserver/registry/core/resource"
	scenestore "hit.edu/framework/pkg/apiserver/registry/core/scene"
	taskstore "hit.edu/framework/pkg/apiserver/registry/core/task"
	workflowstore "hit.edu/framework/pkg/apiserver/registry/core/workflow"
	"hit.edu/framework/pkg/apiserver/registry/rest"
	"hit.edu/framework/pkg/component-base/logs"
)

// func NewRESTStorage(apiResourceConfigSource serverstorage.APIResourceConfigSource, restOptionsGetter generic.RESTOptionsGetter) (server.APIGroupInfo, error) {
func NewRESTStorage(restOptionsGetter generic.RESTOptionsGetter) (server.APIGroupInfo, error) {
	apiGroupInfo := server.APIGroupInfo{
		VersionedResourcesStorageMap: map[string]map[string]rest.Storage{},
		MetaGroupVersion:             &meta.SchemeGroupVersion,
		Scheme:                       legacyscheme.Scheme,
		ParameterCodec:               legacyscheme.ParameterCodec,
		NegotiatedSerializer:         legacyscheme.Codecs,
	}
	logs.Init("etcd")
	nodeStorage, err := nodestore.NewNodeStorage(restOptionsGetter)
	if err != nil {
		logs.Error("error occur while create NodeStorage", err)
		return server.APIGroupInfo{}, err
	}
	workflowStorage, err := workflowstore.NewWorkflowStorage(restOptionsGetter)
	if err != nil {
		logs.Error("error occur while create WorkflowStorage", err)
		return server.APIGroupInfo{}, err
	}
	taskStorage, err := taskstore.NewTaskStorage(restOptionsGetter)
	if err != nil {
		logs.Error("error occur while create TaskStorage", err)
		return server.APIGroupInfo{}, err
	}
	groupStorage, err := groupstore.NewGroupStorage(restOptionsGetter)
	if err != nil {
		logs.Error("error occur while create GroupStorage", err)
		return server.APIGroupInfo{}, err
	}
	actionStorage, err := actionstore.NewActionStorage(restOptionsGetter)
	if err != nil {
		logs.Error("error occur while create ActionStorage", err)
		return server.APIGroupInfo{}, err
	}
	deviceStorage, err := devicestore.NewDeviceStorage(restOptionsGetter)
	if err != nil {
		logs.Error("error occur while create DeviceStorage", err)
		return server.APIGroupInfo{}, err
	}
	sceneStorage, err := scenestore.NewSceneStorage(restOptionsGetter)
	if err != nil {
		logs.Error("error occur while create SceneStorage", err)
		return server.APIGroupInfo{}, err
	}
	resourceStorage, err := resourcestore.NewResourceStorage(restOptionsGetter)
	if err != nil {
		logs.Error("error occur while create ResourceStorage", err)
		return server.APIGroupInfo{}, err
	}
	eventStorage, err := eventstore.NewEventStorage(restOptionsGetter)
	if err != nil {
		logs.Error("error occur while create EventStorage", err)
		return server.APIGroupInfo{}, err
	}
	dataStorage, err := datastore.NewDataStorage(restOptionsGetter)
	if err != nil {
		logs.Error("error occur while create DataStorage", err)
		return server.APIGroupInfo{}, err
	}

	storage := map[string]rest.Storage{}
	if resource := "nodes"; true {
		storage[resource] = nodeStorage.Node
		storage[resource+"/status"] = nodeStorage.Status
		storage[resource+"/spec"] = nodeStorage.Spec
	}
	if resource := "workflows"; true {
		storage[resource] = workflowStorage.Workflow
		storage[resource+"/status"] = workflowStorage.Status
		storage[resource+"/spec"] = workflowStorage.Spec
	}
	if resource := "tasks"; true {
		storage[resource] = taskStorage.Task
		storage[resource+"/status"] = taskStorage.Status
		storage[resource+"/spec"] = taskStorage.Spec
	}
	if resource := "groups"; true {
		storage[resource] = groupStorage.Group
		storage[resource+"/status"] = groupStorage.Status
		storage[resource+"/spec"] = groupStorage.Spec
	}
	if resource := "actions"; true {
		storage[resource] = actionStorage.Action
		storage[resource+"/status"] = actionStorage.Status
		storage[resource+"/spec"] = actionStorage.Spec
	}
	if resource := "devices"; true {
		storage[resource] = deviceStorage.Device
		storage[resource+"/status"] = deviceStorage.Status
		storage[resource+"/spec"] = deviceStorage.Spec
	}
	if resource := "scenes"; true {
		storage[resource] = sceneStorage.Scene
		storage[resource+"/status"] = sceneStorage.Status
		storage[resource+"/spec"] = sceneStorage.Spec
	}
	if resource := "resource_nodes"; true {
		storage[resource] = resourceStorage.Resource
		storage[resource+"/status"] = resourceStorage.Status
		storage[resource+"/spec"] = resourceStorage.Spec
	}
	if resource := "datas"; true {
		storage[resource] = dataStorage.Data
		storage[resource+"/status"] = dataStorage.Status
		storage[resource+"/spec"] = dataStorage.Spec
	}
	if resource := "events"; true {
		storage[resource] = eventStorage.Event
	}
	if len(storage) > 0 {
		apiGroupInfo.VersionedResourcesStorageMap["v1"] = storage
	}
	logs.Info("RESTStorage create successfully")
	return apiGroupInfo, nil
}
