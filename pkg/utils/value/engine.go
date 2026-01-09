package value

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"hit.edu/framework/pkg/component-base/logs"

	apis "hit.edu/framework/pkg/apis/cores"
	metav1 "hit.edu/framework/pkg/apis/meta"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/util/manager"
)

type Engine struct {
	manager  *manager.Manager
	comparor *RegExprComparor
}

// 传入Engine的ClientSet
func NewEngine(clientSet *clients.ClientSet) *Engine {
	e := Engine{
		manager:  manager.NewManager(clientSet),
		comparor: NewRegExprComparor(),
	}
	return &e
}

type KindType string

const (
	ActionType   KindType = "Action"
	RuntimeType  KindType = "Runtime"
	WorkflowType KindType = "Workflow"
	TaskType     KindType = "Task"
	GroupType    KindType = "Group"
	UnknownType  KindType = "Unknown"
)

// Output对应的正则表达式
// Status对应的正则表达式
// Device对应的正则表达式

// 解析Value的值, o为传入的对象，可能为Workflow、Task、Group、Action或Runtime
// 解析的Value为condition里面的左右值、或者是input、output（这俩是Value类型的变量）
// 解析的结果仍然是一个Value类型的变量，其Value.Value为值的string表达，其Value.ValueType为预定义好的类型（例如bool，string等）
// 解析完Value之后，需要判断类型是否符合需求，然后再转换Value.Value为所需类型使用
func (e *Engine) GetValue(value *apis.Value, o interface{}) (*apis.Value, error) {
	// 判断数据类型
	switch value.Type {
	case apis.ConstData:
		// 常量类型，不需要进行处理
		return value, nil
	case apis.LocalData:
		// 本地进行寻址，寻找本地变量
		// 寻址范围在Workflow、Task、Group、Action和Runtime的相关字段，主要为Outputs、Status
		// 使用Spec.Name进行寻址，与实际的Name无关

		// Local的访问格式对应
		//  Workflow{W1}.Task{T1}.Group{G1}.Action{A1}.Runtime{R1}
		//  Task{T1}.Group{G1}.Action{A1}.Runtime{R1}
		//  Group{G1}.Action{A1}.Runtime{R1}
		//  Action{A1}.Runtime{R1}
		//  Runtime{R1}

		//  Workflow{W1}.Task{T1}.Group{G1}.Action{A1}
		//  Workflow{W1}
		//  Workflow{W1}.Task{T1}.Group{G1}
		//  Workflow{W1}.Task{T1}

		// 缺省值，默认访问本地
		resultValue, err := e.ExtractLocalValue(value, o)
		if err != nil {
			return nil, err
		}
		return resultValue, nil
	case apis.DeviceData:
		// 对Device进行寻址
		// result, err := e.ExtractDeviceValue( , value.From, value.NameSpace)
		// if err != nil {
		// 	return nil, err
		// }
		// value.Value = result
		// value.ValueType = apis.StringType
		return value, nil
	case apis.ResultsData:
		return value, nil
	default:
		return nil, errors.New(string("Unsupported DataType " + value.Type))
	}
}

func (e *Engine) ExtractDeviceValue(devices []apis.DeviceSpec, from string, namespace string) (string, error) {
	kind := "Device"

	_, parts, err := e.comparor.Match(kind, from)
	if err != nil {
		return "", errors.New("Unsupported kind " + kind)
	}
	switch kind {
	case "Device":
		name := parts[1]
		realName := ""
		for _, device := range devices {
			if device.Name == name {
				if ep, ok := device.ExpectedProperties["name"]; ok {
					realName = ep.Value
				}
			}
		}
		if realName == "" {
			return "", fmt.Errorf("[ENGINE] can not find Device[%s]'s real name", name)
		}
		namespace := namespace
		ability := parts[2]
		service := parts[3]
		fmt.Println(name, namespace, ability, service)
		s, err := e.ExtractDeviceService(realName, namespace, ability, service)
		if err == nil {
			// fmt.Println("success", s)
			return s, nil
		}
	}
	return "", errors.New("Unsupported kind " + kind)
}

