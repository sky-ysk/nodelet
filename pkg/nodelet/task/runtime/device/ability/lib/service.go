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
		err = PublishArmAngleInst(params, url)
		if err != nil {
			logs.Errorf("publish armangle inst failed, error is %v", err)
			return apis.Output{}, err
		}
	case "Predict":
		imagePath := device.Spec.ExpectedProperties["path"].Value
		url, err := GetServiceUrl(ability, device)
		if err != nil {
			logs.Error(err)
			return apis.Output{}, err
		}
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
		err = PublishLeftArmUpInst(url)
		if err != nil {
			logs.Error("publish predict inst failed, error is %v", err)
			return apis.Output{}, err
		}

		return apis.Output{}, err

	case "LeftArmDown":
		url, err := GetServiceUrl(ability, device)
		if err != nil {
			logs.Error(err)
			return apis.Output{}, err
		}
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
				break
			} else {
				logs.Errorf("service %s is not exist", name)
				return "", fmt.Errorf("service %s is not exist", name)
			}
		}
	}
	return url, nil
}
