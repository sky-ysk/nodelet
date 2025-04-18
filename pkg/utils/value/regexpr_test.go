package value

import (
	"fmt"
	"testing"
)

func TestRegExpr(t *testing.T) {
	comparor := NewRegExprComparor()
	kind := "Task"
	taskStr := "Task{T1}.Group{R1}.Status{phase}"
	name, parts, err := comparor.Match(kind, taskStr)
	if err == nil {
		fmt.Println("--------------")
		fmt.Println(kind)
		fmt.Println(taskStr)
		fmt.Println("----Results-----")
		fmt.Println(name)
		for _, part := range parts {
			fmt.Println(part)
		}
	}
	
	kind = "Task"
	taskStr = "Task{T1}.Group{R1}.Action{A1}.Status{phase}"
	name, parts, err = comparor.Match(kind, taskStr)
	if err == nil {
		fmt.Println("--------------")
		fmt.Println(kind)
		fmt.Println(taskStr)
		fmt.Println("----Results-----")
		fmt.Println(name)
		for _, part := range parts {
			fmt.Println(part)
		}
	}
	
	kind = "Group"
	taskStr = "Group{R1}.Action{A1}.Status{phase}"
	name, parts, err = comparor.Match(kind, taskStr)
	if err == nil {
		fmt.Println("--------------")
		fmt.Println(kind)
		fmt.Println(taskStr)
		fmt.Println("----Results-----")
		fmt.Println(name)
		for _, part := range parts {
			fmt.Println(part)
		}
	}
	
	kind = "Group"
	taskStr = "Group{R1}.Status{phase}"
	name, parts, err = comparor.Match(kind, taskStr)
	if err == nil {
		fmt.Println("--------------")
		fmt.Println(kind)
		fmt.Println(taskStr)
		fmt.Println("----Results-----")
		fmt.Println(name)
		for _, part := range parts {
			fmt.Println(part)
		}
	}
	
	kind = "Runtime"
	taskStr = "Runtime{R1}.Outputs{value}"
	name, parts, err = comparor.Match(kind, taskStr)
	if err == nil {
		fmt.Println("--------------")
		fmt.Println(kind)
		fmt.Println(taskStr)
		fmt.Println("----Results-----")
		fmt.Println(name)
		for _, part := range parts {
			fmt.Println(part)
		}
	}
	
	kind = "Runtime"
	taskStr = "Runtime{R2}.Status{phase}"
	name, parts, err = comparor.Match(kind, taskStr)
	if err == nil {
		fmt.Println("--------------")
		fmt.Println(kind)
		fmt.Println(taskStr)
		fmt.Println("----Results-----")
		fmt.Println(name)
		for _, part := range parts {
			fmt.Println(part)
		}
	}
	
}

func TestDeivceExpr(t *testing.T) {
	comparor := NewRegExprComparor()
	kind := "Device"
	deviceStr := "Device{Robot}.Ability{Move}.Service{Start}"
	
	name, parts, err := comparor.Match(kind, deviceStr)
	if err == nil {
		fmt.Println("--------------")
		fmt.Println(kind)
		fmt.Println(deviceStr)
		fmt.Println("----Results-----")
		fmt.Println(name)
		for _, part := range parts {
			fmt.Println(part)
		}
	}
}
