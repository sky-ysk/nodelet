package lib

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
	"strconv"
	"strings"
)

func SendServiceRequest(ability string, device *apis.Device) (apis.Output, error) {
	switch ability {
	case "ArmAngle":
		params := constructArmAngleParams(device)
		url, err := GetServiceUrl(ability, device)
		if err != nil {
			logs.Error(err)
			return apis.Output{}, err
		}
		logs.Infof("publish %s service", ability)
		err = PublishArmAngleInst(params, url)
		if err != nil {
			logs.Errorf("publish armangle inst failed, error is %v", err)
			return apis.Output{}, err
		}
		return apis.Output{}, nil
	case "Predict":
		imagePath := device.Spec.ExpectedProperties["path"].Value
		url, err := GetServiceUrl(ability, device)
		if err != nil {
			logs.Error(err)
			return apis.Output{}, err
		}
		logs.Infof("publish %s service", ability)
		predictResult, err := PublishPredictInst(imagePath, url)
		if err != nil {
			logs.Error("publish predict inst failed, error is %v", err)
			return apis.Output{}, err
		}

		return apis.Output{
			Type:      apis.ResultsData,
			Name:      "predict_result",
			Value:     strconv.Itoa(predictResult.Prediction),
			ValueType: "int",
		}, err
	case "LeftArmUp":
		url, err := GetServiceUrl(ability, device)
		if err != nil {
			logs.Error(err)
			return apis.Output{}, err
		}
		logs.Infof("publish %s service", ability)
		err = PublishLeftArmUpInst(url)
		if err != nil {
			logs.Error("publish predict inst failed, error is %v", err)
			return apis.Output{}, err
		}

		return apis.Output{}, err

	case "LeftArmDown":
		url, err := GetServiceUrl(ability, device)
		logs.Infof("URL IS %v", url)
		if err != nil {
			logs.Error(err)
			return apis.Output{}, err
		}
		logs.Infof("publish %s service", ability)
		err = PublishLeftArmDownInst(url)
		if err != nil {
			logs.Error("publish predict inst failed, error is %v", err)
			return apis.Output{}, err
		}
		return apis.Output{}, err

	case "PredictByUrl":
		url, err := GetServiceUrl(ability, device)
		if err != nil {
			logs.Error(err)
			return apis.Output{}, err
		}

		// 构建参数
		logs.Infof("construct params for predict by url")

		compressed, err := strconv.ParseBool(device.Spec.ExpectedProperties["compressed"].Value)
		if err != nil {
			logs.Errorf("parse compressed value failed, error is %v", err)
			return apis.Output{}, err
		}
		imageType := device.Spec.ExpectedProperties["imageType"].Value
		position := device.Spec.ExpectedProperties["position"].Value
		cameraUrl := device.Spec.ExpectedProperties["cameraUrl"].Value

		// 构建完毕
		logs.Infof("construct params for predict by url finished")
		logs.Infof("publish %s service", ability)
		predictResult, err := PublishPredictByUrlInst(compressed, cameraUrl, position, imageType, url)
		if err != nil {
			logs.Error("publish predict inst failed, error is %v", err)
			return apis.Output{}, err
		}

		return apis.Output{
			Type:      apis.ResultsData,
			Name:      "predict_result",
			Value:     strconv.Itoa(predictResult.Prediction),
			ValueType: "int",
		}, err
	case "TaskState":
		url, err := GetServiceUrl(ability, device)
		if err != nil {
			logs.Error(err)
			return apis.Output{}, err
		}
		logs.Infof("publish %s service", ability)
		taskState, err := PublishTaskState(url)
		if err != nil {
			logs.Error("publish predict inst failed, error is %v", err)
			return apis.Output{}, err
		}
		return apis.Output{
			Type:  apis.ResultsData,
			Name:  "task_state",
			Value: strconv.Itoa(taskState.TaskState),
		}, err
	case "GoStandBy":
		url, err := GetServiceUrl(ability, device)
		if err != nil {
			logs.Error(err)
			return apis.Output{}, err
		}
		taskType, err := strconv.Atoi(device.Spec.ExpectedProperties["taskType"].Value)
		if err != nil {
			logs.Error(err)
			return apis.Output{}, err
		}
		logs.Infof("construct params for GoStandBy finished")
		logs.Infof("publish %s service", ability)
		resp, err := PublishGoStandBy(taskType, url)
		if err != nil {
			logs.Error("publish go stand by inst failed, error is %v", err)
		}
		return apis.Output{
			Type:  apis.ResultsData,
			Name:  "status",
			Value: resp.Status,
		}, err
	case "StartTask":
		url, err := GetServiceUrl(ability, device)
		if err != nil {
			logs.Error(err)
			return apis.Output{}, err
		}
		taskType, err := strconv.Atoi(device.Spec.ExpectedProperties["taskType"].Value)
		if err != nil {
			logs.Error(err)
			return apis.Output{}, err
		}
		logs.Infof("construct params for GoStandBy finished")
		logs.Infof("publish %s service", ability)
		resp, err := PublishStartTask(taskType, url)
		if err != nil {
			logs.Error("publish start task inst failed, error is %v", err)
			return apis.Output{}, err
		}
		return apis.Output{
			Type:  apis.ResultsData,
			Name:  "status",
			Value: resp.Status,
		}, err
	}

	logs.Errorf("Do not support ability:%v", ability)
	return apis.Output{}, nil
}

func constructArmAngleParams(device *apis.Device) map[string][]float64 {
	params := make(map[string][]float64)
	for name, property := range device.Spec.ExpectedProperties {
		// 按逗号分隔字符串
		parts := strings.Split(property.Value, ",")

		// 创建结果切片
		var result []float64
		// 遍历每个部分，转换为 float64
		for _, part := range parts {
			// 转换为 float64
			value, err := strconv.ParseFloat(part, 64)
			if err != nil {
				return nil
			}
			result = append(result, value)
		}
		params[name] = result
	}
	return params
}

func GetServiceUrl(name string, device *apis.Device) (string, error) {
	var url string
	for _, abilityStatus := range device.Status.Abilities {
		for _, serviceStatus := range abilityStatus.Services {
			if serviceStatus.Name == name {
				// 按照ip 端口 接口的方式来构造url
				url = fmt.Sprintf("http://%s:%s%s", serviceStatus.Ip, serviceStatus.Port, serviceStatus.Interface)
				logs.Infof("Get service url %s", url)
				return url, nil
			}
		}
	}
	logs.Errorf("service %s is not exist", name)
	return "", fmt.Errorf("service %s is not exist", name)

}
