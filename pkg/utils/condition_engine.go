package utils

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"time"

	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/utils/value"
	"hit.edu/framework/pkg/client-go/util/manager"
)

// switch kind {
// case "Runtime":
// 	r := (o).(apis.Runtime)
// 	namespace = r.Namespace
// 	name, kindType, from, fromKey, err = e.GetNameFromRuntime(typeName, parts, &r)
// 	if err != nil {
// 		return nil, err
// 	}
// case "Action":
// 	r := (o).(apis.Action)
// 	namespace = r.Namespace
// 	name, kindType, from, fromKey, err = e.GetNameFromAction(typeName, parts, &r)
// 	if err != nil {
// 		return nil, err
// 	}
// case "Group":
// 	r := (o).(apis.Group)
// 	namespace = r.Namespace
// 	name, kindType, from, fromKey, err = e.GetNameFromGroup(typeName, parts, &r)
// 	if err != nil {
// 		return nil, err
// 	}
// case "Task":
// 	r := (o).(apis.Task)
// 	namespace = r.Namespace
// 	name, kindType, from, fromKey, err = e.GetNameFromTask(typeName, parts, &r)
// 	if err != nil {
// 		return nil, err
// 	}
// case "Workflow":
// 	r := (o).(apis.Workflow)
// 	namespace = r.Namespace
// 	name, kindType, from, fromKey, err = e.GetNameFromWorkflow(typeName, parts, &r)
// 	if err != nil {
// 		return nil, err
// 	}
// }

// 为后续做成有状态的类留出扩展
type ConditionEngine struct {
	engine *value.Engine // 添加 Engine 字段
	manager  *manager.Manager
}

func InitClient() (*clients.ClientSet, error) {
	//初始化ClientSet客户端
	scheme := runtime.NewScheme()
	apis.AddToScheme(scheme)
	c := &rest.Config{
		Host:    GetAPIServerHost(), //http://localhost:10000   http://suda801.wangwanu.com:11006   //连接api-server
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
		Timeout: 3600 * time.Second,
	}
	clientSet, err := clients.NewForConfig(c)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize clientSet: %v", err)
	}
	return clientSet, nil
}
func GetAPIServerHost() string {
	if host := os.Getenv("API_SERVER_HOST"); host != "" {
		return host
	}
	return "http://localhost:10000"
}
func NewConditionEngine() *ConditionEngine {
	client, err := InitClient()
	if err != nil {
		logs.Errorf("Failed to initialize client: %v", err)
	}
	return &ConditionEngine{
		engine: value.NewEngine(client), // 初始化 Engine
		manager: manager.NewManager(client),	//TODO:有一个问题，这里使用相同的client会不会有问题
	}
}

// o传入的是一个对象，可能是workflow、task、group、action、runtime等
func (engine *ConditionEngine) CheckConditions(conditions apis.Conditions, o interface{}) (apis.ResultType, error) {

	if len(conditions.Formulas) == 0 {
		return apis.True, nil
	}

	for _, formula := range conditions.Formulas {
		checkRes, err := engine.checkFormula(formula, o)
		if err != nil {
			return apis.False, err
		}
		if checkRes != apis.True {
			return checkRes, nil
		}
	}
	return apis.True, nil
}

func (engine *ConditionEngine) checkFormula(formula apis.ConditionFormula, o interface{}) (apis.ResultType, error) {
	switch formula.ConditionType {
	case apis.NodeDependency:
		res, err := engine.checkNodeDependency(formula, o)
		if err != nil {
			return apis.False, err
		}
		if res != apis.True {
			return res, nil
		}
		return apis.True, nil
	case apis.DataDependency:
		res, err := engine.checkDataDependency(formula, o)
		if err != nil {
			return apis.False, err
		}
		if res != apis.True {
			return res, nil
		}
		return apis.True, nil
	case apis.ResourceDependency:
		res, err := engine.checkResourceDependency(formula, o)
		if err != nil {
			return apis.False, err
		}
		if res != apis.True {
			return res, nil
		}
		return apis.True, nil
	case apis.ProgramDependency:
		res, err := engine.checkProgramDependency(formula, o)
		if err != nil {
			return apis.False, err
		}
		if res != apis.True {
			return res, nil
		}
		return apis.True, nil
	default:
		logs.Error("unsupported condition type ", formula.ConditionType)
		return apis.False, errors.New(string("unsupported signal type " + formula.Signal))
	}
}

// rightReady, rightVal := engine.extractValue(formula.LeftValue)
// leftReady, leftVal := engine.extractValue(formula.RightValue)
// if !leftReady || !rightReady {
// 	return apis.NotReady, nil
// }
// if formula.Signal == apis.Equal {
// 	if leftVal != rightVal {
// 		return apis.False, nil
// 	}
// 	return apis.True, nil
// } else if formula.Signal == apis.Equal {
// 	if leftVal == rightVal {
// 		return apis.False, nil
// 	}
// 	return apis.True, nil
// }
// logs.Error("unsupported signal type ", formula.Signal)

