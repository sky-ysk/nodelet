package utils

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	apis "hit.edu/framework/pkg/apis/cores"
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

func (engine *ConditionEngine) checkFormula(formula apis.ConditionFormula) (apis.ResultType, error) {

	switch formula.Type {
	case apis.NodeDependency:

	case apis.DataDependency:

	case apis.ResourceDependency:

	case apis.ProgramDependency:

	}

	leftReady, leftVal := engine.extractValue(formula.RightValue)
	rightReady, rightVal := engine.extractValue(formula.LeftValue)
	if !leftReady || !rightReady {
		return apis.NotReady, nil
	}
	if formula.Signal == apis.Equal {
		if leftVal != rightVal {
			return apis.False, nil
		}
		return apis.True, nil
	} else if formula.Signal != apis.Equal {
		if leftVal == rightVal {
			return apis.False, nil
		}
		return apis.True, nil
	}
	logs.Error("unsupported signal type ", formula.Signal)
	return apis.False, errors.New(string("unsupported signal type " + formula.Signal))
}

// TODO 解析具体的值，返回bool表示值是否就绪，string表示值
func (engine *ConditionEngine) extractValue(value apis.ConditionValue) (bool, string) {

	switch value.ValueType {
	case apis.ConstantType:
		return true, value.Value

	//Task{task1}.Group{group1}.Action{action1}.Output{completed}
	//
	//Succeed Group状态
	//目前做一些特殊逻辑，只去捞Action里面的东西
	case apis.ArgumentRefType:
		//re := regexp.MustCompile(`Action\{([^}]+)}`)

		// 查找子匹配
		//match := re.FindStringSubmatch(value.Value)
		//if len(match) < 2 {
		//	logs.Error("no action found ", value.Value)
		//	return false, "" // 没有找到匹配项
		//}
		////
		//actionName := match[1]
		////TODO 从client里面拿结果 校验
	default:
		logs.Fatal("unsupported value type ", value.ValueType)
		return false, ""
	}
	return false, ""
}

//From用于确定来源的对象，例如某个action的statu或者spec
//From格式：Task{Name}.Group{Name}.Action{Name}.Runtime{Name}.Status/Spec
func ParseFrom(input string) (TaskID, GroupID, ActionID, RuntimeID, Field string, ParseTypeID int, err error) {
	// 定义部分名称的顺序
	partNames := []string{"Task", "Group", "Action", "Runtime"}

	// 初始化结果
	result := make(map[string]string)
	for _, name := range partNames {
		result[name] = ""
	}
	result["Field"] = ""

	// 将输入字符串按 '.' 分割
	parts := strings.Split(input, ".")

	

	// 依次处理每个部分
	for _, part := range parts {
		if part == "" {
			continue
		}
		// 检查是否是 Field
		if part == "Field" {
			result["Field"] = part
			continue
		}
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

	// 提取结果
	TaskID = result["Task"]
	GroupID = result["Group"]
	ActionID = result["Action"]
	RuntimeID = result["Runtime"]
	Field = result["Field"]

	ParseTypeID = 0
	// 如果某部分没有匹配到，则将ParseTypeID加上对应值，判断是那一种情况
	//Task{ID}.Group{ID}.Action{ID}.Runtime{ID}
	//     0000 ~ 1111，某些情况去除
	if TaskID == "" {
		ParseTypeID += 1 << 3
	}
	if GroupID == "" {
		ParseTypeID += 1 << 2
	}
	if ActionID == "" {
		ParseTypeID += 1 << 1
	}
	if RuntimeID == "" {
		ParseTypeID += 1
	}
	//ID就是Name
	return TaskID, GroupID, ActionID, RuntimeID, Field, ParseTypeID, err
}

//Field用于解析From获取的item的字段路径
//Field格式：
// 例如获取某个Action的完成情况:  Phase{}
// 例如获取某个action的device的abilities下面的AbilityServiceStatus下面的port字段:  Device{deviceCamera}.Abilities{0}.AbilityServiceStatus{0}.port{}。
// {}里面填对应寻址方式的index，如果是列表就填下标，如果是map就填key。如果直接是某个变量，置为空。
func ParseField(Field interface{}, path string) (interface{}, string, error) {
	//两类数据类型对应两类正则解析式
	// 定义正则表达式
	//遇到一个问题，这里的每一项的类型都可能不一样，比如使用列表、结构体、map，这三种类型的话，{}里面应该怎么填比较合适？通过reflect应该已经解决了，等待测试
	//对于列表，里面填下标可能是不合适的，例如目前的RuntimeStatus
	//格式为 Status/Spec.RuntimeStatus{}
	// 按照 '.' 分割路径
	parts := strings.Split(path, ".")

	// 逐步解析路径
	current := reflect.ValueOf(Field)
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
