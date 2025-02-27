package utils

import (
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/component-base/logs"
)

type ResourceWorker interface {
	CheckStorageResource(spec apis.ResourceSpec, status apis.ResourceStatus) bool
	CheckNetworkResource(spec apis.ResourceSpec, status apis.ResourceStatus) bool
	CheckComputeResource(spec apis.ResourceSpec, status apis.ResourceStatus) bool
	CheckResource(runtime apis.Runtime, action apis.Action) error
	UpdateResourceStatus(runtime *apis.Runtime, action *apis.Action) error
}

// CheckResource 检查各类资源是否满足
func CheckResource(runtime *apis.Runtime, action *apis.Action) error {
	// TODO:确认action中resource的key
	resources := runtime.Resources
	for _, resource := range resources {
		logs.Infof("checking resource: %s\n", resource.Name)
		switch resource.Type {
		case apis.Storage:
			if CheckStorageResource(resource, action.Status.Resources[resource.Name]) {

			} else {
				return fmt.Errorf("storage resource %s is not satisfied", resource.Name)
			}
		case apis.Network:
			if CheckNetworkResource(resource, action.Status.Resources[resource.Name]) {

			} else {
				return fmt.Errorf("network resource %s is not satisfied", resource.Name)
			}
		case apis.Compute:
			if CheckComputeResource(resource, action.Status.Resources[resource.Name]) {

			} else {
				return fmt.Errorf("compute resource %s is not satisfied", resource.Name)
			}
		}
	}
	return nil
}

// CheckStorageResource 检查存储资源的剩余量是否满足期望值
func CheckStorageResource(spec apis.ResourceSpec, status apis.ResourceStatus) bool {
	expectedValueInBytes := ConvertStorageUnit(spec.ExpectedValue, spec.ExpectedValueUnit)
	reservedValueInBytes := ConvertStorageUnit(status.Reserved, status.ReservedUnit)

	if reservedValueInBytes >= expectedValueInBytes {
		return true
	}
	return false
}

// CheckComputeResource 检查计算资源是否满足
func CheckComputeResource(spec apis.ResourceSpec, status apis.ResourceStatus) bool {
	if spec.Name == "cpu" {
		expectedValue := ConvertComputeUnit(spec.ExpectedValue, spec.ExpectedValueUnit)
		reservedValue := ConvertComputeUnit(status.Reserved, status.ReservedUnit)
		if expectedValue <= reservedValue {
			return true
		}
	} else if spec.Name == "gpu" {
		expectedValue := spec.ExpectedValue
		reservedValue := status.Reserved
		if expectedValue <= reservedValue {
			return true
		}
	}
	return false
}

// CheckNetworkResource 检查网络资源是否满足
func CheckNetworkResource(spec apis.ResourceSpec, status apis.ResourceStatus) bool {
	expectedValue := ConvertNetworkUnit(spec.ExpectedValue, spec.ExpectedValueUnit)
	reservedValue := ConvertNetworkUnit(status.Reserved, status.ReservedUnit)

	if reservedValue >= expectedValue {
		return true
	}
	return false
}

// ConvertStorageUnit 进行存储单位的转换
func ConvertStorageUnit(value float64, unit apis.ResourceUnit) float64 {
	switch unit {
	case apis.StorageKB:
		return value * 1024
	case apis.StorageMB:
		return value * 1024 * 1024
	case apis.StorageGB:
		return value * 1024 * 1024 * 1024
	case apis.StorageTB:
		return value * 1024 * 1024 * 1024 * 1024
	default:
		return value
	}
}

// ConvertNetworkUnit 进行网络单位的转换
func ConvertNetworkUnit(value float64, unit apis.ResourceUnit) float64 {
	switch unit {
	case apis.Networkbps:
		return value
	case apis.NetworkKbps:
		return value * 1000
	case apis.NetworkMbps:
		return value * 1_000_000 // 将 Mbps 转换为 bps
	case apis.NetworkGbps:
		return value * 1_000_000_000 // 将 Gbps 转换为 bps
	default:
		return value // 假设默认单位是 bps
	}
}

