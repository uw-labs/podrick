package podman

import (
	"strconv"

	"github.com/uw-labs/podrick"
	podman "github.com/uw-labs/podrick/runtimes/podman/iopodman"
)

func createConfig(conf *podrick.ContainerConfig) podman.Create {
	crt := podman.Create{
		Args: append(
			[]string{
				conf.Repo + ":" + conf.Tag,
			},
			conf.Cmd...,
		),
		Publish:    &[]string{conf.Port},
		Entrypoint: conf.Entrypoint,
	}
	if len(conf.Ulimits) > 0 {
		var ulimits []string
		for _, ulimit := range conf.Ulimits {
			ulimits = append(ulimits, ulimitToPodman(ulimit))
		}
		crt.Ulimit = &ulimits
	}
	if len(conf.Env) > 0 {
		crt.Env = &conf.Env
	}
	return crt
}

func ulimitToPodman(u podrick.Ulimit) string {
	return u.Name + "=" + strconv.Itoa(int(u.Soft)) + ":" + strconv.Itoa(int(u.Hard))
}
