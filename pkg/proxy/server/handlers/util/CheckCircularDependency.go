package util

import (
	"errors"
	apis "hit.edu/framework/pkg/apis/cores"
)

// workflow->task->group->action->runtime

// CheckTaskCircularDependency task 循环依赖检查
func CheckTaskCircularDependency(ew apis.WorkflowSpec) error {
	// logs.Info("CheckCircularDependency")
	if len(ew.Tasks) > 0 {
		for _, task := range ew.Tasks {
			// 对于每个task检查
			for _, name := range task.Parents {
				if name == task.Name {
					// logs.Info("group circular dependency")
					return errors.New("task circular dependency")
				}
			}
			// 对于每个task里面的group检查
			err := CheckGroupCircularDependency(task)
			if err != nil {
				return err
			}

		}
	}

	return nil
}

// CheckGroupCircularDependency group 循环依赖检查
func CheckGroupCircularDependency(et apis.TaskSpec) error {
	// logs.Info("CheckGroupCircularDependency")
	if len(et.Groups) > 0 {
		for _, group := range et.Groups {
			// logs.Info("current group name", group.Name)
			// 对于每个group检查
			for _, name := range group.Parents {
				// logs.Info("dependency group name", name)
				if name == group.Name {
					// logs.Info("group circular dependency")
					return errors.New("group circular dependency")
				}
			}
			//对于每个group里面的action检查
			err := CheckActionCircularDependency(group)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CheckActionCircularDependency action 循环依赖检查
func CheckActionCircularDependency(eg apis.GroupSpec) error {
	// logs.Info("CheckActionCircularDependency")
	if len(eg.Actions) > 0 {
		for _, action := range eg.Actions {
			// 对于每个group检查
			for _, name := range action.Parents {
				if name == action.Name {
					// logs.Info("action circular dependency")
					return errors.New("action circular dependency")
				}
			}
			//对于每个action里面的runtime检查
			err := CheckRuntimeCircularDependency(action)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CheckRuntimeCircularDependency runtime 循环依赖检查
func CheckRuntimeCircularDependency(ea apis.ActionSpec) error {
	// logs.Info("CheckRuntimeCircularDependency")
	if len(ea.Runtimes) > 0 {
		for _, runtime := range ea.Runtimes {
			// 对于每个runtime检查
			for _, name := range runtime.Parents {
				if name == runtime.Name {
					// logs.Info("runtime circular dependency")
					return errors.New("runtime circular dependency")
				}
			}
		}
	}

	return nil
}
