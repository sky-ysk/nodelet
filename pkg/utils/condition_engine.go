package utils

import (
	"errors"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/utils/value"
)

// 为后续做成有状态的类留出扩展
type ConditionEngine struct {
	engine *value.Engine // 添加 Engine 字段
}

func NewConditionEngine(clientSet *clients.ClientSet) *ConditionEngine {
	return &ConditionEngine{
		engine: value.NewEngine(clientSet), // 初始化 Engine
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
	
	rightReady, rightVal := engine.extractValue(formula.LeftValue)
	leftReady, leftVal := engine.extractValue(formula.RightValue)
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
func (ce *ConditionEngine) extractValue(value apis.Value) (bool, string) {
	
	switch value.Type {
	case apis.ConstData:
		return true, value.Value

	case apis.LocalData:
		return true, value.Value
		////TODO 从client里面拿结果 校验
	case apis.DeviceData:
		val, err := ce.engine.GetValue(&value)
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
