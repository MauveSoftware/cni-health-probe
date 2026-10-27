FROM --platform=$BUILDPLATFORM golang:1.27.3-alpine3.24@sha256:d530cf60723abf1ffae0cd16bb546b379ec0550d2f3815e5b16ed69dbc00fba4 AS builder

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


FROM gcr.io/distroless/static-debian13:nonroot@sha256:18f4cc26c9ffa7f20a359298be7deadac5c607c8e03a02b8119f1a5fc2f82c99
ENV CONFIG_FILE=/config/config.yml
COPY --from=builder /out/cni-health-probe /app/cni-health-probe
EXPOSE 9999
ENTRYPOINT ["/app/cni-health-probe"]
