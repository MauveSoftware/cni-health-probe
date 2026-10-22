# TODO

## Open
- [ ] (medium) Drop `cap_net_admin` from `setcap` in Dockerfile: ICMP raw sockets only need `cap_net_raw`. As nonroot, exec fails with EPERM unless the pod adds every file capability (NET_ADMIN is not in the default container capability set).

## Done
- [x] Switch runtime image to distroless `static-debian13:nonroot`; read flags from env (`CONFIG_FILE`, `METRICS_ADDRESS`) since there is no shell to expand `$CMD_ARGS`

## Decisions
- Only CLI flags get env support; the YAML config file stays required (config map mounts work fine with distroless).