func (e *Engine) ExtractDeviceImage(from string) (string, string, string, error) {
	kind := "Device"

	_, parts, err := e.comparor.Match(kind, from)
	logs.Info(from)
	if err != nil {
		logs.Error(err.Error())
		return "", "", "", err
	}
	switch kind {
	case "Device":
		name := parts[1]

		ability := parts[2]
		service := parts[3]
		return name, ability, service, nil

	}
	return "", "", "", errors.New("Unsupported kind " + kind)
}

func (e *Engine) ExtractLocalValue(value *apis.Value, o interface{}) (*apis.Value, error) {
	kind := reflect.TypeOf(o).Name()
	// logs.Infof("kind %s, value %v", kind, value)
	// 下面这一个地方经常报错，原因是Kind匹配问题，如果o是runtime类型的，那么kind就是runtime，无法越级访问上层的正则表达式
	// 导致Action{}.Name{}这种匹配不到====目前的解决方案，调整runtime的规则范围使其能够概括所有需要的规则
	// 依复杂情况可以改成单独设计正则规则和解析规则，与device类似：新增input类型的datatype，然后再编写对应的解析函数
	typeName, parts, err := e.comparor.Match(kind, value.From)
	if err != nil {
		logs.Errorf("Engine: Extract Local Value Err!")
		return nil, errors.New("Unsupported kind " + kind)
	}
	var name string
	var kindType string
	var from string
	var fromKey string
	var namespace string

	switch kind {
	case "Runtime":
		r := (o).(apis.Runtime)
		namespace = r.Namespace
		name, kindType, from, fromKey, err = e.GetNameFromRuntime(typeName, parts, &r)
		logs.Infof(" name %s, kindType %s, from %s %s", name, kindType, from, fromKey)
		if err != nil {
			logs.Error(err.Error())
			return nil, err
		}
	case "Action":
		r := (o).(apis.Action)
		namespace = r.Namespace
		name, kindType, from, fromKey, err = e.GetNameFromAction(typeName, parts, &r)
		logs.Infof(" name %s, kindType %s, from %s %s", name, kindType, from, fromKey)
		if err != nil {
			return nil, err
		}
	case "Group":
		r := (o).(apis.Group)
		namespace = r.Namespace
		name, kindType, from, fromKey, err = e.GetNameFromGroup(typeName, parts, &r)
		// fmt.Println(name, kindType, from, fromKey, err)
		if err != nil {
			return nil, err
		}
	case "Task":
		r := (o).(apis.Task)
		namespace = r.Namespace
		name, kindType, from, fromKey, err = e.GetNameFromTask(typeName, parts, &r)
		if err != nil {
			return nil, err
		}
	case "Workflow":
		r := (o).(apis.Workflow)
		namespace = r.Namespace
		name, kindType, from, fromKey, err = e.GetNameFromWorkflow(typeName, parts, &r)
		if err != nil {
			return nil, err
		}
	}

	switch kindType {
	case "Runtime":
		v, err := e.ExtractRuntimeValue(name, namespace, from, fromKey, value)
		if err == nil {
			return v, nil
		}
	case "Action":
		v, err := e.ExtractActionValue(name, namespace, from, fromKey, value)
		if err == nil {
			return v, nil
		}
	case "Group":
		v, err := e.ExtractGroupValue(name, namespace, from, fromKey, value)
		if err == nil {
			return v, nil
		}
	case "Task":
		v, err := e.ExtractTaskValue(name, namespace, from, fromKey, value)
		if err == nil {
			return v, nil
		}
	case "Workflow":
		v, err := e.ExtractWorkflowValue(name, namespace, from, fromKey, value)
		if err == nil {
			return v, nil
		}
	}
	fmt.Println("unsuported kindType:", kindType)
	return nil, errors.New("Failed to extract local value")
}

func (e *Engine) GetNameFromWorkflow(name string, parts []string, workflow *apis.Workflow) (string, string, string, string, error) {
	switch name {
	case "WorkflowExpr":
		wn := workflow.Name
		from := parts[2]
		fromKey := parts[3]
		return wn, string(WorkflowType), from, fromKey, nil

	case "WorkflowAbsoluteExpr":
		// 只允许Workflow为最高层时使用
		uuid := workflow.Labels["uuid"]
		// TODO: 检查Parts长度
		workflowName := parts[1]
		taskName := parts[2]
		groupName := parts[3]
		actionName := parts[4]
		runtimeName := parts[5]
		from := parts[6]
		fromKey := parts[7]
		rn := fmt.Sprintf("%s.%s.%s.%s.%s-%s", workflowName, taskName, groupName, actionName, runtimeName, uuid)
		return rn, string(RuntimeType), from, fromKey, nil
	case "WorkflowTaskExpr":
		// 无需实现
		return "", string(UnknownType), "", "", errors.New(string("Unsupported workflow " + name))
	}
	return "", string(UnknownType), "", "", errors.New(string("Unsupported workflow " + name))
}