// TODO:细化每一种依赖里面的每一种情况
func (ce *ConditionEngine) checkNodeDependency(formula apis.ConditionFormula, o interface{}) (apis.ResultType, error) {
	kind := reflect.TypeOf(o).Name()
	switch kind {
	case "Runtime":
		r := (o).(apis.Runtime)
		logs.Info(r)
	case "Action":
		r := (o).(apis.Action)
		logs.Info(r)
	case "Group":
		r := (o).(apis.Group)
		logs.Info(r)
	case "Task":
		r := (o).(apis.Task)
		logs.Info(r)
	case "Workflow":
		r := (o).(apis.Workflow)
		logs.Info(r)
	}
	
	return apis.True, nil
}

func (ce *ConditionEngine) checkDataDependency(formula apis.ConditionFormula, o interface{}) (apis.ResultType, error) {

	return apis.True, nil
}

func (ce *ConditionEngine) checkResourceDependency(formula apis.ConditionFormula, o interface{}) (apis.ResultType, error) {

	return apis.True, nil
}

func (ce *ConditionEngine) checkProgramDependency(formula apis.ConditionFormula, o interface{}) (apis.ResultType, error) {

	return apis.True, nil
}

// TODO 解析具体的值，返回bool表示值是否就绪，string表示值
func (ce *ConditionEngine) extractValue(value apis.Value, o interface{}) (bool, string) {

	switch value.Type {
	case apis.ConstData:
		return true, value.Value

	case apis.LocalData:
		return true, value.Value
		////TODO 从client里面拿结果 校验
	case apis.DeviceData:
		val, err := ce.engine.GetValue(&value, o)
		if err != nil {
			logs.Error("extract value error: ", err)
			return false, "0"
		}
		if val.Value == "" {
			return false, "0"
		}
		//检查数据是否存在，存在返回“1”即可，与rightVal的“1”进行比较
		return true, "1"

	default:
		logs.Fatal("unsupported value type ", value.ValueType)
		return false, ""
	}
}

// //后续可能会用到，目前没用
// // 防止有这样的需求：依然是根据From、Field来选择。目前是根据每一种情况的正则表达式来访问数据
// // GetItem返回Task/Group/Action/Runtime这四个里面的其中一个结构体本身
// func (ce *ConditionEngine) GetItem(FromInput string) (interface{}, error) {

// 	FromItemInfo, err := ParseFrom(FromInput)
// 	if err != nil {
// 		logs.Error("Get item err, Parse From Failed")
// 		return nil, errors.New("")
// 	}

// 	//TODO 本地的情况
// 	//Local还需要设计，目前全部按照etcd获取
// 	if FromItemInfo.IsLocal {
// 		logs.Info("Local Type is not supported now.")
// 		return nil, errors.New("local Type is not supported now")
// 	}

// 	var currentItem interface{}
// 	//有TaskName，直接Get
// 	if FromItemInfo.TaskName != "" {
// 		task, err := ce.manager.GetTask(FromItemInfo.TaskName)

// 		if err != nil {
// 			logs.Error("Get task by taskName error from etcd:%v", err)
// 			return nil, errors.New("condition Get Task Item error")
// 		}
// 		currentItem = *task
// 	}
// 	//有GroupName，看看是否有父亲Task
// 	if FromItemInfo.GroupName != "" {
// 		if currentItem == nil {
// 			group, err := eg.groupClient.Get(context.TODO(), FromItemInfo.GroupName, metav1.GetOptions{})
// 			if err != nil {
// 				logs.Error("Get group by GroupName error from etcd:%v", err)
// 				return nil, errors.New("condition Get Group Item error")
// 			}
// 			currentItem = *group
// 		} else {
// 			// 类型断言
// 			if task, ok := currentItem.(apis.Task); ok {
// 				for _, group := range task.Spec.Groups {
// 					if group.Spec.Name == FromItemInfo.GroupName {
// 						currentItem = group
// 					}
// 				}
// 			} else {
// 				logs.Error("can not get group before no parent task!")
// 			}
// 		}
// 	}
// 	//有ActionName，查看是否有父亲Group
// 	if FromItemInfo.ActionName != "" {
// 		if currentItem == nil {
// 			action, err := eg.actionClient.Get(context.TODO(), FromItemInfo.ActionName, metav1.GetOptions{})
// 			if err != nil {
// 				logs.Error("Get action by ActionName error from etcd:%v", err)
// 				return nil, errors.New("condition Get Action Item error")
// 			}
// 			currentItem = *action
// 		} else {
// 			// 类型断言
// 			if group, ok := currentItem.(apis.Group); ok {
// 				for _, action := range group.Spec.Actions {
// 					if action.Spec.Name == FromItemInfo.ActionName {
// 						currentItem = action
// 					}
// 				}
// 			} else {
// 				logs.Error("can not get runtime before no parent action!")
// 			}
// 		}
// 	}
// 	//有Runtime，从Action去查找：
// 	if FromItemInfo.RuntimeName != "" {
// 		// 类型断言
// 		if action, ok := currentItem.(apis.Action); ok {
// 			for _, runtime := range action.Spec.Runtimes {
// 				if runtime.Name == FromItemInfo.RuntimeName {
// 					currentItem = runtime
// 				}
// 			}
// 		} else {
// 			logs.Error("can not get runtime before no parent action!")
// 		}
// 	}
// 	//debug
// 	logs.Info("get currentItem:%v", currentItem)

