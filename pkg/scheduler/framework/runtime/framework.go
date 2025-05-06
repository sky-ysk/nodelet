package runtime

import (
	"context"
	"fmt"
	apis "hit.edu/framework/pkg/apis/cores"
	"hit.edu/framework/pkg/scheduler/utils"
	"time"

	"hit.edu/framework/pkg/component-base/logs"
	"hit.edu/framework/pkg/scheduler/apis/config"
	"hit.edu/framework/pkg/scheduler/framework"
)

const (
	// 设置超时时间
	maxTimeout = 15 * time.Minute
)

// 初始化并运行所有的插件，Framework的具体实现类
type frameworkImpl struct {
	registry Registry

	filterPlugins    []framework.FilterPlugin
	generatorPlugins []framework.GeneratorPlugin
	scorePlugins     []framework.ScorePlugin
	bindPlugins      []framework.BindPlugin

	// 所有的插件
	pluginsMap map[string]framework.Plugin

	profileName string

	parallelizer utils.Parallelizer

	//TODO 调度器权重暂时不做
	//scorePluginWeight map[string]int
}

func (f *frameworkImpl) GetDTSPlugin() framework.ScorePlugin {
	for _, plugin := range f.scorePlugins {
		if plugin.Name() == "ScorePluginForDuBoyu" {
			return plugin
		}
	}
	return nil
}

func (f *frameworkImpl) RunBindPlugins(ctx context.Context, state *framework.CycleState, group *apis.Group, nodeName string) (status *framework.Status) {
	if len(f.bindPlugins) == 0 {
		logs.Error("no bind plugins")
		return framework.NewStatus(framework.Skip, "no bind plugins")
	}
	for _, pl := range f.bindPlugins {
		ctx := ctx
		status = f.runBindPlugin(ctx, pl, state, group, nodeName)
		if status.IsSkip() {
			continue
		}
		if !status.IsSuccess() {
			if status.IsRejected() {
				logs.Info("Group rejected by Bind plugin", "Group", group.Name, "node", nodeName, "plugin", pl.Name(), "status", status.Message())
				status.SetPlugin(pl.Name())
				return status
			}
			err := status.AsError()
			logs.Error(err, "Plugin Failed", "plugin", pl.Name(), "Group", group.Name, "node", nodeName)
			return framework.AsStatus(fmt.Errorf("running Bind plugin %q: %w", pl.Name(), err))
		}
		return status
	}
	return status
}

func (f *frameworkImpl) runBindPlugin(ctx context.Context, bp framework.BindPlugin, state *framework.CycleState, group *apis.Group, nodeName string) *framework.Status {
	status := bp.Bind(ctx, state, group, nodeName)
	return status
}

func (f *frameworkImpl) RunReservePluginsReserve(ctx context.Context, state *framework.CycleState, group *apis.Group, nodeName string) *framework.Status {
	ret := framework.NewStatus(framework.Success, "")
	return ret
}

func (f *frameworkImpl) RunReservePluginsUnreserve(ctx context.Context, state *framework.CycleState, group *apis.Group, nodeName string) {
	return
}

func (f *frameworkImpl) PercentageOfNodesToScore() *int32 {
	var percentage int32 = 1
	return &percentage
}

func (f *frameworkImpl) HasScorePlugins() bool {
	return len(f.scorePlugins) > 0
}

func (f *frameworkImpl) HasFilterPlugins() bool {
	return len(f.filterPlugins) > 0
}

func (f *frameworkImpl) RunFilterPlugins(ctx context.Context, nodeInfo *config.NodeInfo, state *framework.CycleState,
	group *apis.Group) *framework.Status {

	for _, pl := range f.filterPlugins {
		if state.SkipFilterPlugins.Has(pl.Name()) {
			continue
		}
		ctx := ctx
		if status := f.runFilterPlugin(ctx, pl, state, group, nodeInfo); !status.IsSuccess() {
			if !status.IsRejected() {
				// Filter plugins are not supposed to return any status other than
				// Success or Unschedulable.
				status = framework.AsStatus(fmt.Errorf("running %q filter plugin: %w", pl.Name(), status.AsError()))
			}
			status.SetPlugin(pl.Name())
			return status
		}
	}
	return nil
}

func (f *frameworkImpl) runFilterPlugin(ctx context.Context, pl framework.FilterPlugin, state *framework.CycleState,
	group *apis.Group, nodeInfo *config.NodeInfo) *framework.Status {
	status := pl.Filter(ctx, group, nodeInfo)
	return status
}

