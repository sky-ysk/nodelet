package utils

import (
	"errors"
	"os"
	"reflect"
	"strings"

	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/utils/value"
)

// 为后续做成有状态的类留出扩展
type ConditionEngine struct {
	engine *value.Engine // 添加 Engine 字段
}

func NewConditionEngine(client *clients.ClientSet) *ConditionEngine {
	engine := value.NewEngine(client)
	return &ConditionEngine{
		engine: engine, // 初始化 Engine
	}
}

// o传入的是一个对象，可能是workflow、task、group、action、runtime等
// o传入的是传入值,而不是指针
func (ce *ConditionEngine) CheckConditions(conditions *apis.Conditions, o interface{}) (apis.ResultType, error) {
	if conditions == nil {
		logs.Error("condition is nil")
		return apis.True, errors.New("condition is nil")
	}

	if len(conditions.Formulas) == 0 {
		return apis.True, nil
	}

	for _, formula := range conditions.Formulas {
		checkRes, err := ce.checkFormula(&formula, o)
		if err != nil {
			return apis.False, err
		}
		if checkRes != apis.True {
			return checkRes, nil
		}
	}
	return apis.True, nil
}

func (ce *ConditionEngine) checkFormula(formula *apis.ConditionFormula, o interface{}) (apis.ResultType, error) {
	switch formula.ConditionType {
	case apis.NodeDependency:
		res, err := ce.checkNodeDependency(formula, o)
		if err != nil {
			return apis.False, err
		}
		return res, nil
	case apis.DataDependency:
		res, err := ce.checkDataDependency(formula, o)
		if err != nil {
			return apis.False, err
		}
		return res, nil
	case apis.ResourceDependency:
		res, err := ce.checkResourceDependency(formula, o)
		if err != nil {
			return apis.False, err
		}
		return res, nil
	case apis.ProgramDependency:
		res, err := ce.checkProgramDependency(formula, o)
		if err != nil {
			return apis.False, err
		}
		return res, nil
	default:
		logs.Error("unsupported condition type ", formula.ConditionType)
		return apis.False, errors.New(string("unsupported signal type " + formula.Signal))
	}
}

func (ce *ConditionEngine) checkNodeDependency(formula *apis.ConditionFormula, o interface{}) (apis.ResultType, error) {
	logs.Infof("checkNodeDependency val : %v", o)
	val := reflect.ValueOf(o)
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // 解引用指针，获取指针所指向的值
	}
	kind := val.Type().Name()
	Name := val.FieldByName("Name")
	leftValue := &formula.LeftValue
	//debug日志
	// 解析parent的Phase的值
	newLeftValue, err := ce.engine.ExtractLocalValue(leftValue, o)

	if err != nil {
		logs.Error("condititon Engine: check Dodedependency GetValue error: ", err)
		return apis.False, err
	}
	if newLeftValue.Value == string(apis.Running) || newLeftValue.Value == string(apis.Successed) {
		logs.Infof("nodeDependency satisfied!")
		return apis.True, nil
	} else {
		// 前置节点任务未运行或未完成
		logs.Warnf("condititon Engine: nodedependency is not ready: %v %v's parent not running or seccessed!", kind, Name)
		return apis.NotReady, nil
	}
}

