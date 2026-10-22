# TODO

## Open
(none)

## Done
- [x] Switch runtime image to distroless `static-debian13:nonroot`; read flags from env (`CONFIG_FILE`, `METRICS_ADDRESS`) since there is no shell to expand `$CMD_ARGS`
- [x] Add version constant, `--version` flag and log version on start
- [x] Drop `cap_net_admin` from `setcap`: ICMP raw sockets only need `cap_net_raw`, and as nonroot every file capability must be granted to the container or exec fails with EPERM

## Decisions
- Only CLI flags get env support; the YAML config file stays required (config map mounts work fine with distroless).
