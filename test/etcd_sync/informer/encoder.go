package informer

import (
	"context"
	"fmt"

	"hit.edu/framework/pkg/apimachinery/runtime"
	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
)

func GetResourceObj(name string) (interface{}, error) {
	switch name {
	case "nodes":
		return &apis.Node{}, nil
	case "events":
		return &apis.Event{}, nil
	case "datas":
		return &apis.Data{}, nil
	case "workflows":
		return &apis.Workflow{}, nil
	case "tasks":
		return &apis.Task{}, nil
	case "groups":
		return &apis.Group{}, nil
	case "actions":
		return &apis.Action{}, nil
	case "scenes":
		return &apis.Scene{}, nil
	case "devices":
		return &apis.Device{}, nil
	case "resources":
		return &apis.Resource_Node{}, nil
	default:
		return nil, fmt.Errorf("unknown resource type")

	}
}

func JudgeResourceType(obj interface{}) (runtime.Object, string, error) {
	switch obj.(type) {
	case *apis.Node:
		return &apis.Node{}, "nodes", nil
	case *apis.Event:
		return &apis.Event{}, "events", nil
	case *apis.Data:
		return &apis.Data{}, "datas", nil
	case *apis.Workflow:
		return &apis.Workflow{}, "workflows", nil
	case *apis.Task:
		return &apis.Task{}, "tasks", nil
	case *apis.Group:
		return &apis.Group{}, "groups", nil
	case *apis.Action:
		return &apis.Action{}, "actions", nil
	case *apis.Scene:
		return &apis.Scene{}, "scenes", nil
	case *apis.Device:
		return &apis.Device{}, "devices", nil
	case *apis.Resource_NodeList:
		return &apis.Resource_NodeList{}, "resources", nil
	default:
		return nil, "", fmt.Errorf("unknown resource type")

	}

}

func EncodeResourceType(obj interface{}) (runtime.Object, string, error) {

	switch obj.(type) {
	case *apis.Node:
		return obj.(*apis.Node), "nodes", nil
	case *apis.Event:
		return obj.(*apis.Event), "events", nil
	case *apis.Data:
		return obj.(*apis.Data), "datas", nil
	case *apis.Workflow:
		return obj.(*apis.Workflow), "workflows", nil
	case *apis.Task:
		return obj.(*apis.Task), "tasks", nil
	case *apis.Group:
		return obj.(*apis.Group), "groups", nil
	case *apis.Action:
		return obj.(*apis.Action), "actions", nil
	case *apis.Scene:
		return obj.(*apis.Scene), "scenes", nil
	case *apis.Device:
		return obj.(*apis.Device), "devices", nil
	case *apis.Resource_Node:
		return obj.(*apis.Resource_Node), "resources", nil
	default:
		return nil, "", fmt.Errorf("unknown resource type")

	}

}