func (e *Engine) GetNameFromTask(name string, parts []string, task *apis.Task) (string, string, string, string, error) {
	switch name {
	case "TaskExpr":
		// Task{}位置
		target := parts[1]
		from := parts[2]
		fromKey := parts[3]
		if target == task.Spec.Name {
			// 寻址的是当前的Task
			tn := task.Name
			return tn, string(TaskType), from, fromKey, nil
		} else {
			// 寻址的是当前Workflow下的Task
			// 根据Belong查找Workflow
			// TODO: 改成根据belong的label去查询
			w, err := e.manager.GetWorkflow(task.Status.Belong.Name, task.Status.Belong.Namespace)
			if err == nil {
				t, ok := w.Status.Tasks[target]
				if ok {
					return t.Name, string(TaskType), from, fromKey, nil
				}
			}
		}
	case "TaskGroupExpr":
		// Task{}位置
		targetTask := parts[1]
		targetGroup := parts[2]
		from := parts[3]
		fromKey := parts[4]
		if targetTask == task.Spec.Name {
			// 寻址的是当前的Task
			g, ok := task.Status.Groups[targetGroup]
			if ok {
				return g.Name, string(GroupType), from, fromKey, nil
			}
		} else {
			// 寻址的是当前Workflow下的Task
			workflow, err := e.manager.GetWorkflow(task.Status.Belong.Name, task.Status.Belong.Namespace)
			if err == nil {
				tn, ok := workflow.Status.Tasks[targetTask]
				if ok {
					t, err := e.manager.GetTask(tn.Name, tn.Namespace)
					if err == nil {
						g, ok := t.Status.Groups[targetGroup]
						if ok {
							return g.Name, string(GroupType), from, fromKey, nil
						}
					}
				}
			}
		}
	case "TaskAbsoluteExpr":
		// 只允许Task为最高层时使用
		if task.Status.Belong == nil {
			uuid := task.Labels["uuid"]
			taskName := parts[1]
			groupName := parts[2]
			actionName := parts[3]
			runtimeName := parts[4]
			from := parts[5]
			fromKey := parts[6]
			rn := fmt.Sprintf("%s.%s.%s.%s-%s", taskName, groupName, actionName, runtimeName, uuid)
			return rn, string(RuntimeType), from, fromKey, nil
		}
	case "WorkflowTaskExpr":
		// TODO:
		return "", string(UnknownType), "", "", errors.New(string("Unsupported task " + name))
	}
	return "", string(UnknownType), "", "", errors.New(string("Unsupported task " + name))
}

