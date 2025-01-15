package rest

import (
	"hit.edu/framework/pkg/apis/legacyscheme"
	"hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/apiserver/registry/generic"
	"hit.edu/framework/pkg/apiserver/server"

	//genericapiserver "k8s.io/apiserver/pkg/server"
	actionstore "hit.edu/framework/pkg/apiserver/registry/core/action"
	eventstore "hit.edu/framework/pkg/apiserver/registry/core/event"
	groupstore "hit.edu/framework/pkg/apiserver/registry/core/group"
	nodestore "hit.edu/framework/pkg/apiserver/registry/core/node"
	taskstore "hit.edu/framework/pkg/apiserver/registry/core/task"
	workflowstore "hit.edu/framework/pkg/apiserver/registry/core/workflow"
	"hit.edu/framework/pkg/apiserver/registry/rest"
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
	nodeStorage, err := nodestore.NewNodeStorage(restOptionsGetter)
	if err != nil {
		return server.APIGroupInfo{}, err
	}
	workflowStorage, err := workflowstore.NewWorkflowStorage(restOptionsGetter)
	if err != nil {
		return server.APIGroupInfo{}, err
	}
	taskStorage, err := taskstore.NewTaskStorage(restOptionsGetter)
	if err != nil {
		return server.APIGroupInfo{}, err
	}
	groupStorage, err := groupstore.NewGroupStorage(restOptionsGetter)
	if err != nil {
		return server.APIGroupInfo{}, err
	}
	actionStorage, err := actionstore.NewActionStorage(restOptionsGetter)
	if err != nil {
		return server.APIGroupInfo{}, err
	}
	eventStorage, err := eventstore.NewEventStorage(restOptionsGetter)
	if err != nil {
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
	if resource := "events"; true {
		storage[resource] = eventStorage.Event
	}
	if len(storage) > 0 {
		apiGroupInfo.VersionedResourcesStorageMap["v1"] = storage
	}
	return apiGroupInfo, nil
}
