package docker

import (
	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/uw-labs/podrick"
)

func createConfig(conf *podrick.ContainerConfig) (*dockertest.RunOptions, []func(*docker.HostConfig)) {
	runOpts := &dockertest.RunOptions{
		Repository:   conf.Repo,
		Tag:          conf.Tag,
		ExposedPorts: []string{conf.Port},
	}
	if len(conf.Cmd) > 0 {
		runOpts.Cmd = conf.Cmd
	}
	if len(conf.Env) > 0 {
		runOpts.Env = conf.Env
	}
	if conf.Entrypoint != nil {
		runOpts.Entrypoint = []string{*conf.Entrypoint}
	}
	var hostOpts []func(*docker.HostConfig)
	if len(conf.Ulimits) > 0 {
		hostOpts = append(hostOpts, func(in *docker.HostConfig) {
			for _, ulimit := range conf.Ulimits {
				in.Ulimits = append(in.Ulimits, docker.ULimit(ulimit))
			}
		})
	}
	return runOpts, hostOpts
}
