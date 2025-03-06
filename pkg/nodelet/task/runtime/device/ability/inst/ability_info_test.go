package inst

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetAbilityInstances(t *testing.T) {

	responseBody := `[
	{
		"id": "20d3ff60-8520-442f-b228-3d4ab25477d2",
		"kind": "AtomAbility",
		"metadata": {
		"labels": {},"name": "Robot.MonitorAbility"
	},
		"owner": {
		"abilityInstanceId": "00000000-0000-0000-0000-000000000000",
			"position": ""
	},
		"runInfo": {
		"lastConnect": 0,
			"lastUpdate": 0,
			"lifecycleState": "Inactive"
	},
		"sharers": null,
		"spec": {
		"abilityName": "Robot.MonitorAbility",
			"config": null,
			"package": "abstract.test.org",
			"position": "localhost",
			"priority": 0,
			"version": "0.1.0"
	},
		"status": null,
		"tag": {
		"parent": "abstract-monitor-1",
			"source": "AbstractAbility"
	}
	},
	{
		"id": "48780900-98d5-4332-abcf-e8996e3b599a",
		"kind": "AtomAbility",
		"metadata": {
		"labels": {},
		"name": "Fixed.MonitorAbility"
	},
		"owner": {
		"abilityInstanceId": "00000000-0000-0000-0000-000000000000",
			"position": ""
	},
		"runInfo": {
		"lastConnect": 0,
			"lastUpdate": 0,
			"lifecycleState": "Inactive"
	},
		"sharers": null,
		"spec": {
		"abilityName": "Fixed.MonitorAbility",
			"config": null,
			"package": "abstract.test.org",
			"position": "localhost","priority": 20,
			"version": "0.1.0"
	},
		"status": null,
		"tag": {
		"parent": "abstract-monitor-1",
			"source": "AbstractAbility"
	}
	},
	{
		"id": "9ef71d93-f575-4cdb-8e6b-ec30d3dfb72c",
		"kind": "AbstractAbility",
		"metadata": {
		"labels": {},
		"name": "abstract-monitor-1"
	},
		"owner": {
		"abilityInstanceId": "00000000-0000-0000-0000-000000000000",
			"position": ""
	},
		"runInfo": {
		"lastConnect": 0,
			"lastUpdate": 0,
			"lifecycleState": "Inactive"
	},
		"sharers": null,
		"spec": {
		"abilityName": "Abstract.MonitorAbility",
			"activityCondition": {
			"jq": ".status.covered | not"
		},
		"config": null,
			"package": "abstract.test.org",
			"position": "localhost",
			"subabilities": [
	{
	"abilityName": "Fixed.MonitorAbility",
	"config": null,
	"package": "abstract.test.org",
	"position": "localhost",
	"priority": 20,
	"version": "0.1.0"
	},
	{
	"abilityName": "Robot.MonitorAbility",
	"config": null,
	"package": "abstract.test.org",
	"position": "localhost","priority": 0,
	"version": "0.1.0"
	}
	],
	"version": "0.1.0"
	},
	"status": null,
	"subabilities": [
	{
	"id": "48780900-98d5-4332-abcf-e8996e3b599a",
	"position": "localhost"
	},
	{
	"id": "20d3ff60-8520-442f-b228-3d4ab25477d2",
	"position": "localhost"
	}
	],
	"tag": {
	"parent": "file",
	"source": "file"
	}
	}
	]`

	var abilityInstances []AbilityInstance
	err := json.Unmarshal([]byte(responseBody), &abilityInstances)
	if err != nil {
		fmt.Println(err)
	}
	if abilityInstances[1].Subabilities == nil {
		fmt.Println("this is a test !")
	}
	fmt.Println(abilityInstances[2].Subabilities[1].Position)
	fmt.Println(abilityInstances[2].Spec.SubAbilities[0].Package)
}
