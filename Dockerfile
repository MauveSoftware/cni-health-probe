FROM --platform=$BUILDPLATFORM golang:1.27.1-alpine3.24@sha256:cf6fca6641884b8433441b2b0652976f975e1d0fdd26d177eaaf8596087f3125 AS builder
ADD . /go/cni-health-probe/
WORKDIR /go/cni-health-probe
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo

FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6
RUN apk --no-cache add ca-certificates bash libcap
ENV CONFIG_FILE=/config/config.yml
ENV CMD_ARGS=""
COPY --from=builder /go/cni-health-probe/cni-health-probe /app/cni-health-probe
RUN setcap cap_net_raw,cap_net_admin+eip /app/cni-health-probe
EXPOSE 9999
ENTRYPOINT /app/cni-health-probe --config $CONFIG_FILE $CMD_ARGS