// ConvertComputeUnit 进行计算单位的转换
func ConvertComputeUnit(value float64, unit apis.ResourceUnit) float64 {
	switch unit {
	case apis.ComputeMilliCPU:
		return value / 1000 // 将 milliCPU 转换为 CPU
	case apis.ComputeNanoCPU:
		return value / 1_000_000_000 // 将 nanocpu 转换为 CPU
	case apis.ComputeCPU:
		return value // 假设默认单位是 CPU
	default:
		return value
	}
}

// UpdateResourceStatus 更新Resource状态
func UpdateResourceStatus(runtime *apis.Runtime, action *apis.Action) error {
	resources := runtime.Resources
	for _, resource := range resources {
		switch resource.Type {
		case apis.Storage:
			// TODO:更新usage 考虑unit
			r := apis.ResourceStatus{
				Reserved:     action.Status.Resources[resource.Name].Reserved - resource.ExpectedValue,
				Usage:        action.Status.Resources[resource.Name].Usage,
				UsageUnit:    action.Status.Resources[resource.Name].UsageUnit,
				ReservedUnit: action.Status.Resources[resource.Name].ReservedUnit,
			}
			action.Status.Resources[resource.Name] = r
		case apis.Network:
			r := apis.ResourceStatus{
				Reserved:     action.Status.Resources[resource.Name].Reserved - resource.ExpectedValue,
				Usage:        action.Status.Resources[resource.Name].Usage,
				UsageUnit:    action.Status.Resources[resource.Name].UsageUnit,
				ReservedUnit: action.Status.Resources[resource.Name].ReservedUnit,
			}
			action.Status.Resources[resource.Name] = r
		case apis.Compute:
			r := apis.ResourceStatus{
				Reserved:     action.Status.Resources[resource.Name].Reserved - resource.ExpectedValue,
				Usage:        action.Status.Resources[resource.Name].Usage,
				UsageUnit:    action.Status.Resources[resource.Name].UsageUnit,
				ReservedUnit: action.Status.Resources[resource.Name].ReservedUnit,
			}
			action.Status.Resources[resource.Name] = r
		}
	}
	return nil
}
func RecoverResourceStatus(runtime *apis.Runtime, action *apis.Action) error {
	resources := runtime.Resources
	for _, resource := range resources {
		switch resource.Type {
		case apis.Storage:
			// TODO:更新usage 考虑unit
			r := apis.ResourceStatus{
				Reserved:     action.Status.Resources[resource.Name].Reserved + resource.ExpectedValue,
				Usage:        action.Status.Resources[resource.Name].Usage,
				UsageUnit:    action.Status.Resources[resource.Name].UsageUnit,
				ReservedUnit: action.Status.Resources[resource.Name].ReservedUnit,
			}
			action.Status.Resources[resource.Name] = r
		case apis.Network:
			r := apis.ResourceStatus{
				Reserved:     action.Status.Resources[resource.Name].Reserved + resource.ExpectedValue,
				Usage:        action.Status.Resources[resource.Name].Usage,
				UsageUnit:    action.Status.Resources[resource.Name].UsageUnit,
				ReservedUnit: action.Status.Resources[resource.Name].ReservedUnit,
			}
			action.Status.Resources[resource.Name] = r
		case apis.Compute:
			r := apis.ResourceStatus{
				Reserved:     action.Status.Resources[resource.Name].Reserved + resource.ExpectedValue,
				Usage:        action.Status.Resources[resource.Name].Usage,
				UsageUnit:    action.Status.Resources[resource.Name].UsageUnit,
				ReservedUnit: action.Status.Resources[resource.Name].ReservedUnit,
			}
			action.Status.Resources[resource.Name] = r
		}
	}
	return nil
}
