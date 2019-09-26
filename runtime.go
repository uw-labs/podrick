package podrick

import (
	"errors"
	"fmt"
	"io"
)

// Runtime supports starting containers.
type Runtime interface {
	io.Closer
	Connect() error
	StartContainer(*ContainerConfig) (Container, error)
}

// Container represents a running container.
type Container interface {
	io.Closer
	// Address contains the IP and port of the running container.
	Address() string
	// StreamLogs asynchronously streams logs from the
	// running container to the writer. The writer must
	// be safe for concurrent use.
	// This function is called automatically on the runtimes
	// configured logger, so there is no need to explicitly call this.
	StreamLogs(io.Writer) error
}

var autoRuntimes []Runtime

// RegisterAutoRuntime allows a runtime to register itself
// for auto-selection of a runtime, when one isn't explicitly specified.
func RegisterAutoRuntime(r Runtime) {
	autoRuntimes = append(autoRuntimes, r)
}

type autoRuntime struct {
	Runtime
}

// Connect establishes a connection with the underlying runtime.
func (r *autoRuntime) Connect() error {
	if len(autoRuntimes) == 0 {
		return errors.New("no container runtimes registered, import one or choose explicitly")
	}

	var errs []error
	for _, r.Runtime = range autoRuntimes {
		err := r.Runtime.Connect()
		if err == nil {
			return nil
		}
		errs = append(errs, err)
	}

	errStr := "failed to automatically choose runtime:\n"
	for _, err := range errs {
		errStr += fmt.Sprintf("\t%q", err.Error())
	}

	return errors.New(errStr)
}
