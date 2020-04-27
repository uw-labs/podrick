update:
	rm io.podman.varlink
	wget https://raw.githubusercontent.com/containers/libpod/v1.5.1/cmd/podman/varlink/io.podman.varlink

generate:
	go install github.com/varlink/go/cmd/varlink-go-interface-generator
	varlink-go-interface-generator io.podman.varlink

all: update generate
