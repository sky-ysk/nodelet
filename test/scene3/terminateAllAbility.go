package main

import (
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/device"
)

func main() {
	moduleName := "testModule"
	logs.Init(moduleName)
	logs.Infof("[test] testing run.....\n")
	//长春现场机器人地址
	//初检设备
	//url1 := "http://192.168.1.237:8080"
	////复检设备
	//url2 := "http://192.168.1.234:8080"

	var url1 string = "http://172.130.0.61:8080" // 更改ip和端口
	var url2 string = "http://172.130.0.59:8080"
	var name1 string = "GrabObject.Leju.801"   // 更改能力名字
	var name2 string = "PutObject.Leju.801"    // 更改能力名字
	var name3 string = "Turn.Leju.801"         // 更改能力名字
	var name4 string = "DetectObject.Leju.801" // 更改能力名字

	nameList := []string{name1, name2, name3, name4}
	for _, name := range nameList {
		abilityManager := device.NewAbilityManager(url1, name)
		err := abilityManager.BindUUID()
		if err != nil {
			logs.Errorf("bind uuid fail")
		}
		err = abilityManager.TerminateAbility()
		if err != nil {
			logs.Errorf("terminate ability fail")
		}
	}
	for _, name := range nameList {
		abilityManager := device.NewAbilityManager(url2, name)
		err := abilityManager.BindUUID()
		if err != nil {
			logs.Errorf("bind uuid fail")
		}
		err = abilityManager.TerminateAbility()
		if err != nil {
			logs.Errorf("terminate ability fail")
		}
	}

}