func (e *Engine) GetNameFromGroup(name string, parts []string, group *apis.Group) (string, string, string, string, error) {
	switch name {
	case "GroupExpr":
		// Group{}位置
		// fmt.Println("GroupExpr")
		target := parts[1]
		from := parts[2]
		fromKey := parts[3]
		if target == group.Spec.Name {
			// 寻址的是当前的Group
			gn := group.Name
			return gn, string(GroupType), from, fromKey, nil
		} else {
			// 寻址的是当前Workflow下的Group
			// 根据Belong查找Workflow
			// TODO: 改成根据belong的label去查询
			t, err := e.manager.GetTask(group.Status.Belong.Name, group.Status.Belong.Namespace)
			if err == nil {
				g, ok := t.Status.Groups[target]
				if ok {
					return g.Name, string(GroupType), from, fromKey, nil
				}
			}
		}
	case "GroupActionExpr":
		// Group{}位置
		fmt.Println("GroupActionExpr")
		targetGroup := parts[1]
		targetAction := parts[2]
		from := parts[3]
		fromKey := parts[4]
		fmt.Println(targetAction, from, fromKey)
		fmt.Println(targetGroup, group.Spec.Name)
		if targetGroup == group.Spec.Name {
			// 寻址的是当前的Group
			a, ok := group.Status.Actions[targetAction]
			fmt.Println("okkkk", ok)
			if ok {
				fmt.Println("ok")
				return a.Name, string(ActionType), from, fromKey, nil
			}
		} else {
			// 寻址的是当前Task下的Group的Action
			fmt.Println("else")
			// 根据Belong查找Task
			task, err := e.manager.GetTask(group.Status.Belong.Name, group.Status.Belong.Namespace)
			if err == nil {
				gn, ok := task.Status.Groups[targetGroup]
				if ok {
					g, err := e.manager.GetGroup(gn.Name, gn.Namespace)
					if err == nil {
						a, ok := g.Status.Actions[targetAction]
						if ok {
							return a.Name, string(ActionType), from, fromKey, nil
						}
					}
				}
			}
		}
	case "GroupActionRuntimeExpr":
		// 只允许Group为最高层时使用
		if group.Status.Belong == nil {
			uuid := group.Labels["uuid"]
			groupName := parts[1]
			actionName := parts[2]
			runtimeName := parts[3]
			rn := fmt.Sprintf("%s.%s.%s-%s", groupName, actionName, runtimeName, uuid)
			return rn, string(RuntimeType), "", "", nil
		}

		// add
		// Group{}位置
		emptyStr := ""
		fmt.Println("GroupActionRuntimeExpr")
		targetGroup := parts[1]
		targetAction := parts[2]
		targetRuntime := parts[3]
		from := parts[4]
		fromKey := parts[5]
		fmt.Println(targetRuntime, from, fromKey)
		fmt.Println(targetGroup, group.Spec.Name)
		if targetGroup == group.Spec.Name {
			// 寻址的是当前的Group
			fmt.Println("current group")

			// group下actions
			an, ok := group.Status.Actions[targetAction]
			if !ok {
				logs.Infof("no action found for %s in %s", targetAction, targetGroup)
				return emptyStr, emptyStr, emptyStr, emptyStr, errors.New("no action found")
			}

			// action
			a, err := e.manager.GetAction(an.Name, an.Namespace)
			if err != nil {
				logs.Info(err)
				return emptyStr, emptyStr, emptyStr, emptyStr, err
			}

			// action下runtimes
			r, ok := a.Status.Runtimes[targetRuntime]
			if !ok {
				logs.Infof("no runtime found for %s in %s", targetRuntime, targetAction)
				return emptyStr, emptyStr, emptyStr, emptyStr, errors.New("runtime not found")
			}

			return r.Name, string(RuntimeType), from, fromKey, nil

		} else {
			// 寻址的是当前Task下的其他Group
			fmt.Println("other group belong same task")
			// 根据Belong查找Task
			task, err := e.manager.GetTask(group.Status.Belong.Name, group.Status.Belong.Namespace)
			if err != nil {
				logs.Info(err)
				return emptyStr, emptyStr, emptyStr, emptyStr, err
			}

			//task下groups
			gn, ok := task.Status.Groups[targetGroup]
			if !ok {
				logs.Info(err)
				return emptyStr, emptyStr, emptyStr, emptyStr, errors.New("group not found")
			}

			// group
			g, err := e.manager.GetGroup(gn.Name, gn.Namespace)
			if err != nil {
				logs.Info(err)
				return emptyStr, emptyStr, emptyStr, emptyStr, err
			}

			// group下actions
			an, ok := g.Status.Actions[targetAction]
			if !ok {
				logs.Info(err)
				return emptyStr, emptyStr, emptyStr, emptyStr, errors.New("action not found")
			}

			// action
			a, err := e.manager.GetAction(an.Name, an.Namespace)
			if err != nil {
				logs.Info(err)
				return emptyStr, emptyStr, emptyStr, emptyStr, err
			}

			// action下runtimes
			r, ok := a.Status.Runtimes[targetRuntime]
			if !ok {
				logs.Info(err)
				return emptyStr, emptyStr, emptyStr, emptyStr, errors.New("runtime not found")
			}
			fmt.Printf("return r.Name:%v, string(RuntimeType):%v, from:%v, fromKey:%v, err:%v\n ", r.Name, string(RuntimeType), from, fromKey, nil)

			return r.Name, string(RuntimeType), from, fromKey, nil

		}
	//case "GroupAbsoluteExpr":
	//	fmt.Println("GroupAbsoluteExpr")
	//	// 只允许Group为最高层时使用
	//	if group.Status.Belong == nil {
	//		uuid := group.Labels["uuid"]
	//		groupName := parts[1]
	//		actionName := parts[2]
	//		runtimeName := parts[3]
	//		rn := fmt.Sprintf("%s.%s.%s-%s", groupName, actionName, runtimeName, uuid)
	//		return rn, string(RuntimeType), "", "", nil
	//	}

	case "TaskGroupExpr":
		fmt.Println("TaskGroupExpr")
		// TODO:
		return "", string(UnknownType), "", "", errors.New(string("Unsupported group " + name))
	}

	fmt.Println("end", name)
	return "", string(UnknownType), "", "", errors.New(string("Unsupported group " + name))
}

