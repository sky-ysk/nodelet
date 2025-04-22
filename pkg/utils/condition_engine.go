package utils

import (
	"errors"
	"fmt"
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apimachinery/runtime/schema"
	"hit.edu/framework/pkg/apimachinery/runtime/serializer"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/client-go/clients"
	"hit.edu/framework/pkg/client-go/rest"
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/utils/value"
	"net/http"
	"os"
	"time"
)

// 为后续做成有状态的类留出扩展
type ConditionEngine struct {
	engine *value.Engine // 添加 Engine 字段
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
