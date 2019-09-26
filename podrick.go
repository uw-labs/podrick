package podrick

import (
	"fmt"

	"logur.dev/logur"
)

// StartContainer starts a container using the configured runtime.
// By default, a runtime is chosen automatically from those registered.
func StartContainer(repo, tag, port string, opts ...func(*config)) (_ Container, err error) {
	conf := config{
		ContainerConfig: ContainerConfig{
			Repo: repo,
			Tag:  tag,
			Port: port,
		},
		logger:  logur.NewNoopLogger(),
		runtime: &autoRuntime{},
	}
	for _, o := range opts {
		o(&conf)
	}

	err = conf.runtime.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to runtime: %w", err)
	}
	defer func() {
		if err != nil {
			cErr := conf.runtime.Close()
			if cErr != nil {
				conf.logger.Error("failed to close runtime", map[string]interface{}{
					"error": cErr.Error(),
				})
			}
		}
	}()

	ctr, err := conf.runtime.StartContainer(&conf.ContainerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}
	defer func() {
		if err != nil {
			cErr := ctr.Close()
			if cErr != nil {
				conf.logger.Error("failed to close container", map[string]interface{}{
					"error": cErr.Error(),
				})
			}
		}
	}()

	err = ctr.StreamLogs(logur.NewWriter(conf.logger))
	if err != nil {
		return nil, fmt.Errorf("failed to stream container logs: %w", err)
	}

	return ctr, nil
}