func (e *Engine) GetNameFromAction(name string, parts []string, action *apis.Action) (string, string, string, string, error) {
	switch name {
	case "ActionExpr":
		// Action{}位置
		target := parts[1]
		from := parts[2]
		fromKey := parts[3]
		if target == action.Spec.Name {
			// 寻址的是当前的Action
			gn := action.Name
			return gn, string(ActionType), from, fromKey, nil
		} else {
			// 寻址的是当前Workflow下的Action
			// 根据Belong查找Workflow
			// TODO: 改成根据belong的label去查询
			g, err := e.manager.GetGroup(action.Status.Belong.Name, action.Status.Belong.Namespace)
			if err == nil {
				a, ok := g.Status.Actions[target]
				if ok {
					return a.Name, string(ActionType), from, fromKey, nil
				}
			}
		}
	case "ActionRuntimeExpr":
		// Action{}位置
		targetAction := parts[1]
		targetRuntime := parts[2]
		from := parts[3]
		fromKey := parts[4]
		if targetAction == action.Spec.Name {
			// 寻址的是当前的Action
			a, ok := action.Status.Runtimes[targetRuntime]
			if ok {
				return a.Name, string(RuntimeType), from, fromKey, nil
			}
		} else {
			// 寻址的是当前Workflow下的Action
			group, err := e.manager.GetGroup(action.Status.Belong.Name, action.Status.Belong.Namespace)
			if err == nil {
				an, ok := group.Status.Actions[targetAction]
				if ok {
					a, err := e.manager.GetAction(an.Name, an.Namespace)
					if err == nil {
						r, ok := a.Status.Runtimes[targetRuntime]
						if ok {
							return r.Name, string(RuntimeType), from, fromKey, nil
						}
					}
				}
			}
		}
	case "GroupActionExpr":
		// TODO: 暂不支持
		// 要寻址Group
		//  如果Group的Spec.Name与Target相同，则与ActionExpr相同
		//  如果不同，则需要找到上层Group的Task,然后找到对应的Group，最后再找到对应的Action
		return "", string(UnknownType), "", "", errors.New(string("Unsupported action " + name))
	}
	return "", string(UnknownType), "", "", errors.New(string("Unsupported action " + name))
}

