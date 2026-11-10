FROM --platform=$BUILDPLATFORM golang:1.27.4-alpine3.24@sha256:0a5f7fe796599ff487605987432e7619ae45beeb2fba4c2ce0b907879932aebf AS builder

ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT

ADD . /go/cni-health-probe/
WORKDIR /go/cni-health-probe

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
  GOARM=${TARGETVARIANT#v} \
  go build -trimpath \
  -ldflags="-s -w" \
  -o /out/cni-health-probe .

RUN apk --no-cache add libcap && \
  setcap cap_net_raw+eip /out/cni-health-probe


FROM gcr.io/distroless/static-debian13:nonroot@sha256:25ac8f0c3b2334a9eb05452266bd4697864a8c8ef0ef71e5e6480d7f92868329
ENV CONFIG_FILE=/config/config.yml
COPY --from=builder /out/cni-health-probe /app/cni-health-probe
EXPOSE 9999
ENTRYPOINT ["/app/cni-health-probe"]