// 	return currentItem, nil
// }

// // From用于确定来源的对象，例如某个action的statu或者spec
// // From格式：Task{Name}.Group{Name}.Action{Name}.Runtime{Name}
// // TODO 返回结构体形式
// func ParseFrom(input string) (apis.FromItemInfo, error) {
// 	// 定义部分名称的顺序
// 	partNames := []string{"Task", "Group", "Action", "Runtime"}
// 	result := make(map[string]string)
// 	for _, name := range partNames {
// 		result[name] = ""
// 	}
// 	// result["Field"] = ""

// 	// 将输入字符串按 '.' 分割
// 	parts := strings.Split(input, ".")
// 	for _, part := range parts {
// 		if part == "" {
// 			continue
// 		}
// 		//Status/Spec的检查移到Field里了
// 		// // 检查是否是 Field
// 		// if part == "Field" {
// 		//  result["Field"] = part
// 		//  continue
// 		// }

// 		// 分割 Part{ID}
// 		partSplit := strings.SplitN(part, "{", 2)
// 		if len(partSplit) != 2 {
// 			continue
// 		}
// 		partName := partSplit[0]
// 		partID := strings.TrimSuffix(partSplit[1], "}")
// 		if _, exists := result[partName]; exists {
// 			result[partName] = partID
// 		}
// 	}
// 	// var FieldType apis.FieldType
// 	// if result["Field"] == "Status" {
// 	//  FieldType = apis.StatusType
// 	// } else if result["Field"] == "Spec" {
// 	//  FieldType = apis.Spectype
// 	// }
// 	FromItemInfo := apis.FromItemInfo{
// 		TaskName:    result["Task"],
// 		GroupName:   result["Group"],
// 		ActionName:  result["Action"],
// 		RuntimeName: result["Runtime"],
// 	}
// 	if FromItemInfo.TaskName != "" || FromItemInfo.GroupName != "" {
// 		FromItemInfo.IsLocal = false
// 	} else {
// 		FromItemInfo.IsLocal = true
// 	}
// 	if FromItemInfo.TaskName == "" && FromItemInfo.GroupName == "" && FromItemInfo.ActionName == "" && FromItemInfo.RuntimeName == "" {
// 		return FromItemInfo, errors.New("parse From err! Format err")
// 	}
// 	// DEBUG打印解析结果
// 	// logs.Trace("From Item:%v", FromItemInfo)

// 	return FromItemInfo, nil
// }

// // Field用于解析From获取的item的字段路径
// // Field格式：(Status和Spec在这里最前面列出来)
// // 例如获取某个Action的完成情况:  Status{}.Phase{}
// // 例如获取某个action的device的abilities下面的AbilityServiceStatus下面的port字段:  Spec{}.Device{deviceCamera}.Abilities{0}.AbilityServiceStatus{0}.port{}。
// // {}里面填对应寻址方式的index，如果是列表就填下标，如果是map就填key。如果直接是某个变量，置为空。
// func ParseField(Item interface{}, Field string) (interface{}, string, error) {
// 	// 按照 '.' 分割路径
// 	parts := strings.Split(Field, ".")
// 	// 逐步解析路径
// 	current := reflect.ValueOf(Item)
// 	for _, part := range parts {
// 		// 解析字段名和索引
// 		field := strings.Split(part, "{")
// 		fieldName := field[0]
// 		var index string
// 		if len(field) > 1 {
// 			index = strings.TrimSuffix(field[1], "}")
// 		} else {
// 			index = ""
// 		}
// 		// 获取字段值
// 		fieldValue := current.FieldByName(fieldName)
// 		if !fieldValue.IsValid() {
// 			return nil, "", errors.New("")
// 		}
// 		// 如果有索引，处理索引
// 		if index != "" {
// 			if fieldValue.Kind() == reflect.Map {
// 				mapKey := reflect.ValueOf(index)
// 				fieldValue = fieldValue.MapIndex(mapKey)
// 			} else if fieldValue.Kind() == reflect.Slice {
// 				indexInt, err := strconv.Atoi(index)
// 				if err != nil || indexInt < 0 || indexInt >= fieldValue.Len() {
// 					return nil, "", errors.New("")
// 				}
// 				fieldValue = fieldValue.Index(indexInt)
// 			}
// 		}
// 		// 更新 current 为当前字段值
// 		current = fieldValue
// 	}
// 	// 返回最终的值
// 	return current.Interface(), current.Type().String(), nil
// }
