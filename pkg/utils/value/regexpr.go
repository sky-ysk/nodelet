package value

import (
	"errors"
	"fmt"
	"regexp"
)

const (
	// TODO: 支持更多变量类型
	WorkflowExpr = `^Workflow{([^}]+)}\.(Status){([^}]+)}$`
	TaskExpr     = `^Task{([^}]+)}\.(Status){([^}]+)}$`
	// GroupExpr    = `^Group{([^}]+)}\.(Status){([^}]+)}$`
	GroupExpr = `^Group{([^}]+)}\.(Status|Name){([^}]*)}$`
	// ActionExpr   = `^Action{([^}]+)}\.(Status|Outputs){([^}]+)}$`
	ActionExpr = `^Action{([^}]+)}\.(Status|Outputs|Name){([^}]*)}$`
	// RuntimeExpr  = `^Runtime{([^}]+)}\.(Status|Outputs|Name){([^}]+)}$`
	RuntimeExpr = `^Runtime{([^}]+)}\.(Status|Outputs|Name){([^}]*)}$`

	// 用于父子节点之间的相互引用
	// 目前只支持父引用子
	WorkflowTaskExpr       = `^Workflow{([^}]+)}\.Task{([^}]+)}\.(Status){([^}]+)}$`
	TaskGroupExpr          = `^Task{([^}]+)}\.Group{([^}]+)}\.(Status){([^}]+)}$`
	GroupActionExpr        = `^Group{([^}]+)}\.Action{([^}]+)}\.(Status){([^}]+)}$`
	ActionRuntimeExpr      = `^Action{([^}]+)}\.Runtime{([^}]+)}\.(Status|Outputs|Name){([^}]*)}$`
	GroupActionRuntimeExpr = `^Group{([^}]+)}\.Action{([^}]+)}\.Runtime{([^}]+)}\.(Status|Outputs|Name){([^}]*)}$`

	// 绝对位置寻址
	// 只支持到Group层级
	// 系统中最高层级可能为Workflow、Task或Group
	WorkflowAbsoluteExpr = `^Workflow{([^}]+)}\.Task{([^}]+)}\.Group{([^}]+)}\.Action{([^}]+)}\.Runtime{([^}]+)}\.(Status|Outputs){([^}]+)}$`
	TaskAbsoluteExpr     = `^Task{([^}]+)}\.Group{([^}]+)}\.Action{([^}]+)}\.Runtime{([^}]+)}\.(Status|Outputs){([^}]+)}$`
	// GroupAbsoluteExpr    = `^Group{([^}]+)}\.Action{([^}]+)}\.Runtime{([^}]+)}\.(Status|Outputs){([^}]+)}$`

	// 设备寻址
	DeviceExpr = `^Device{([^}]+)}\.Ability{([^}]+)}\.Service{([^}]+)}`
)

var SupportedExprs = map[string]string{
	"WorkflowExpr":           WorkflowExpr,
	"TaskExpr":               TaskExpr,
	"GroupExpr":              GroupExpr,
	"ActionExpr":             ActionExpr,
	"RuntimeExpr":            RuntimeExpr,
	"WorkflowTaskExpr":       WorkflowTaskExpr,
	"TaskGroupExpr":          TaskGroupExpr,
	"GroupActionExpr":        GroupActionExpr,
	"ActionRuntimeExpr":      ActionRuntimeExpr,
	"GroupActionRuntimeExpr": GroupActionRuntimeExpr,
	"WorkflowAbsoluteExpr":   WorkflowAbsoluteExpr,
	"TaskAbsoluteExpr":       TaskAbsoluteExpr,
	// "GroupAbsoluteExpr":      GroupAbsoluteExpr,
	"DeviceExpr": DeviceExpr,
}

var WorkflowSupportedExprs = []string{
	"WorkflowExpr",
	"WorkflowTaskExpr",
	"WorkflowAbsoluteExpr",
	"TaskGroupExpr",
	"GroupAbsoluteExpr",
}

