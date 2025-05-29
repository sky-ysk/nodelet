package main

import (
	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/nodelet/device"
)

func main() {
	moduleName := "testModule"
	logs.Init(moduleName)
	logs.Infof("[test] testing run.....\n")
	// 乐聚
	url1 := "http://192.168.8.165:8080"
	// 星海图
	url2 := "http://192.168.8.197:8080"

	var name1 string = "GrabBall.Leju.Guochuang"       // 更改能力名字
	var name2 string = "DetectPosition.Leju.Guochuang" // 更改能力名字

	var name3 string = "DetectPosition.Galaxea.Guochuang" // 更改能力名字
	var name4 string = "GrabBall.Galaxea.Guochuang"       // 更改能力名字

	nameList1 := []string{name1, name2}
	nameList2 := []string{name3, name4}
	for _, name := range nameList1 {
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
	for _, name := range nameList2 {
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
