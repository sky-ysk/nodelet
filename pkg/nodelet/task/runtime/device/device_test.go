package device

import (
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"net/http"
	"testing"
	"time"
)

func InitClient() (*clients.ClientSet, error) {
	//初始化ClientSet客户端
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	c := &rest.Config{
		Host:    "http://localhost:10000",
		APIPath: "/apis/resources/v1",
		ContentConfig: rest.ContentConfig{
			AcceptContentTypes: "application/json; charset=UTF-8", //text/plain; charset=UTF-8
			ContentType:        "application/json; charset=UTF-8", //application/json; charset=UTF-8
			GroupVersion: &schema.GroupVersion{
				Group:   "resources",
				Version: "v1",
			},
			NegotiatedSerializer: serializer.NewCodecFactory(scheme),
		},
		UserAgent: "defaultUserAgent",
		Transport: &http.Transport{
			MaxIdleConns:        100,              // 最大空闲连接数
			IdleConnTimeout:     90 * time.Second, // 空闲连接超时时间
			TLSHandshakeTimeout: 10 * time.Second, // TLS 握手超时时间
		},
		Timeout: 10 * time.Second,
	}
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize clientSet: %v", err)
	}
	return clientSet, nil
}

func TestRun(t *testing.T) {
	moduleName := "testModule"
	logs.Init(moduleName)
	logs.Infof("[test] testing run.....\n")
	clientSet, err := InitClient()
	if err != nil {
		logs.Errorf("[test] init clientSet failed: %v", err)
	}
	deviceClient := clientSet.Core().Devices("")
	dr := NewDeviceRuntime(deviceClient)
	action, runtimeForTest := NewActionAndRuntime()
	err = dr.Run(&apis.Group{}, action, runtimeForTest, 0, 0)
	if err != nil {
		fmt.Println(err)
	}
}

func NewActionAndRuntime() (*apis.Action, *apis.Runtime) {
	action := apis.Action{
		Spec: apis.ActionSpec{
			Name: "ActionTest",
		},
		Status: apis.ActionStatus{
			ActionID: "Action1",
			Resources: []apis.ResourceStatus{
				apis.ResourceStatus{
					Name:         "cpu",
					Reserved:     10,
					ReservedUnit: apis.ComputeCPU,
				},
				apis.ResourceStatus{
					Name:         "memory",
					Reserved:     4096,
					ReservedUnit: apis.StorageMB,
				},
				apis.ResourceStatus{
					Name:         "disk",
					Reserved:     200,
					ReservedUnit: apis.StorageGB,
				},
			},

			Devices: []apis.DeviceStatus{
				apis.DeviceStatus{
					Lock: apis.Lock{
						IsLocked: true,
					},
					Status:     "idle",
					Phase:      apis.DeviceIdle,
					InstanceID: "",
					ActionID:   "",
					DeviceID:   "transferRobot",
				},
			},

			//Scenes: map[string]apis.SceneStatus{
			//	"o1": apis.SceneStatus{
			//		AttachedTask:   "",
			//		AttachedDevice: "",
			//		AttachedScene:  "",
			//		Lock: apis.Lock{
			//			IsLocked: true,
			//		},
			//	},
			//	"p1": apis.SceneStatus{
			//		AttachedTask:   "",
			//		AttachedDevice: "",
			//		AttachedScene:  "",
			//		Lock: apis.Lock{
			//			IsLocked: true,
			//		},
			//	},
			//},
		},
	}

	runtime := apis.Runtime{
		Image: "Move",
		Name:  "RuntimeTest",
		Devices: []apis.DeviceSpec{
			apis.DeviceSpec{
				Name:               "transferRobot",
				ExpectedProperties: map[string]apis.Property{},
				AccessMethod: apis.AccessMethod{
					Type:  apis.AccessByRmf,
					URL:   "http://192.168.1.225:8000",
					Group: "tinyRobot",
					Alias: "transferRobot",
				},
				Desc: apis.DeviceDesc{
					Label: []string{"Move"},
				},
			},
		},
		Outputs: make([]apis.Output, 0),
		Inputs: []apis.Input{
			apis.Input{
				Type:      apis.LocalData,
				Name:      "dest",
				Value:     "R201",
				ValueType: "string",
			},
			apis.Input{
				Type:      apis.LocalData,
				Name:      "orientation",
				Value:     "-3.12",
				ValueType: "double",
			},
			apis.Input{
				Type:      apis.LocalData,
				Name:      "dock",
				Value:     "true",
				ValueType: "bool",
			},
		},
		//Resources: []apis.ResourceSpec{
		//	apis.ResourceSpec{
		//		Name:              "cpu",
		//		ExpectedValue:     2,
		//		ExpectedValueUnit: apis.ComputeCPU,
		//		Type:              apis.Compute,
		//	},
		//	apis.ResourceSpec{
		//		Name:              "memory",
		//		ExpectedValue:     1024,
		//		ExpectedValueUnit: apis.StorageMB,
		//		Type:              apis.Storage,
		//	},
		//	apis.ResourceSpec{
		//		Name:              "disk",
		//		ExpectedValue:     10,
		//		ExpectedValueUnit: apis.StorageGB,
		//		Type:              apis.Storage,
		//	},
		//},
		//Scenes: []apis.SceneSpec{
		//	apis.SceneSpec{
		//		SceneID:          "o1",
		//		Type:             apis.ObjectType,
		//		ExpectedProperty: map[string]apis.Property{"position": apis.Property{Name: "position", Type: apis.StringType, Value: "R201"}},
		//	},
		//	apis.SceneSpec{
		//		SceneID:          "p1",
		//		Type:             apis.PositionType,
		//		ExpectedProperty: map[string]apis.Property{"isOccupied": apis.Property{Name: "isOccupied", Type: apis.BoolType, Value: "true"}},
		//	},
		//},
	}
	action.Spec.Runtimes = []apis.Runtime{runtime}
	return &action, &runtime
}
