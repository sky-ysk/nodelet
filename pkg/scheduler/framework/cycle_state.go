package framework

import (
	apis "hit.edu/framework/pkg/apis/cores"
	"k8s.io/apimachinery/pkg/util/sets"
	"sync"
)

// CycleState provides a mechanism for plugins to store and retrieve arbitrary data.
// StateData stored by one plugin can be read, altered, or deleted by another plugin.
// CycleState does not provide any data protection, as all plugins are assumed to be
// trusted.
// Note: CycleState uses a sync.Map to back the storage, because it is thread safe. It's aimed to optimize for the "write once and read many times" scenarios.
// It is the recommended pattern used in all in-tree plugins - plugin-specific state is written once in PreFilter/PreScore and afterward read many times in Filter/Score.
type CycleState struct {
	// storage is keyed with StateKey, and valued with StateData.
	storage sync.Map
	// if recordPluginMetrics is true, metrics.PluginExecutionDuration will be recorded for this cycle.
	recordPluginMetrics bool
	// SkipFilterPlugins are plugins that will be skipped in the Filter extension point.
	SkipFilterPlugins sets.Set[string]
	// SkipScorePlugins are plugins that will be skipped in the Score extension point.
	SkipScorePlugins sets.Set[string]
}

// NewCycleState initializes a new CycleState and returns its pointer.
func NewCycleState(group *apis.Group) *CycleState {
	skipScorePlugins := sets.New[string]()
	skipFilterPlugins := sets.New[string]()
	for _, plName := range group.Spec.SkipFilterPlugins {
		skipFilterPlugins.Insert(plName)
	}
	for _, plName := range group.Spec.SkipScorePlugins {
		skipScorePlugins.Insert(plName)
	}
	return &CycleState{
		SkipFilterPlugins: skipFilterPlugins,
		SkipScorePlugins:  skipScorePlugins,
	}
}
