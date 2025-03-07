package inst

import (
	"fmt"
	"hit.edu/framework/pkg/component-base/logs"
	"testing"
)

func TestGetAbilityInstances(t *testing.T) {
	moduleName := "testModule"
	logs.Init(moduleName)
	logs.Infof("[test] testing run.....\n")
	var url string

	var abilityInstances []AbilityInstance

	abilityInstances, err := GetAbilityInstances(url)
	if err != nil {
		fmt.Println(err)
	}
	if abilityInstances[1].Subabilities == nil {
		fmt.Println("this is a test !")
	}
	fmt.Println(abilityInstances[2].Subabilities[1].Position)
	fmt.Println(abilityInstances[2].Spec.SubAbilities[0].Package)
}
