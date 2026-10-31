ARG GO_VERSION=1.27

FROM --platform=$BUILDPLATFORM docker.io/library/golang:$GO_VERSION AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

COPY go.mod go.sum Makefile ./
RUN make download

COPY . .
RUN make build \
  BUILD_GOOS="$TARGETOS" \
  BUILD_GOARCH="$TARGETARCH"

FROM scratch

COPY --from=build /src/build/smsq /smsq

USER 65532:65532
ENTRYPOINT ["/smsq"]
CMD ["version"]
