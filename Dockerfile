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


FROM gcr.io/distroless/static-debian13:nonroot@sha256:aafe993665bf6200db38a02c11a9351e782a84b6a45e7f33b6695638c88e3931
ENV CONFIG_FILE=/config/config.yml
COPY --from=builder /out/cni-health-probe /app/cni-health-probe
EXPOSE 9999
ENTRYPOINT ["/app/cni-health-probe"]