func (e *Engine) GetNameFromRuntime(name string, parts []string, runtime *apis.Runtime) (string, string, string, string, error) {
	switch name {
	case "RuntimeExpr":
		// Runtime{}位置
		target := parts[1]
		from := parts[2]
		fromKey := parts[3]

		if target == runtime.Spec.Name {
			// 寻址的是当前的Runtime
			gn := runtime.Name
			return gn, string(RuntimeType), from, fromKey, nil
		} else {
			// 寻址的是当前Workflow下的Runtime
			// 根据Belong查找Workflow
			// TODO: 改成根据belong的label去查询
			a, err := e.manager.GetAction(runtime.Status.Belong.Name, runtime.Status.Belong.Namespace)
			if err == nil {
				r, ok := a.Status.Runtimes[target]
				if ok {
					return r.Name, string(RuntimeType), from, fromKey, nil
				}
			}
		}
	case "ActionRuntimeExpr":
		// Runtime{}位置
		//targetAction := parts[1]
		targetRuntime := parts[2]
		from := parts[3]
		fromKey := parts[4]

		// 寻址的是当前Action下的Runtime
		action, err := e.manager.GetAction(runtime.Status.Belong.Name, runtime.Status.Belong.Namespace)
		if err == nil {
			r, ok := action.Status.Runtimes[targetRuntime]
			if ok {
				return r.Name, string(RuntimeType), from, fromKey, nil
			}
		}
	case "GroupExpr":
		// Group{G1}.Name{xx}
		// targetGroup := parts[1]
		from := parts[2]
		fromKey := parts[3]

		// 寻址的是当前Runtime的父亲节点
		parentAction, err := e.manager.GetAction(runtime.Status.Belong.Name, runtime.Status.Belong.Namespace)
		if err != nil {
			return "", string(UnknownType), "", "", errors.New("Failed to get parent action: " + err.Error())
		}
		parentGroup, err := e.manager.GetGroup(parentAction.Status.Belong.Name, parentAction.Status.Belong.Namespace)
		if err != nil {
			return "", string(UnknownType), "", "", errors.New("Failed to get parent group: " + err.Error())
		}
		return parentGroup.Name, string(GroupType), from, fromKey, nil
		// TODO:如果需要，还可以新增上一级task，然后去遍历找符合groupName的group
	case "GroupAbsoluteExpr":
		// Group{G81}.Action{A1}.Runtime{R1}.Outputs{0}
		// 下面寻址的前提是Group不是本身的Group, 如果需要在本身Group寻址，不填Group这一项，属于其他规则的寻址
		targetGroup := parts[1]
		targetAction := parts[2]
		targetRuntime := parts[3]
		from := parts[4]
		fromKey := parts[5]
		logs.Infof("GetNameFromRuntime:g a r f fk:%v, %v, %v, %v, %v", targetGroup, targetAction, targetRuntime, from, fromKey)
		// 下面先去查对应的Task，从runtime开始
		parentAction, err := e.manager.GetAction(runtime.Status.Belong.Name, runtime.Status.Belong.Namespace)
		if err != nil {
			return "", string(UnknownType), "", "", errors.New("Failed to get parent action: " + err.Error())
		}
		parentGroup, err := e.manager.GetGroup(parentAction.Status.Belong.Name, parentAction.Status.Belong.Namespace)
		if err != nil {
			return "", string(UnknownType), "", "", errors.New("Failed to get parent group: " + err.Error())
		}
		parentTask, err := e.manager.GetTask(parentGroup.Status.Belong.Name, parentGroup.Status.Belong.Namespace)
		if err != nil {
			return "", string(UnknownType), "", "", errors.New("Failed to get parent task: " + err.Error())
		}
		// 找到Task之后，去找对应的G A R
		gref, _ := parentTask.Status.Groups[targetGroup]
		g, err := e.manager.GetGroup(gref.Name, gref.Namespace)
		aref, _ := g.Status.Actions[targetAction]
		a, err := e.manager.GetAction(aref.Name, aref.Namespace)
		rref, _ := a.Status.Runtimes[targetRuntime]
		r, err := e.manager.GetRuntime(rref.Name, rref.Namespace)
		logs.Infof("Get final runtime %v", r.Name)
		return r.Name, string(RuntimeType), from, fromKey, nil
	}
	return "", string(UnknownType), "", "", errors.New(string("Unsupported runtime " + name))
}

// 已经确定是WorkflowExpr
// 目前只支持解析Status的State
// TODO: 使用抽象类
func (e *Engine) ExtractWorkflowValue(workflow string, namespace string, target string, subTarget string, value *apis.Value) (*apis.Value, error) {
	w, err := e.manager.GetWorkflow(workflow, namespace)
	if err != nil {
		return nil, err
	}

	switch target {
	case "Status":
		// 目前只支持Status.Phase
		// TODO: 增加更多类型
		value.Value = string(w.Status.Phase)
		return value, nil
	}
	return nil, errors.New(string("Unsupported Target " + target))
}

// 目前只支持解析Status的State
func (e *Engine) ExtractTaskValue(task string, namespace string, target string, subTarget string, value *apis.Value) (*apis.Value, error) {
	t, err := e.manager.GetTask(task, namespace)
	if err != nil {
		return nil, err
	}

	switch target {
	case "Status":
		// 目前只支持Status.Phase
		// TODO: 增加更多类型
		value.Value = string(t.Status.Phase)
		return value, nil
	}
	return nil, errors.New(string("Unsupported Target " + target))
}

