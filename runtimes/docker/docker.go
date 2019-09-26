package docker

import (
	"fmt"
	"io"
	"net"
	"runtime"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"logur.dev/logur"

	"github.com/uw-labs/podrick"
)

func init() {
	podrick.RegisterAutoRuntime(&Runtime{})
}

// Runtime implements the Runtime interface with
// a Docker runtime backend.
//
// The Docker API address can be configured using the environment
// variables DOCKER_HOST or DOCKER_URL, or from docker-machine if the
// environment variable DOCKER_MACHINE_NAME is set, or if neither is
// defined a sensible default for the operating system you are on.
// TLS pools are automatically configured if the DOCKER_CERT_PATH
// environment variable exists.
type Runtime struct {
	Logger podrick.Logger

	pool *dockertest.Pool
}

// Connect connects to the Docker API.
func (r *Runtime) Connect() error {
	if r.Logger == nil {
		r.Logger = logur.NewNoopLogger()
	}
	var err error
	r.pool, err = dockertest.NewPool("")
	if err != nil {
		return fmt.Errorf("failed to connect to docker: %w", err)
	}
	err = r.pool.Client.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping docker: %w", err)
	}

	return nil
}

// Close releases the resources of the Runtime.
func (Runtime) Close() error {
	return nil
}

// StartContainer starts a container with Docker as the backing runtime.
func (r *Runtime) StartContainer(conf *podrick.ContainerConfig) (podrick.Container, error) {
	ro, ho := createConfig(conf)
	ctr := &container{
		runtime: r,
	}
	var err error
	ctr.resource, err = r.pool.RunWithOptions(ro, ho...)
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	ctr.close = func() error {
		return r.pool.Purge(ctr.resource)
	}

	ctr.address = net.JoinHostPort(ctr.resource.Container.NetworkSettings.IPAddress, conf.Port)

	// Workaround for docker networking differences on Mac
	if runtime.GOOS == "darwin" {
		ctr.address = net.JoinHostPort(ctr.resource.GetBoundIP(conf.Port+"/tcp"), ctr.resource.GetPort(conf.Port+"/tcp"))
	}

	return ctr, nil
}

type container struct {
	address string
	close   func() error

	resource *dockertest.Resource
	runtime  *Runtime
}

func (c *container) Address() string {
	return c.address
}

func (c *container) Close() error {
	return c.close()
}

func (c *container) StreamLogs(w io.Writer) error {
	logWaiter, err := c.runtime.pool.Client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
		Container:    c.resource.Container.ID,
		OutputStream: logur.NewWriter(c.runtime.Logger),
		ErrorStream:  logur.NewWriter(c.runtime.Logger),
		Stderr:       true,
		Stdout:       true,
		Stream:       true,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to container log output: %w", err)
	}

	cls := c.close
	c.close = func() error {
		cErr := logWaiter.Close()
		if cErr != nil {
			c.runtime.Logger.Error("failed to close container log", map[string]interface{}{
				"error": cErr.Error(),
			})
		}
		cErr = logWaiter.Wait()
		if cErr != nil {
			c.runtime.Logger.Error("failed to wait for container log to close", map[string]interface{}{
				"error": cErr.Error(),
			})
		}
		return cls()
	}

	return nil
}
