package v1

// 存储所有支持的Resource及Schema
var (
	Resources map[string]Item = make(map[string]Item)
)

type SchemaType string

const (
	HTTP SchemaType = "http"
	FILE SchemaType = "file"
)

type Item struct {
	//
	Name string `json:"name"`
	//
	Type SchemaType `json:"type"`
	//
	Schema string `json:"schema"`
}

func init() {

	// 注册不同的Schema
	Resources["Action"] = Item{Type: FILE, Schema: "./api/resource/v1/action.json"}
	Resources["ActionSpec"] = Item{Type: FILE, Schema: "./api/resource/v1/action_spec.json"}
	Resources["ActionTemplate"] = Item{Type: FILE, Schema: "./api/resource/v1/action_template.json"}
	Resources["ConditionFormula"] = Item{Type: FILE, Schema: "./api/resource/v1/condition_formula.json"}
	Resources["ConditionValue"] = Item{Type: FILE, Schema: "./api/resource/v1/condition_value.json"}
	Resources["Conditions"] = Item{Type: FILE, Schema: "./api/resource/v1/conditions.json"}
	Resources["ContainerImage"] = Item{Type: FILE, Schema: "./api/resource/v1/container_image.json"}
	Resources["DataSpec"] = Item{Type: FILE, Schema: "./api/resource/v1/data_spec.json"}
	Resources["DataStatus"] = Item{Type: FILE, Schema: "./api/resource/v1/data_status.json"}
	Resources["Deployment"] = Item{Type: FILE, Schema: "./api/resource/v1/deployment.json"}
	Resources["Description"] = Item{Type: FILE, Schema: "./api/resource/v1/description.json"}
	Resources["DeviceSpec"] = Item{Type: FILE, Schema: "./api/resource/v1/device_spec.json"}
	Resources["DeviceStatus"] = Item{Type: FILE, Schema: "./api/resource/v1/device_status.json"}
	Resources["EnvVar"] = Item{Type: FILE, Schema: "./api/resource/v1/env_var.json"}
	Resources["Event"] = Item{Type: FILE, Schema: "./api/resource/v1/event.json"}
	Resources["Group"] = Item{Type: FILE, Schema: "./api/resource/v1/group.json"}
	Resources["GroupSpec"] = Item{Type: FILE, Schema: "./api/resource/v1/GroupSpec.json"}
	Resources["GroupStatus"] = Item{Type: FILE, Schema: "./api/resource/v1/GroupStatus.json"}
	Resources["GroupTemplate"] = Item{Type: FILE, Schema: "./api/resource/v1/group_template.json"}
	Resources["IDRef"] = Item{Type: FILE, Schema: "./api/resource/v1/id_ref.json"}
	Resources["Input"] = Item{Type: FILE, Schema: "./api/resource/v1/input.json"}
	Resources["Node"] = Item{Type: FILE, Schema: "./api/resource/v1/node.json"}
	Resources["NodeAddress"] = Item{Type: FILE, Schema: "./api/resource/v1/node_address.json"}
	Resources["NodeList"] = Item{Type: FILE, Schema: "./api/resource/v1/node_list.json"}
	Resources["NodeSpec"] = Item{Type: FILE, Schema: "./api/resource/v1/node_spec.json"}
	Resources["NodeStatus"] = Item{Type: FILE, Schema: "./api/resource/v1/node_status.json"}
	Resources["NodeSystemInfo"] = Item{Type: FILE, Schema: "./api/resource/v1/node_system_info.json"}
	Resources["Output"] = Item{Type: FILE, Schema: "./api/resource/v1/output.json"}
	Resources["Pod"] = Item{Type: FILE, Schema: "./api/resource/v1/pod.json"}
	Resources["ResourceSpec"] = Item{Type: FILE, Schema: "./api/resource/v1/resource_spec.json"}
	Resources["ResourceStatus"] = Item{Type: FILE, Schema: "./api/resource/v1/resource_status.json"}
	Resources["Result"] = Item{Type: FILE, Schema: "./api/resource/v1/result.json"}
	Resources["Runtime"] = Item{Type: FILE, Schema: "./api/resource/v1/runtime.json"}
	Resources["RuntimeStatus"] = Item{Type: FILE, Schema: "./api/resource/v1/runtime_status.json"}
	Resources["SceneSpec"] = Item{Type: FILE, Schema: "./api/resource/v1/scene_spec.json"}
	Resources["SceneStatus"] = Item{Type: FILE, Schema: "./api/resource/v1/scene_status.json"}
	Resources["Service"] = Item{Type: FILE, Schema: "./api/resource/v1/service.json"}
	Resources["Task"] = Item{Type: FILE, Schema: "./api/resource/v1/task.json"}
	Resources["TaskSpec"] = Item{Type: FILE, Schema: "./api/resource/v1/task_spec.json"}
	Resources["TaskStatus"] = Item{Type: FILE, Schema: "./api/resource/v1/task_status.json"}
	Resources["TaskTemplate"] = Item{Type: FILE, Schema: "./api/resource/v1/task_template.json"}
	Resources["Time"] = Item{Type: FILE, Schema: "./api/resource/v1/time.json"}
	Resources["VM"] = Item{Type: FILE, Schema: "./api/resource/v1/vm.json"}
	Resources["WasmImage"] = Item{Type: FILE, Schema: "./api/resource/v1/wasm_image.json"}
	Resources["Workflow"] = Item{Type: FILE, Schema: "./api/resource/v1/workflow.json"}
	Resources["WorkflowSpec"] = Item{Type: FILE, Schema: "./api/resource/v1/workflow_spec.json"}
	Resources["WorkflowStatus"] = Item{Type: FILE, Schema: "./api/resource/v1/workflow_status.json"}
}

func Has(name string) (item Item, ok bool) {
	item, ok = Resources[name]
	return item, ok
}
