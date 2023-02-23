package bitrise

// envStore ...
type envStore struct {
	Envs []map[string]string `json:"envs" yaml:"envs"`
}