var TaskSupportedExprs = []string{
	"TaskExpr",
	"WorkflowTaskExpr",
	"TaskGroupExpr",
	"WorkflowAbsoluteExpr",
	"TaskGroupExpr",
	"GroupAbsoluteExpr",
}

var GroupSupportedExprs = []string{
	"GroupExpr",
	"TaskGroupExpr",
	"GroupActionExpr",
	"WorkflowAbsoluteExpr",
	"TaskGroupExpr",
	"GroupAbsoluteExpr",
	"GroupActionRuntimeExpr",
}

var ActionSupportedExprs = []string{
	"ActionExpr",
	"GroupActionExpr",
	"ActionRuntimeExpr",
	"WorkflowAbsoluteExpr",
	"TaskGroupExpr",
	"GroupAbsoluteExpr",
}

var RuntimeSupportedExprs = []string{
	"RuntimeExpr",
	"ActionRuntimeExpr",
	"WorkflowAbsoluteExpr",
	"TaskGroupExpr",
	"GroupAbsoluteExpr",
	"GroupExpr",
	"ActionExpr",
}

var DeviceSupportedExprs = []string{
	"DeviceExpr",
}

type RegExpr struct {
	expr map[string]*regexp.Regexp
}

type RegExprComparor struct {
	// 构造不同的Expr对比器
	// Key为值类型，如Task、Device等
	// Value为类型对应的对比器
	exprs          map[string]*RegExpr
	supportedExprs *map[string]*regexp.Regexp
}

func NewRegExprComparor() *RegExprComparor {
	exprMap, err := CreateSupportedExpr()
	if err != nil {
		return nil
	}

	regExprComparor := &RegExprComparor{
		exprs:          make(map[string]*RegExpr),
		supportedExprs: exprMap,
	}

	kind := "Workflow"
	workflowRegExpr, err := CreateRegExpr(WorkflowSupportedExprs, exprMap)
	if err != nil {
		return nil
	}
	regExprComparor.exprs[kind] = workflowRegExpr

	kind = "Task"
	taskRegExpr, err := CreateRegExpr(TaskSupportedExprs, exprMap)
	if err != nil {
		return nil
	}
	regExprComparor.exprs[kind] = taskRegExpr

	kind = "Group"
	groupRegExpr, err := CreateRegExpr(GroupSupportedExprs, exprMap)
	if err != nil {
		return nil
	}
	regExprComparor.exprs[kind] = groupRegExpr

	kind = "Action"
	actionRegExpr, err := CreateRegExpr(ActionSupportedExprs, exprMap)
	if err != nil {
		return nil
	}
	regExprComparor.exprs[kind] = actionRegExpr

	kind = "Runtime"
	runtimeRegExpr, err := CreateRegExpr(RuntimeSupportedExprs, exprMap)
	if err != nil {
		return nil
	}
	regExprComparor.exprs[kind] = runtimeRegExpr

	kind = "Device"
	deviceRegExpr, err := CreateRegExpr(DeviceSupportedExprs, exprMap)
	if err != nil {
		return nil
	}
	regExprComparor.exprs[kind] = deviceRegExpr

	kind = "Value"
	return regExprComparor
}

func CreateSupportedExpr() (*map[string]*regexp.Regexp, error) {
	expr := make(map[string]*regexp.Regexp)
	for name, exprStr := range SupportedExprs {
		re, err := regexp.Compile(exprStr)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		expr[name] = re
	}
	return &expr, nil
}

func CreateRegExpr(expr []string, supportedExpr *map[string]*regexp.Regexp) (*RegExpr, error) {
	regExpr := RegExpr{
		expr: make(map[string]*regexp.Regexp),
	}

	// TODO: 空值检查
	for _, exprStr := range expr {
		regExpr.expr[exprStr] = (*supportedExpr)[exprStr]
	}
	return &regExpr, nil
}

func (r *RegExprComparor) Match(kind string, target string) (string, []string, error) {
	regExpr := r.exprs[kind]
	for name, expr := range regExpr.expr {
		if expr == nil {
			continue
		}
		parts := expr.FindStringSubmatch(target)
		if parts != nil {
			return name, parts, nil
		}
	}
	return "", nil, errors.New("invalid expression")
}
