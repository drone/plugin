package bitrise

type EnvironmentItemModel map[string]string

// EnvsSerializeModel ...
type EnvsSerializeModel struct {
	Envs []EnvironmentItemModel `json:"envs" yaml:"envs"`
}
