package podman

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"logur.dev/logur"
	"github.com/varlink/go/varlink"

	"github.com/uw-labs/podrick"
	podman "github.com/uw-labs/podrick/runtimes/podman/iopodman"
)

func init() {
	podrick.RegisterAutoRuntime(&Runtime{})
}

// Runtime implements the Runtime interface with
// a Podman runtime backend.
//
// The Podman API address can be configured using the environment variable
// PODMAN_VARLINK_ADDRESS. It defaults to "unix:/run/podman/io.podman".
type Runtime struct {
	Logger podrick.Logger

	address string
	conn    *varlink.Connection
	close   func() error
}

// Connect connects to the podman varlink API.
func (r *Runtime) Connect() (err error) {
	if r.Logger == nil {
		r.Logger = logur.NewNoopLogger()
	}
	r.address = os.Getenv("PODMAN_VARLINK_ADDRESS")
	if r.address == "" {
		// Default to root unix socket
		r.address = "unix:/run/podman/io.podman"
	}

	r.conn, err = varlink.NewConnection(r.address)
	if err != nil {
		return fmt.Errorf("failed to connect to podman: %w", err)
	}
	r.close = r.conn.Close
	defer func() {
		if err != nil {
			cErr := r.Close()
			if cErr != nil {
				r.Logger.Error("failed to close runtime during error", map[string]interface{}{
					"error": cErr.Error(),
				})
			}
		}
	}()

	_, err = podman.GetInfo().Call(r.conn)
	if err != nil {
		return fmt.Errorf("failed to ping podman: %w", err)
	}

	return nil
}

// Close releases the resources of the Runtime.
func (r *Runtime) Close() error {
	return r.close()
}

// StartContainer starts a container with Podman as the backing runtime.
func (r *Runtime) StartContainer(conf *podrick.ContainerConfig) (_ podrick.Container, err error) {
	ctr := &container{
		runtime: r,
	}
	ctr.id, err = podman.CreateContainer().Call(r.conn, createConfig(conf))
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	ctr.close = func() error {
		_, rErr := podman.RemoveContainer().Call(r.conn, ctr.id, false, true)
		if rErr != nil {
			return fmt.Errorf("failed to remove container: %w", rErr)
		}
		return nil
	}
	defer func() {
		if err != nil {
			cErr := ctr.Close()
			if cErr != nil {
				r.Logger.Error("failed to close container during error", map[string]interface{}{
					"error": cErr.Error(),
				})
			}
		}
	}()

	_, err = podman.StartContainer().Call(r.conn, ctr.id)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}
	cls2 := ctr.close
	ctr.close = func() error {
		_, kErr := podman.StopContainer().Call(r.conn, ctr.id, 5)
		if kErr != nil {
			return fmt.Errorf("failed to stop container: %w", kErr)
		}
		return cls2()
	}

	ct, err := podman.GetContainer().Call(r.conn, ctr.id)
	if err != nil {
		return nil, fmt.Errorf("failed to get container information: %w", err)
	}

	for _, p := range ct.Ports {
		if p.Container_port == conf.Port {
			ctr.address = net.JoinHostPort(p.Host_ip, p.Host_port)
			break
		}
	}
	if ctr.address == "" {
		return nil, fmt.Errorf("failed to get container IP")
	}

	return ctr, nil
}

type container struct {
	address string
	id      string
	close   func() error

	runtime *Runtime
}

func (c *container) Address() string {
	return c.address
}

func (c *container) Close() error {
	return c.close()
}

func (c *container) StreamLogs(w io.Writer) (err error) {
	logC, err := varlink.NewConnection(c.runtime.address)
	if err != nil {
		return fmt.Errorf("failed to get log connection: %w", err)
	}
	cls2 := c.close
	c.close = func() error {
		cErr := logC.Close()
		if cErr != nil {
			c.runtime.Logger.Error("failed to close logger connection", map[string]interface{}{
				"error": cErr.Error(),
			})
		}
		return cls2()
	}
	defer func() {
		if err != nil {
			cErr := c.Close()
			if cErr != nil {
				c.runtime.Logger.Error("failed to close container during error", map[string]interface{}{
					"error": cErr.Error(),
				})
			}
		}
	}()

	logFn, err := podman.GetContainerLogs().Send(logC, varlink.More, c.id)
	if err != nil {
		return fmt.Errorf("Failed get container logs: %w", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	cls3 := c.close
	c.close = func() error {
		cancel()
		// Ensure goroutine has exited
		<-done
		return cls3()
	}
	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			lines, f, err := logFn()
			if err != nil {
				c.runtime.Logger.Error("failed to get container logs", map[string]interface{}{
					"error": err.Error(),
				})
				return
			}
			for _, l := range lines {
				_, err = w.Write([]byte(l))
				if err != nil {
					c.runtime.Logger.Error("failed to write container logs", map[string]interface{}{
						"error": err.Error(),
					})
					return
				}
			}
			if f&varlink.Continues == 0 {
				return
			}
		}
	}()

	return nil
}
