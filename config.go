package podrick

// ContainerConfig is used by runtimes to start
// containers.
type ContainerConfig struct {
	Repo string
	Tag  string
	Port string

	// Optional
	Env        []string
	Entrypoint *string
	Cmd        []string
	Ulimits    []Ulimit
}

// Ulimit describes a container ulimit.
type Ulimit struct {
	Name string
	Soft int64
	Hard int64
}
