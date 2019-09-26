# Podman Varlink API

Assumes you have podman installed.

## Enabling the varlink API

This command blocks, so you can add it as a background startup process
if you want to be always be available, or you can start temporarily.

```bash
$ podman varlink unix:/home/$(whoami)/podman.socket -t 0
```

## Testing the API

```bash
$ PODMAN_VARLINK_ADDRESS=unix:/home/$(whoami)/podman.socket podman-remote info
client:
  Connection: unix:/home/johan/podman.socket
  Connection Type: DirectConnection
  OS Arch: linux/amd64
  Podman Version: 1.5.1
  RemoteAPI Version: 1
host:
  arch: amd64
  buildah_version: 1.10.1
  cpus: 8
  distribution:
    distribution: manjaro
    version: unknown
  eventlogger: file
  hostname: johan-x1
  kernel: 4.19.69-1-MANJARO
  mem_free: 907395072
  mem_total: 16569856000
  os: linux
  swap_free: 17130545152
  swap_total: 18223570944
  uptime: 213h 33m 17.91s (Approximately 8.88 days)
insecure registries:
  registries: null
registries:
  registries: null
store:
  containers: 1
  graph_driver_name: vfs
  graph_driver_options: ""
  graph_root: /home/johan/.local/share/containers/storage
  graph_status:
    backing_filesystem: ""
    native_overlay_diff: ""
    supports_d_type: ""
  images: 118
  run_root: /run/user/1000
```

## Generating the Go API

```bash
$ make all
```