func (ce *ConditionEngine) checkDataDependency(formula *apis.ConditionFormula, o interface{}) (apis.ResultType, error) {
	val := reflect.ValueOf(o)
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // 解引用指针，获取指针所指向的值
	}
	kind := val.Type().Name()

	// Debug日志
	// Name := val.FieldByName("Name")
	// logs.Infof("checkNodeDependency val : %v", val)
	// logs.Infof("checkNodeDependency after val : %v", val)
	// logs.Infof("checkNodeDependency Name : %v, kind:%v", Name, kind)
	leftValue := &formula.LeftValue
	rightValue := &formula.RightValue
	Datatype := formula.LeftValue.Type
	switch Datatype {
	case apis.ConstData:

	case apis.LocalData:
		// newLeftValue为更新后的leftValue，更改的方式是指针引用
		newLeftValue, err := ce.engine.ExtractLocalValue(leftValue, o)
		if err != nil {
			logs.Errorf("checkDataDependency Err: ce.engine.ExtractLocalValue get value failed")
			return apis.False, nil
		}
		if newLeftValue.Value == "" {
			return apis.NotReady, nil
		}
		if newLeftValue.Value != rightValue.Value {
			// return apis.False, errors.New("Local data check: get leftValue != expect rightValue.Get leftValue:" + newLeftValue.Value)
			logs.Tracef("checkDataDependency: %v %v's LocalData check failed, get leftValue %v != expect rightValue.Get leftValue:%v", kind, val.FieldByName("Name"), newLeftValue.Value, rightValue.Value)
			return apis.False, nil
		}
		return apis.True, nil

	case apis.DeviceData:

	case apis.ResultsData:

	case apis.FileData:
		// 后续看情况是否改成使用GetValue的形式获取Value。估计是没必要，直接在Runtime这里访问Data
		// 由于没有Data的Client，暂时使用RuntimeSpec里面的Data而不是Status里面的Data，检查Data[]里面所有的文件是否下载到本地
		switch kind {
		case "Runtime":
			r := (o).(apis.Runtime)
			//拼接目录,直接拼上runtime的Name
			folder := r.Status.Directory
			if _, err := os.Stat(folder); os.IsNotExist(err) {
				// 文件夹不存在的日志
				logs.Tracef("checkDataDependency: %v %v's folder:%v is not exist", kind, r.Name, folder)
			} else if err != nil {
				// 其他异常错误
				logs.Errorf("checkDataDependency: %v %v's folder:%v check failed", kind, r.Name, folder)
				return apis.False, errors.New("dataDependency:folder check failed")
			}
			// 上面检查完目录之后，拼接每一个文件的路径，检查每一个文件是否存在，不存在则NotReady状态。只检查文件是否存在，启动文件下载在其他地方
			for _, data := range r.Spec.Data {
				// 需要根据是文件还是文件夹来分别检查，目前是依据data.Name是否包含.来判断是否是文件
				if data.FileFormat == "file" || strings.Contains(data.Name, ".") {
					// 是文件
					filePath := folder + "/" + data.Name
					// logs.Infof("checkDataDependency: %v %v's file:%v", kind, r.Name, filePath)
					if _, err := os.Stat(filePath); err != nil {
						// 文件不存在的日志
						logs.Warnf("checkDataDependency: %v %v's file:%v is not exist", kind, r.Name, filePath)
						return apis.NotReady, errors.New("dataDependency:file not exist")
					}
				} else {
					// 是文件夹
					folderPath := strings.TrimSuffix(folder, "/") + "/" + data.Name
					// logs.Infof("checkDataDependency: %v %v's folder:%v", kind, r.Name, folderPath)
					// 检查文件夹是否存在
					if _, err := os.Stat(folderPath); os.IsNotExist(err) {
						// 文件夹不存在的日志
						logs.Warnf("checkDataDependency: %v %v's folder:%v is not exist", kind, r.Name, folderPath)
						return apis.NotReady, errors.New("dataDependency:folder not exist")
					}
				}
				// logs.Infof("checkDataDependency: %v %v's file:%v is not exist", kind, r.Name, file)
			}
			return apis.True, nil
		default:
			logs.Error("DataDependency unsupported fileData kind:", kind)
			return apis.False, errors.New("unsupported fileData kind:" + kind)
		}
	default:
		logs.Error("DataDependency unsupported type:", Datatype)
	}
	return apis.True, nil
}

func (ce *ConditionEngine) checkResourceDependency(formula *apis.ConditionFormula, o interface{}) (apis.ResultType, error) {

	return apis.True, nil
}

func (ce *ConditionEngine) checkProgramDependency(formula *apis.ConditionFormula, o interface{}) (apis.ResultType, error) {

	return apis.True, nil
}

// // TODO 解析具体的值，返回bool表示值是否就绪，string表示值
// func (ce *ConditionEngine) extractValue(value apis.Value, o interface{}) (bool, string) {

// 	switch value.Type {
// 	case apis.ConstData:
// 		return true, value.Value

// 	case apis.LocalData:
// 		return true, value.Value
// 		////TODO 从client里面拿结果 校验
// 	case apis.DeviceData:
// 		val, err := ce.engine.GetValue(&value, o)
// 		if err != nil {
// 			logs.Error("extract value error: ", err)
// 			return false, "0"
// 		}
// 		if val.Value == "" {
// 			return false, "0"
// 		}
// 		//检查数据是否存在，存在返回“1”即可，与rightVal的“1”进行比较
// 		return true, "1"

// 	default:
// 		logs.Fatal("unsupported value type ", value.ValueType)
// 		return false, ""
// 	}
// }

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
