package apis

import (
	"hit.edu/framework/pkg/apimachinery/runtime"
	"hit.edu/framework/pkg/apis/meta"
)

type WorkflowList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Workflow
}
type TaskList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Task
}
type GroupList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Group
}
type ActionList struct {
	meta.TypeMeta
	meta.ListMeta
	Items []Action
}

func (in *Time) DeepCopyInto(out *Time) {
	*out = *in
}

func (in *Workflow) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
func (in *Workflow) DeepCopy() *Workflow {
	if in == nil {
		return nil
	}
	out := new(Workflow)
	in.DeepCopyInto(out)
	return out
}
func (in *Workflow) DeepCopyInto(out *Workflow) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *WorkflowList) DeepCopyInto(out *WorkflowList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Workflow, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *WorkflowList) DeepCopy() *WorkflowList {
	if in == nil {
		return nil
	}
	out := new(WorkflowList)
	in.DeepCopyInto(out)
	return out
}

func (in *WorkflowList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *Description) DeepCopyInto(out *Description) {
	*out = *in
	if in.Label != nil {
		out.Label = make([]string, len(in.Label))
		copy(out.Label, in.Label)
	}
}

func (in *WorkflowSpec) DeepCopyInto(out *WorkflowSpec) {
	*out = *in
	in.Desc.DeepCopyInto(&out.Desc)
	if in.Tasks != nil {
		in, out := &in.Tasks, &out.Tasks
		*out = make([]Task, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}
func (in *WorkflowStatus) DeepCopyInto(out *WorkflowStatus) {
	*out = *in
	in.StartAt.DeepCopyInto(&out.StartAt)
	in.FinishAt.DeepCopyInto(&out.FinishAt)
	in.LastTime.DeepCopyInto(&out.LastTime)
	if in.TaskStatus != nil {
		in, out := &in.TaskStatus, &out.TaskStatus
		*out = make([]TaskStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}
func (in *Task) DeepCopyInto(out *Task) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}
func (in *Task) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
func (in *Task) DeepCopy() *Task {
	if in == nil {
		return nil
	}
	out := new(Task)
	in.DeepCopyInto(out)
	return out
}

func (in *TaskList) DeepCopyInto(out *TaskList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Task, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *TaskList) DeepCopy() *TaskList {
	if in == nil {
		return nil
	}
	out := new(TaskList)
	in.DeepCopyInto(out)
	return out
}

func (in *TaskList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// TaskSpec 结构体的 DeepCopyInto 方法
func (in *TaskSpec) DeepCopyInto(out *TaskSpec) {
	*out = *in
	if in.Parents != nil {
		out.Parents = make([]string, len(in.Parents))
		copy(out.Parents, in.Parents)
	}
	in.Desc.DeepCopyInto(&out.Desc)
	in.Conditions.DeepCopyInto(&out.Conditions)
	if in.Groups != nil {
		out.Groups = make([]Group, len(in.Groups))
		for i := range in.Groups {
			in.Groups[i].DeepCopyInto(&out.Groups[i])
		}
	}
}

// TaskStatus 结构体的 DeepCopyInto 方法
func (in *TaskStatus) DeepCopyInto(out *TaskStatus) {
	*out = *in
	in.Belongs.DeepCopyInto(&out.Belongs)
	in.StartAt.DeepCopyInto(&out.StartAt)
	in.FinishAt.DeepCopyInto(&out.FinishAt)
	in.LastTime.DeepCopyInto(&out.LastTime)
	if in.GroupStatus != nil {
		out.GroupStatus = make([]GroupStatus, len(in.GroupStatus))
		for i := range in.GroupStatus {
			in.GroupStatus[i].DeepCopyInto(&out.GroupStatus[i])
		}
	}
}

func (in *IDRef) DeepCopyInto(out *IDRef) {
	*out = *in
}
func (in *Conditions) DeepCopyInto(out *Conditions) {
	*out = *in
	if in.Formulas != nil {
		out.Formulas = make([]ConditionFormula, len(in.Formulas))
		for i := range in.Formulas {
			in.Formulas[i].DeepCopyInto(&out.Formulas[i])
		}
	}
}
func (in *ConditionFormula) DeepCopyInto(out *ConditionFormula) {
	*out = *in
	// 深度复制 LeftValue 和 RightValue
	in.LeftValue.DeepCopyInto(&out.LeftValue)
	in.RightValue.DeepCopyInto(&out.RightValue)
}
func (in *ConditionValue) DeepCopyInto(out *ConditionValue) {
	*out = *in
}

// Group 结构体的 DeepCopyInto 方法
func (in *Group) DeepCopyInto(out *Group) {
	*out = *in
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}
func (in *Group) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
func (in *Group) DeepCopy() *Group {
	if in == nil {
		return nil
	}
	out := new(Group)
	in.DeepCopyInto(out)
	return out
}
func (in *GroupList) DeepCopyInto(out *GroupList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Group, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *GroupList) DeepCopy() *GroupList {
	if in == nil {
		return nil
	}
	out := new(GroupList)
	in.DeepCopyInto(out)
	return out
}

func (in *GroupList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// GroupSpec 结构体的 DeepCopyInto 方法
func (in *GroupSpec) DeepCopyInto(out *GroupSpec) {
	*out = *in
	if in.Parents != nil {
		out.Parents = make([]string, len(in.Parents))
		copy(out.Parents, in.Parents)
	}
	in.Desc.DeepCopyInto(&out.Desc)
	in.Conditions.DeepCopyInto(&out.Conditions)
	if in.Actions != nil {
		out.Actions = make([]Action, len(in.Actions))
		for i := range in.Actions {
			in.Actions[i].DeepCopyInto(&out.Actions[i])
		}
	}
}

func (in *GroupStatus) DeepCopyInto(out *GroupStatus) {
	*out = *in
	in.Belongs.DeepCopyInto(&out.Belongs)
	if in.ActionStatus != nil {
		out.ActionStatus = make([]ActionStatus, len(in.ActionStatus))
		copy(out.ActionStatus, in.ActionStatus)
	}
	in.StartAt.DeepCopyInto(&out.StartAt)
	in.FinishAt.DeepCopyInto(&out.FinishAt)
	in.LastTime.DeepCopyInto(&out.LastTime)
}

// ActionSpec 结构体的 DeepCopyInto 方法
func (in *ActionSpec) DeepCopyInto(out *ActionSpec) {
	*out = *in
	if in.Parents != nil {
		out.Parents = make([]string, len(in.Parents))
		copy(out.Parents, in.Parents)
	}
	in.Desc.DeepCopyInto(&out.Desc)
	in.Conditions.DeepCopyInto(&out.Conditions)
	if in.Runtimes != nil {
		out.Runtimes = make([]Runtime, len(in.Runtimes))
		for i := range in.Runtimes {
			in.Runtimes[i].DeepCopyInto(&out.Runtimes[i])
		}
	}
}
func (in *ActionStatus) DeepCopyInto(out *ActionStatus) {
	*out = *in
	in.Belongs.DeepCopyInto(&out.Belongs)
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make([]ResourceStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Devices != nil {
		in, out := &in.Devices, &out.Devices
		*out = make([]DeviceStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}

	if in.Data != nil {
		in, out := &in.Data, &out.Data
		*out = make([]DataStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}

	if in.Scenes != nil {
		in, out := &in.Scenes, &out.Scenes
		*out = make([]SceneStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}

	if in.RuntimeStatus != nil {
		in, out := &in.RuntimeStatus, &out.RuntimeStatus
		*out = make([]RuntimeStatus, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}

	if in.Results != nil {
		in, out := &in.Results, &out.Results
		*out = make([]Result, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.StartAt.DeepCopyInto(&out.StartAt)
	in.FinishAt.DeepCopyInto(&out.FinishAt)
	in.LastTime.DeepCopyInto(&out.LastTime)
}

func (in *DataStatus) DeepCopyInto(out *DataStatus) {
	*out = *in
}
func (in *ResourceStatus) DeepCopyInto(out *ResourceStatus) {
	*out = *in
}
func (in *DeviceStatus) DeepCopyInto(out *DeviceStatus) {
	*out = *in
}
func (in *SceneStatus) DeepCopyInto(out *SceneStatus) {
	*out = *in
}
func (in *RuntimeStatus) DeepCopyInto(out *RuntimeStatus) {
	*out = *in
}

// Runtime 结构体的 DeepCopyInto 方法
func (in *Runtime) DeepCopyInto(out *Runtime) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = make([]ResourceSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}

	if in.Devices != nil {
		in, out := &in.Devices, &out.Devices
		*out = make([]DeviceSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}

	if in.Data != nil {
		in, out := &in.Data, &out.Data
		*out = make([]DataSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}

	if in.Scenes != nil {
		in, out := &in.Scenes, &out.Scenes
		*out = make([]SceneSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}

	if in.Command != nil {
		in, out := &in.Command, &out.Command
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Args != nil {
		in, out := &in.Args, &out.Args
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.EnvVar != nil {
		in, out := &in.EnvVar, &out.EnvVar
		*out = make([]EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Inputs.DeepCopyInto(&out.Inputs)
	in.Outputs.DeepCopyInto(&out.Outputs)
}

func (in *DataSpec) DeepCopyInto(out *DataSpec) {
	*out = *in
}
func (in *ResourceSpec) DeepCopyInto(out *ResourceSpec) {
	*out = *in
}
func (in *DeviceSpec) DeepCopyInto(out *DeviceSpec) {
	*out = *in
}
func (in *SceneSpec) DeepCopyInto(out *SceneSpec) {
	*out = *in
}
func (in *EnvVar) DeepCopyInto(out *EnvVar) {
	*out = *in
}

func (in *Input) DeepCopyInto(out *Input) {
	*out = *in
}
func (in *Output) DeepCopyInto(out *Output) {
	*out = *in
}
func (in *Result) DeepCopyInto(out *Result) {
	*out = *in
	in.Belongs.DeepCopyInto(&out.Belongs)
}

// Event相关结构体
func (in *Event) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
func (in *Event) DeepCopy() *Event {
	if in == nil {
		return nil
	}
	out := new(Event)
	in.DeepCopyInto(out)
	return out
}
func (in *Event) DeepCopyInto(out *Event) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.ObjectReference.DeepCopyInto(&out.ObjectReference)
	in.Source.DeepCopyInto(&out.Source)
}

func (in *EventList) DeepCopyInto(out *EventList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Events != nil {
		in, out := &in.Events, &out.Events
		*out = make([]Event, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *EventList) DeepCopy() *EventList {
	if in == nil {
		return nil
	}
	out := new(EventList)
	in.DeepCopyInto(out)
	return out
}

func (in *EventList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *EventSource) DeepCopyInto(out *EventSource) {
	*out = *in
}

func (in *ObjectReference) DeepCopyInto(out *ObjectReference) {
	*out = *in
}
