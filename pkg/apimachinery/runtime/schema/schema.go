package schema

// TODO: Schema格式定义
// TODO: 序列化和反序列化

type Schema struct {
	// Schema name
	schemaName string
}

func NewSchema() *Schema {
	s := &Schema{}
	return s
}