// 目前只支持解析Status的State
func (e *Engine) ExtractGroupValue(group string, namespace string, target string, subTarget string, value *apis.Value) (*apis.Value, error) {
	g, err := e.manager.GetGroup(group, namespace)
	if err != nil {
		return nil, err
	}

	switch target {
	case "Status":
		// 目前只支持Status.Phase
		// TODO: 增加更多类型
		value.Value = string(g.Status.Phase)
		return value, nil
	case "Name":
		value.Value = string(g.Name)
		value.ValueType = apis.StringType
		return value, nil
	}
	return nil, errors.New(string("Unsupported Target " + target))
}

// 目前只支持解析Status的State
func (e *Engine) ExtractActionValue(action string, namespace string, target string, subTarget string, value *apis.Value) (*apis.Value, error) {
	a, err := e.manager.GetAction(action, namespace)
	if err != nil {
		return nil, err
	}

	switch target {
	case "Status":
		// 目前只支持Status.Phase
		// TODO: 增加更多类型
		value.Value = string(a.Status.Phase)
		return value, nil
	case "Name":
		value.Value = string(a.Name)
		value.ValueType = apis.StringType
		return value, nil
	}
	return nil, errors.New(string("Unsupported Target " + target))
}

// 目前只支持解析Status的State和Outputs
func (e *Engine) ExtractRuntimeValue(runtime string, namespace string, target string, subTarget string, value *apis.Value) (*apis.Value, error) {
	logs.Infof("runtime %s, namespace %s, target %s, subtarget %s", runtime, namespace, target, subTarget)
	r, err := e.manager.GetRuntime(runtime, namespace)
	if err != nil {
		return nil, err
	}
	fmt.Printf("target:%v\n", target)

	switch target {
	case "Status":
		// 目前只支持Status.Phase
		// TODO: 增加更多类型
		value.Value = string(r.Status.Phase)
		return value, nil
	case "Outputs":
		// 检查
		// v, ok := r.Status.Outputs[subTarget]
		// if !ok {
		// 	return nil, errors.New(string("SubTarget is not existed" + subTarget))
		// }
		// 下面是测试使用的修改，获取r.Spec的列表Outputs，而不是r.Status里面的map Outputs
		// 将subTarget转换为数字
		index, err := strconv.Atoi(subTarget)
		if err != nil {
			return nil, errors.New("SubTarget is not a valid number: " + subTarget)
		}

		// 检查索引是否有效
		if index < 0 || index >= len(r.Spec.Outputs) {
			return nil, errors.New("SubTarget index out of range: " + subTarget)
		}

		// 获取对应的输出
		v := r.Spec.Outputs[index]
		fmt.Printf("output:%v\n", v)

		value.Value = v.Value
		value.ValueType = v.ValueType
		return value, nil
	case "Name":
		value.Value = string(r.Name)
		value.ValueType = apis.StringType
		return value, nil
	}
	return nil, errors.New(string("Unsupported Target " + target))
}

// 解析Device的Service
//
//	支持解析Ability和对应的Service, 用于Device类型任务的部署
//	格式为Device{Robot}.Abiltiy{Move}.Service{S1}
//
// TODO: 解析Device字段
func (e *Engine) ExtractDeviceService(robot string, namespace string, target string, subTarget string) (string, error) {
	// 暂时直接使用客户端，后续改为使用manager
	client := e.manager.ClientSet.Core().Devices(namespace)
	d, err := client.Get(context.TODO(), robot, metav1.GetOptions{})
	// logs.Info("get device:", d)
	if err != nil {
		return "", err
	}

	// 直接访问对应的能力
	a, ok := d.Status.Abilities[target]
	// logs.Info("get a:", a)
	if ok {
		s, ok := a.Services[subTarget]
		if ok {
			r := fmt.Sprintf("%s:%s/%s", *s.Ip, *s.Port, *s.Interface)
			return r, nil
		}
	} else {
	}

	return "", errors.New(string("Unsupported Target " + target))
}

// TODO:数据类型转换
