package apis

// TODO:自动生成
var map_NodeList = map[string]string{
	"":         "list of all Nodes",
	"listmeta": "list metadata. ",
	"items":    "List of nodes",
}

func (NodeList) SwaggerDoc() map[string]string {
	return map_NodeList
}

var map_Node = map[string]string{
	"":         "Node is a worker node ",
	"metadata": "Standard object's metadata. ",
	"spec":     "Spec defines the behavior of a node. ",
	"status":   "Most recently observed status of the node. Populated by the system. ",
}

func (Node) SwaggerDoc() map[string]string {
	return map_Node
}