// 更新资源的函数，根据资源类型执行更新操作
func UpdateResource(clientSet *clients.ClientSet, obj interface{}, types string) (interface{}, error) {
	switch types {
	case "nodes":
		node := obj.(*apis.Node)
		namespace := node.Namespace
		nodeClient := clientSet.Core().Nodes(namespace)
		updatedNode, err := nodeClient.Update(context.TODO(), node, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("Failed to update Node: %v", err)
		}
		return updatedNode, nil
	case "tasks":
		task := obj.(*apis.Task)
		namespace := task.Namespace
		taskClient := clientSet.Core().Tasks(namespace)
		updatedTask, err := taskClient.Update(context.TODO(), task, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("Failed to update Task: %v", err)
		}
		return updatedTask, nil
	case "actions":
		action := obj.(*apis.Action)
		namespace := action.Namespace
		actionClient := clientSet.Core().Actions(namespace)
		updatedAction, err := actionClient.Update(context.TODO(), action, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("Failed to update Action: %v", err)
		}
		return updatedAction, nil
	case "datas":
		data := obj.(*apis.Data)
		namespace := data.Namespace
		dataClient := clientSet.Core().Datas(namespace)
		updatedData, err := dataClient.Update(context.TODO(), data, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("Failed to update Data: %v", err)
		}
		return updatedData, nil
	case "devices":
		device := obj.(*apis.Device)
		namespace := device.Namespace
		deviceClient := clientSet.Core().Devices(namespace)
		updatedDevice, err := deviceClient.Update(context.TODO(), device, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("Failed to update Device: %v", err)
		}
		return updatedDevice, nil
	case "events":
		event := obj.(*apis.Event)
		namespace := event.Namespace
		eventClient := clientSet.Core().Events(namespace)
		updatedEvent, err := eventClient.Update(context.TODO(), event, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("Failed to update Event: %v", err)
		}
		return updatedEvent, nil
	case "groups":
		group := obj.(*apis.Group)
		namespace := group.Namespace
		groupClient := clientSet.Core().Groups(namespace)
		updatedGroup, err := groupClient.Update(context.TODO(), group, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("Failed to update Group: %v", err)
		}
		return updatedGroup, nil
	case "resources":
		resourceNode := obj.(*apis.Resource_Node)
		namespace := resourceNode.Namespace
		resourceNodeClient := clientSet.Core().Resource_Nodes(namespace)
		updatedResourceNode, err := resourceNodeClient.Update(context.TODO(), resourceNode, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("Failed to update Resource Node: %v", err)
		}
		return updatedResourceNode, nil
	case "scenes":
		scene := obj.(*apis.Scene)
		namespace := scene.Namespace
		sceneClient := clientSet.Core().Scenes(namespace)
		updatedScene, err := sceneClient.Update(context.TODO(), scene, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("Failed to update Scene: %v", err)
		}
		return updatedScene, nil
	case "workflows":
		workflow := obj.(*apis.Workflow)
		namespace := workflow.Namespace
		workflowClient := clientSet.Core().Workflows(namespace)
		updatedWorkflow, err := workflowClient.Update(context.TODO(), workflow, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("Failed to update Workflow: %v", err)
		}
		return updatedWorkflow, nil

	default:
		return nil, fmt.Errorf("Unknown resource type: %s", types)
	}
}

func CreateResource(clientSet *clients.ClientSet, obj interface{}, types string) (interface{}, error) {

	switch types {
	case "nodes":
		node := obj.(*apis.Node)
		namespace := node.Namespace
		Client := clientSet.Core().Nodes(namespace)
		return Client.Create(context.TODO(), node, metav1.CreateOptions{})
	case "tasks":
		task := obj.(*apis.Task)
		namespace := task.Namespace
		Client := clientSet.Core().Tasks(namespace)
		return Client.Create(context.TODO(), task, metav1.CreateOptions{})
	case "actions":
		action := obj.(*apis.Action)
		namespace := action.Namespace
		client := clientSet.Core().Actions(namespace)
		return client.Create(context.TODO(), action, metav1.CreateOptions{})

	case "datas":
		data := obj.(*apis.Data)
		namespace := data.Namespace
		client := clientSet.Core().Datas(namespace)
		return client.Create(context.TODO(), data, metav1.CreateOptions{})

	case "devices":
		device := obj.(*apis.Device)
		namespace := device.Namespace
		client := clientSet.Core().Devices(namespace)
		return client.Create(context.TODO(), device, metav1.CreateOptions{})

	case "events":
		event := obj.(*apis.Event)
		namespace := event.Namespace
		client := clientSet.Core().Events(namespace)
		return client.Create(context.TODO(), event, metav1.CreateOptions{})

	case "groups":
		group := obj.(*apis.Group)
		namespace := group.Namespace
		client := clientSet.Core().Groups(namespace)
		return client.Create(context.TODO(), group, metav1.CreateOptions{})

	case "resources":
		resourceNode := obj.(*apis.Resource_Node)
		namespace := resourceNode.Namespace
		client := clientSet.Core().Resource_Nodes(namespace)
		return client.Create(context.TODO(), resourceNode, metav1.CreateOptions{})

	case "scenes":
		scene := obj.(*apis.Scene)
		namespace := scene.Namespace
		client := clientSet.Core().Scenes(namespace)
		return client.Create(context.TODO(), scene, metav1.CreateOptions{})

	case "workflows":
		workflow := obj.(*apis.Workflow)
		namespace := workflow.Namespace
		client := clientSet.Core().Workflows(namespace)
		return client.Create(context.TODO(), workflow, metav1.CreateOptions{})
	default:
		return nil, fmt.Errorf("unknown resource type %s", types)
	}
}

