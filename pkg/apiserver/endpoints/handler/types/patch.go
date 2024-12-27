package types

type PatchType string

const (
	JSONPatchType  PatchType = "application/json-patch+json"
	MergePatchType PatchType = "application/merge-patch+json"
)
