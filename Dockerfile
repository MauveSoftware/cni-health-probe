FROM golang as builder
ADD . /go/cni-health-probe/
WORKDIR /go/cni-health-probe
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo

FROM alpine:latest
RUN apk --no-cache add ca-certificates bash
ENV CONFIG_FILE /config/config.yml
ENV CMD_ARGS ""
COPY --from=builder /go/cni-health-probe/cni-health-probe /app/cni-health-probe
EXPOSE 9999
ENTRYPOINT /app/cni-health-probe --config $CONFIG_FILE $CMD_ARGS