func DeleteResource(clientSet *clients.ClientSet, types, namespace, name string) error {
	switch types {
	case "nodes":
		Client := clientSet.Core().Nodes(namespace)
		return Client.Delete(context.TODO(), name, metav1.DeleteOptions{})

	case "tasks":
		Client := clientSet.Core().Tasks(namespace)
		return Client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	case "actions":
		client := clientSet.Core().Actions(namespace)
		return client.Delete(context.TODO(), name, metav1.DeleteOptions{})

	case "datas":
		client := clientSet.Core().Datas(namespace)
		return client.Delete(context.TODO(), name, metav1.DeleteOptions{})

	case "devices":
		client := clientSet.Core().Devices(namespace)
		return client.Delete(context.TODO(), name, metav1.DeleteOptions{})

	case "events":
		client := clientSet.Core().Events(namespace)
		return client.Delete(context.TODO(), name, metav1.DeleteOptions{})

	case "groups":
		client := clientSet.Core().Groups(namespace)
		return client.Delete(context.TODO(), name, metav1.DeleteOptions{})

	case "resources":
		client := clientSet.Core().Resource_Nodes(namespace)
		return client.Delete(context.TODO(), name, metav1.DeleteOptions{})

	case "scenes":
		client := clientSet.Core().Scenes(namespace)
		return client.Delete(context.TODO(), name, metav1.DeleteOptions{})

	case "workflows":
		client := clientSet.Core().Workflows(namespace)
		return client.Delete(context.TODO(), name, metav1.DeleteOptions{})
	default:
		return fmt.Errorf("unknown resource type: %s", types)
	}
}

func GetResource(clientSet *clients.ClientSet, types, namespace, name string) (interface{}, error) {
	switch types {
	case "nodes":
		Client := clientSet.Core().Nodes(namespace)
		return Client.Get(context.TODO(), name, metav1.GetOptions{})

	case "tasks":
		Client := clientSet.Core().Tasks(namespace)
		return Client.Get(context.TODO(), name, metav1.GetOptions{})
	case "actions":
		client := clientSet.Core().Actions(namespace)
		return client.Get(context.TODO(), name, metav1.GetOptions{})

	case "datas":
		client := clientSet.Core().Datas(namespace)
		return client.Get(context.TODO(), name, metav1.GetOptions{})

	case "devices":
		client := clientSet.Core().Devices(namespace)
		return client.Get(context.TODO(), name, metav1.GetOptions{})

	case "events":
		client := clientSet.Core().Events(namespace)
		return client.Get(context.TODO(), name, metav1.GetOptions{})

	case "groups":
		client := clientSet.Core().Groups(namespace)
		return client.Get(context.TODO(), name, metav1.GetOptions{})

	case "resources":
		client := clientSet.Core().Resource_Nodes(namespace)
		return client.Get(context.TODO(), name, metav1.GetOptions{})

	case "scenes":
		client := clientSet.Core().Scenes(namespace)
		return client.Get(context.TODO(), name, metav1.GetOptions{})

	case "workflows":
		client := clientSet.Core().Workflows(namespace)
		return client.Get(context.TODO(), name, metav1.GetOptions{})
	default:
		return nil, fmt.Errorf("unknown resource type: %s", types)
	}
}
