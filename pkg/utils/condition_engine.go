package utils

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients/typed/core"
	"hit.edu/framework/pkg/component-base/logs"
)

// 为后续做成有状态的类留出扩展
type ConditionEngine struct {
	nodeClient   core.NodeInterface //需要查node信息
	groupClient  core.GroupInterface
	taskClient   core.TaskInterface
	actionClient core.ActionInterface
	stopCh       chan struct{}
}

func NewConditionEngine(nodeClient core.NodeInterface, taskClient core.TaskInterface, groupClient core.GroupInterface, actionClient core.ActionInterface) *ConditionEngine {
	return &ConditionEngine{
		nodeClient:   nodeClient,
		taskClient:   taskClient,
		groupClient:  groupClient,
		actionClient: actionClient,
		stopCh:       make(chan struct{}),
	}
}

// TODO：condition engine不断检查本地的所有Task Group Action Runtime的 condition
//

func (engine *ConditionEngine) CheckConditions(conditions apis.Conditions) (apis.ResultType, error) {

	if len(conditions.Formulas) == 0 {
		return apis.True, nil
	}

	for _, formula := range conditions.Formulas {
		checkRes, err := engine.checkFormula(formula)
		if err != nil {
			return apis.False, err
		}
		if checkRes != apis.True {
			return checkRes, nil
		}
	}
	return apis.True, nil
}

// TODO：规则检查，根据不同类型的condition进行不同的操作
// NodeDependency：检查parent节点状态是否是完成
// DataDependency：检查data是否已经下载完成（数据的下载时机？部署的时候开始下载？这里只负责检查）
// ResourceDependency：检查资源是否能够满足（内存 CPU占用率等）
// ProgramDependency：检查程序依赖是否满足（python包等）
func (engine *ConditionEngine) checkFormula(formula apis.ConditionFormula) (apis.ResultType, error) {
	switch formula.Type {
	case apis.NodeDependency:

	case apis.DataDependency:

		leftReady, _, _ := engine.extractValue(formula.LeftValue)
		rightReady, rightVal, _ := engine.extractValue(formula.RightValue)
		if !leftReady || !rightReady {
			return apis.NotReady, nil
		}
	
		if rightVal != "" {
			return apis.True, nil
		} else {
			return apis.False, errors.New("do not get leftVal")
		}

		// if leftType != rightType {
		// 	logs.Error("checkFormula err, left and right type are not the same")
		// 	return apis.False, errors.New("checkFormula err, left and right type are not the same")
		// }

	case apis.ResourceDependency:

	case apis.ProgramDependency:

	}

	logs.Error("unsupported signal type ", formula.Signal)
	return apis.False, errors.New(string("unsupported signal type " + formula.Signal))
}

// TODO 解析具体的值，返回bool表示值是否就绪，string表示值
func (engine *ConditionEngine) extractValue(value apis.ConditionValue) (bool, string, string) {

	switch value.ValueType {
	case apis.ConstantType:
		return true, value.Value, ""

	//Task{task1}.Group{group1}.Action{action1}.Output{completed}
	//
	//Succeed Group状态
	//目前做一些特殊逻辑，只去捞Action里面的东西
	case apis.ArgumentRefType:
		val, typ, err := engine.GetValue(value.From, value.Field)
		if err != nil {
			logs.Error("extractValue err, bacause GetValue err")
			return false, "", ""
		}
		str := fmt.Sprintf("%v", val)
		return true, str, typ
	default:
		logs.Fatal("unsupported value type ", value.ValueType)
		return false, "", ""
	}
}

func (eg *ConditionEngine) GetValue(From, Field string) (interface{}, string, error) {
	Item, err := eg.GetItem(From)
	if err != nil {
		logs.Error("when getting value, Get Item err")
	}
	Value, ValueType, err := ParseField(Item, Field)
	if err != nil {
		logs.Error("when getting value, Parse Field err")
	}
	return Value, ValueType, nil
}

// GetItem返回Task/Group/Action/Runtime这四个里面的其中一个结构体本身
func (eg *ConditionEngine) GetItem(FromInput string) (interface{}, error) {

	FromItemInfo, err := ParseFrom(FromInput)
	if err != nil {
		logs.Error("Get item err, Parse From Failed")
		return nil, errors.New("")
	}

	//TODO 本地的情况
	//Local还需要设计，目前全部按照etcd获取
	if FromItemInfo.IsLocal {
		logs.Info("Local Type is not supported now.")
		return nil, errors.New("local Type is not supported now")
	}

	var currentItem interface{}
	//有TaskName，直接Get
	if FromItemInfo.TaskName != "" {
		task, err := eg.taskClient.Get(context.TODO(), FromItemInfo.TaskName, metav1.GetOptions{})
		if err != nil {
			logs.Error("Get task by taskName error from etcd:%v", err)
			return nil, errors.New("condition Get Task Item error")
		}
		currentItem = *task
	}
	//有GroupName，看看是否有父亲Task
	if FromItemInfo.GroupName != "" {
		if currentItem == nil {
			group, err := eg.groupClient.Get(context.TODO(), FromItemInfo.GroupName, metav1.GetOptions{})
			if err != nil {
				logs.Error("Get group by GroupName error from etcd:%v", err)
				return nil, errors.New("condition Get Group Item error")
			}
			currentItem = *group
		} else {
			// 类型断言
			if task, ok := currentItem.(apis.Task); ok {
				for _, group := range task.Spec.Groups {
					if group.Spec.Name == FromItemInfo.GroupName {
						currentItem = group
					}
				}
			} else {
				logs.Error("can not get group before no parent task!")
			}
		}
	}
	//有ActionName，查看是否有父亲Group
	if FromItemInfo.ActionName != "" {
		if currentItem == nil {
			action, err := eg.actionClient.Get(context.TODO(), FromItemInfo.ActionName, metav1.GetOptions{})
			if err != nil {
				logs.Error("Get action by ActionName error from etcd:%v", err)
				return nil, errors.New("condition Get Action Item error")
			}
			currentItem = *action
		} else {
			// 类型断言
			if group, ok := currentItem.(apis.Group); ok {
				for _, action := range group.Spec.Actions {
					if action.Spec.Name == FromItemInfo.ActionName {
						currentItem = action
					}
				}
			} else {
				logs.Error("can not get runtime before no parent action!")
			}
		}
	}
	//有Runtime，从Action去查找：
	if FromItemInfo.RuntimeName != "" {
		// 类型断言
		if action, ok := currentItem.(apis.Action); ok {
			for _, runtime := range action.Spec.Runtimes {
				if runtime.Name == FromItemInfo.RuntimeName {
					currentItem = runtime
				}
			}
		} else {
			logs.Error("can not get runtime before no parent action!")
		}
	}
	return currentItem, nil
}

