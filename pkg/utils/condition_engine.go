package utils

import (
	"errors"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
)

// 为后续做成有状态的类留出扩展
type ConditionEngine struct {
}

func NewConditionEngine() *ConditionEngine {
	return &ConditionEngine{}
}

func (engine *ConditionEngine) CheckConditions(conditions apis.Conditions) (apis.ResultType, error) {
	
	//if len(conditions.Formulas) == 0 {
	//	return apis.True, nil
	//}
	//
	//for _, formula := range conditions.Formulas {
	//	checkRes, err := engine.checkFormula(formula)
	//	if err != nil {
	//		return apis.False, err
	//	}
	//	if checkRes != apis.True {
	//		return checkRes, nil
	//	}
	//}
	return apis.True, nil
}

func (engine *ConditionEngine) checkFormula(formula apis.ConditionFormula) (apis.ResultType, error) {
	
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
	} else if formula.Signal == apis.Equal {
		if leftVal == rightVal {
			return apis.False, nil
		}
		return apis.True, nil
	}
	logs.Error("unsupported signal type ", formula.Signal)
	return apis.False, errors.New(string("unsupported signal type " + formula.Signal))
}

// TODO 解析具体的值，返回bool表示值是否就绪，string表示值
func (engine *ConditionEngine) extractValue(value apis.Value) (bool, string) {
	
	switch value.DataType {
	case apis.Constant:
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
