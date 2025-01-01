// 插件模块接口定义
package framework

import (
	"context"
	"errors"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/scheduler/apis/config"
	"strings"
)

// TODO: 传入资源信息
// TODO: 传入任务在插件上的运行状态
// 插件返回的状态
type Code int

// This list should be exactly the same as the codes iota defined above in the same order.
var codes = []string{"Success", "Error", "Unschedulable", "UnschedulableAndUnresolvable", "Wait", "Skip", "Pending"}

// 插件运行状态定义
const (
	Success Code = iota
	Error
	Unschedulable
	UnschedulableAndUnresolvable
	Wait
	Skip
	Pending
)

// Plugin的运行结果
type Status struct {
	code    Code
	reasons []string
	err     error
	plugin  string
}

func (s *Status) WithError(err error) *Status {
	s.err = err
	return s
}

func (s *Status) Code() Code {
	if s == nil {
		return Success
	}
	return s.code
}

func (s *Status) Message() string {
	if s == nil {
		return ""
	}
	return strings.Join(s.Reasons(), ", ")
}

func (s *Status) SetPlugin(plugin string) {
	s.plugin = plugin
}

func (s *Status) WithPlugin(plugin string) *Status {
	s.SetPlugin(plugin)
	return s
}

func (s *Status) Plugin() string {
	return s.plugin
}

func (s *Status) Reasons() []string {
	if s.err != nil {
		return append([]string{s.err.Error()}, s.reasons...)
	}
	return s.reasons
}

func (s *Status) AppendReason(reason string) {
	s.reasons = append(s.reasons, reason)
}

func (s *Status) IsSuccess() bool {
	return s.Code() == Success
}

func (s *Status) IsWait() bool {
	return s.Code() == Wait
}

func (s *Status) IsSkip() bool {
	return s.Code() == Skip
}

func (s *Status) IsRejected() bool {
	code := s.Code()
	return code == Unschedulable || code == Pending
}

func (s *Status) AsError() error {
	if s.IsSuccess() || s.IsWait() || s.IsSkip() {
		return nil
	}
	if s.err != nil {
		return s.err
	}
	return errors.New(s.Message())
}

// func (s *Status) Equal(x *Status) bool {
// 	if s == nil || x == nil {
// 		return s.IsSuccess() && x.IsSuccess()
// 	}
// 	if s.code != x.code {
// 		return false
// 	}
// 	if !cmp.Equal(s.err, x.err, cmpopts.EquateErrors()) {
// 		return false
// 	}
// 	if !cmp.Equal(s.reasons, x.reasons) {
// 		return false
// 	}
// 	return cmp.Equal(s.plugin, x.plugin)
// }

func (s *Status) String() string {
	return s.Message()
}

func NewStatus(code Code, reasons ...string) *Status {
	s := &Status{
		code:    code,
		reasons: reasons,
	}
	return s
}

func AsStatus(err error) *Status {
	if err == nil {
		return nil
	}
	return &Status{
		code: Error,
		err:  err,
	}
}

// 所有插件类型的接口父类
type Plugin interface {
	Name() string
}

// 插件类型定义
// 调度链包含几种插件类型
// 1. 预调度插件 PreSchedulePlugin, 用于在调度之前执行，检查资源是否满足要求
// 2. 生成插件 GeneratorPlugin, 用于根据当前资源情况动态生成任务调度流程，主要用于端侧调度
// 3. 调度插件 SchedulePlugin, 用于执行调度操作
// 4. 后调度插件 PostSchedulePlugin, 用于在调度之后执行，清理资源等操作
// TODO: 定义插件的Extension

type FilterPlugin interface {
	Plugin
	// 在任务进度可调度队列前执行
	Filter(ctx context.Context, group *apis.Group, node *config.NodeInfo) *Status
}

type GeneratorPlugin interface {
	Plugin
	// 生成任务执行流程
	Generate(ctx context.Context, group *apis.Group) *Status
}

type ScorePlugin interface {
	Plugin
	// 执行任务调度
	Score(ctx context.Context, group *apis.Group, nodeName string) (int64, *Status)
}

type BindPlugin interface {
	Plugin
	Bind(ctx context.Context, state *CycleState, group *apis.Group, nodeName string) *Status
}

//TODO: 定义插件的运行结果

type Framework interface {
	Handle

	PercentageOfNodesToScore() *int32

	HasScorePlugins() bool

	HasFilterPlugins() bool
	// TODO: 插件运行结果相关结构定义、方法
	Close() error
	// TODO 一致性
	RunReservePluginsReserve(ctx context.Context, state *CycleState, group *apis.Group, nodeName string) *Status

	// RunReservePluginsUnreserve runs the Unreserve method of the set of
	// configured Reserve plugins.
	RunReservePluginsUnreserve(ctx context.Context, state *CycleState, group *apis.Group, nodeName string)
	//TODO 资源检查
	RunBindPlugins(ctx context.Context, state *CycleState, group *apis.Group, nodeName string) *Status
}

type Handle interface {
	PluginsRunner
}

// TODO: 插件运行结果相关结构定义、方法
type PluginsRunner interface {
	//
	//[]nodes Filter([]nodes)
	//RunPreSchedulePlugins(ctx context.Context, group *workflow.Group) *Status

	RunFilterPlugins(ctx context.Context, nodeInfo *config.NodeInfo, state *CycleState, group *apis.Group) *Status

	RunScorePlugins(ctx context.Context, state *CycleState, group *apis.Group, infos []*config.NodeInfo) ([]NodePluginScores, *Status)

	//
	RunGeneratorPlugins(ctx context.Context, group *apis.Group) *Status
	//
	//RunSchedulePlugins(ctx context.Context, group *workflow.Group) *Status
	//
	//RunPostSchedulePlugins(ctx context.Context, group *workflow.Group) *Status

	//Sort
}

type NominatingMode int

const (
	ModeNoop NominatingMode = iota
	ModeOverride
)

type NominatingInfo struct {
	NominatedNodeName string
	NominatingMode    NominatingMode
}

type NodePluginScores struct {
	// Name is node name.
	Name string
	// Scores is scores from plugins and extenders.
	Scores []PluginScore
	// TotalScore is the total score in Scores.
	TotalScore int64
}

// PluginScore is a struct with plugin/extender name and score.
type PluginScore struct {
	// Name is the name of plugin or extender.
	Name  string
	Score int64
}

type NodeScoreList []NodeScore

// NodeScore is a struct with node name and score.
type NodeScore struct {
	Name  string
	Score int64
}

type LessFunc func(groupInfo1, groupInfo2 *config.QueuedGroupInfo) bool