// From用于确定来源的对象，例如某个action的statu或者spec
// From格式：Task{Name}.Group{Name}.Action{Name}.Runtime{Name}
// TODO 返回结构体形式
func ParseFrom(input string) (apis.FromItemInfo, error) {
	// 定义部分名称的顺序
	partNames := []string{"Task", "Group", "Action", "Runtime"}
	result := make(map[string]string)
	for _, name := range partNames {
		result[name] = ""
	}
	// result["Field"] = ""

	// 将输入字符串按 '.' 分割
	parts := strings.Split(input, ".")
	for _, part := range parts {
		if part == "" {
			continue
		}
		//Status/Spec的检查移到Field里了
		// // 检查是否是 Field
		// if part == "Field" {
		// 	result["Field"] = part
		// 	continue
		// }

		// 分割 Part{ID}
		partSplit := strings.SplitN(part, "{", 2)
		if len(partSplit) != 2 {
			continue
		}
		partName := partSplit[0]
		partID := strings.TrimSuffix(partSplit[1], "}")
		if _, exists := result[partName]; exists {
			result[partName] = partID
		}
	}
	// var FieldType apis.FieldType
	// if result["Field"] == "Status" {
	// 	FieldType = apis.StatusType
	// } else if result["Field"] == "Spec" {
	// 	FieldType = apis.Spectype
	// }
	FromItemInfo := apis.FromItemInfo{
		TaskName:    result["Task"],
		GroupName:   result["Group"],
		ActionName:  result["Action"],
		RuntimeName: result["Runtime"],
	}
	if FromItemInfo.TaskName != "" || FromItemInfo.GroupName != "" {
		FromItemInfo.IsLocal = false
	} else {
		FromItemInfo.IsLocal = true
	}
	if FromItemInfo.TaskName == "" && FromItemInfo.GroupName == "" && FromItemInfo.ActionName == "" && FromItemInfo.RuntimeName == "" {
		return FromItemInfo, errors.New("parse From err! Format err")
	}
	// DEBUG打印解析结果
	// logs.Trace("From Item:%v", FromItemInfo)

	return FromItemInfo, nil
}

// Field用于解析From获取的item的字段路径
// Field格式：(Status和Spec在这里最前面列出来)
// 例如获取某个Action的完成情况:  Status{}.Phase{}
// 例如获取某个action的device的abilities下面的AbilityServiceStatus下面的port字段:  Spec{}.Device{deviceCamera}.Abilities{0}.AbilityServiceStatus{0}.port{}。
// {}里面填对应寻址方式的index，如果是列表就填下标，如果是map就填key。如果直接是某个变量，置为空。
func ParseField(Item interface{}, Field string) (interface{}, string, error) {
	// 按照 '.' 分割路径
	parts := strings.Split(Field, ".")
	// 逐步解析路径
	current := reflect.ValueOf(Item)
	for _, part := range parts {
		// 解析字段名和索引
		field := strings.Split(part, "{")
		fieldName := field[0]
		var index string
		if len(field) > 1 {
			index = strings.TrimSuffix(field[1], "}")
		} else {
			index = ""
		}
		// 获取字段值
		fieldValue := current.FieldByName(fieldName)
		if !fieldValue.IsValid() {
			return nil, "", errors.New("")
		}
		// 如果有索引，处理索引
		if index != "" {
			if fieldValue.Kind() == reflect.Map {
				mapKey := reflect.ValueOf(index)
				fieldValue = fieldValue.MapIndex(mapKey)
			} else if fieldValue.Kind() == reflect.Slice {
				indexInt, err := strconv.Atoi(index)
				if err != nil || indexInt < 0 || indexInt >= fieldValue.Len() {
					return nil, "", errors.New("")
				}
				fieldValue = fieldValue.Index(indexInt)
			}
		}
		// 更新 current 为当前字段值
		current = fieldValue
	}
	// 返回最终的值
	return current.Interface(), current.Type().String(), nil
}