// TODO 插件筛选 @linbohai
func (f *frameworkImpl) RunScorePlugins(ctx context.Context, state *framework.CycleState, group *apis.Group,
	infos []*config.NodeInfo) ([]framework.NodePluginScores, *framework.Status) {

	allNodePluginScores := make([]framework.NodePluginScores, len(infos))
	numPlugins := len(f.scorePlugins)
	plugins := make([]framework.ScorePlugin, 0, numPlugins)
	pluginToNodeScores := make(map[string]framework.NodeScoreList, numPlugins)
	for _, pl := range f.scorePlugins {
		if state.SkipScorePlugins.Has(pl.Name()) {
			continue
		}
		if len(group.Spec.SkipScorePlugins) != 0 {
			skip := false
			for _, skipPlugin := range group.Spec.SkipScorePlugins {
				if skipPlugin == pl.Name() {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
		}
		plugins = append(plugins, pl)
		pluginToNodeScores[pl.Name()] = make(framework.NodeScoreList, len(infos))
	}
	//errCh := utils.NewErrorChannel()
	//TODO @linbohai  错误处理
	if len(plugins) > 0 {
		// Run Score method for each node in parallel.
		f.Parallelizer().Until(ctx, len(infos), func(index int) {
			nodeName := infos[index].Node().Name
			for _, pl := range plugins {

				ctx := ctx
				s, status := f.runScorePlugin(ctx, pl, state, group, nodeName)
				if !status.IsSuccess() {
					//err := fmt.Errorf("plugin %q failed with: %w", pl.Name(), status.AsError())
					//errCh.SendErrorWithCancel(err, cancel)
					logs.Errorf("plugin %q failed with: %s , node %s ", pl.Name(), status.AsError().Error(), nodeName)
					logs.Errorf("plugin %q fail on node %s , use default score", pl.Name(), nodeName)
					s = 5
				}
				pluginToNodeScores[pl.Name()][index] = framework.NodeScore{
					Name:  nodeName,
					Score: s,
				}
			}
		})
		//if err := errCh.ReceiveError(); err != nil {
		//	return nil, framework.AsStatus(fmt.Errorf("running Score plugins: %w", err))
		//}
	}

	//TODO  Run NormalizeScore method for each ScorePlugin in parallel.

	// Apply score weight for each ScorePlugin in parallel,
	// and then, build allNodePluginScores.
	f.Parallelizer().Until(ctx, len(infos), func(index int) {
		nodePluginScores := framework.NodePluginScores{
			Name:   infos[index].Node().Name,
			Scores: make([]framework.PluginScore, len(plugins)),
		}

		for i, pl := range plugins {
			//weight := f.scorePluginWeight[pl.Name()]
			nodeScoreList := pluginToNodeScores[pl.Name()]
			score := nodeScoreList[index].Score

			//if score > framework.MaxNodeScore || score < framework.MinNodeScore {
			//	err := fmt.Errorf("plugin %q returns an invalid score %v, it should in the range of [%v, %v] after normalizing", pl.Name(), score, framework.MinNodeScore, framework.MaxNodeScore)
			//	errCh.SendErrorWithCancel(err, cancel)
			//	return
			//}
			weightedScore := score * 1
			nodePluginScores.Scores[i] = framework.PluginScore{
				Name:  pl.Name(),
				Score: weightedScore,
			}
			nodePluginScores.TotalScore += weightedScore
		}
		allNodePluginScores[index] = nodePluginScores
	})
	//if err := errCh.ReceiveError(); err != nil {
	//	return nil, framework.AsStatus(fmt.Errorf("applying score defaultWeights on Score plugins: %w", err))
	//}

	return allNodePluginScores, nil
}

func (f *frameworkImpl) RunGeneratorPlugins(ctx context.Context, group *apis.Group) *framework.Status {
	return nil
}

func (f *frameworkImpl) Close() error {
	return nil
}

// Parallelizer returns a parallelizer holding parallelism for scheduler.
func (f *frameworkImpl) Parallelizer() utils.Parallelizer {
	return f.parallelizer
}

//TODO: 挂载点

type frameworkOptions struct {
	// TODO: 设置event Handler
	// TODO: 配置
}

type Option func(*frameworkOptions)

func defaultFrameworkOptions() frameworkOptions {
	return frameworkOptions{}
}

var _ framework.Framework = &frameworkImpl{}

func NewDefaultFramework(ctx context.Context, r Registry, name string) (framework.Framework, error) {
	//options := defaultFrameworkOptions()
	//for _, opt := range opts {
	//	opt(&options)
	//}

	f := &frameworkImpl{
		registry:     r,
		parallelizer: utils.NewParallelizer(5),
	}

	//if profile == nil {
	//	return f, nil
	//}

	f.profileName = name
	//if profile.Plugins == nil {
	//	return f, nil
	//}

	// outputProfile := config.SchedulerProfile{
	// 	SchedulerName: f.profileName,
	// 	Plugins:       profile.Plugins,
	// }

	// 配置需要的插件
	f.pluginsMap = make(map[string]framework.Plugin)
	f.bindPlugins = make([]framework.BindPlugin, 0)
	f.scorePlugins = make([]framework.ScorePlugin, 0)
	f.filterPlugins = make([]framework.FilterPlugin, 0)
	for name, factory := range r {
		p, err := factory(ctx, f)
		if err != nil {
			logs.Errorf("initializing plugin %q: %s", name, err.Error())
			continue
			//return nil, fmt.Errorf("initializing plugin %q: %w", name, err)
		}
		f.pluginsMap[name] = p
		//在这里加入不同插件队列
		if bp, ok := p.(framework.BindPlugin); ok {
			f.bindPlugins = append(f.bindPlugins, bp)
		} else if fp, ok := p.(framework.FilterPlugin); ok {
			f.filterPlugins = append(f.filterPlugins, fp)
		} else if sp, ok := p.(framework.ScorePlugin); ok {
			f.scorePlugins = append(f.scorePlugins, sp)
		} else {
			fmt.Println("add plugins fail!!")
		}
	}

	// 配置需要的挂载点

	logs.Info("the scheduler starts to work with those plugins") //TODO: 列出所有的插件
	msg := fmt.Sprintf("score plugins num %d", len(f.scorePlugins))
	logs.Info(msg)
	msg1 := fmt.Sprintf("filter plugins num %d", len(f.filterPlugins))
	logs.Info(msg1)
	msg2 := fmt.Sprintf("bind plugins num %d", len(f.bindPlugins))
	logs.Info(msg2)
	fmt.Println(msg1)
	fmt.Println(msg2)
	fmt.Println(msg)
	return f, nil
}

func (f *frameworkImpl) runScorePlugin(ctx context.Context, pl framework.ScorePlugin, state *framework.CycleState,
	group *apis.Group, nodeName string) (int64, *framework.Status) {
	return pl.Score(ctx, group, nodeName)
}
