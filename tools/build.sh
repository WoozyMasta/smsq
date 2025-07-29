#!/usr/bin/env bash
# require upx
set -eu

: "${WORK_DIR:=cmd}"
: "${BIN_NAME:=smsq}"

build() {
  local GOOS="${1:-linux}" GOARCH="${2:-amd64}" bin

  bin=$BIN_NAME-$GOOS-$GOARCH
  [ "$GOOS" = windows ] && bin+=.exe

  printf 'Build:\t%-10s%-7s' "$GOOS" "$GOARCH"

  CGO_ENABLED=0 GOARCH="$GOARCH" GOOS="$GOOS" \
  GOFLAGS="-buildvcs=false -trimpath" \
    go build -ldflags="-s -w -X '$version' -X '$commit' -X '$date' -X '$url'" \
      -o "./build/$bin" -tags=forceposix "$WORK_DIR"/*.go

  if command -v cyclonedx-gomod &>/dev/null; then
    cyclonedx-gomod bin -json -output "./build/$bin.sbom.json" "./build/$bin"
  fi

  if [ "$GOOS" == "linux" ] && command -v upx &>/dev/null; then
    upx --lzma --best "./build/$bin" > /dev/null
    upx -t "./build/$bin" > /dev/null
    printf '\tUPX packed\t'
  fi

  echo "./build/$bin"
}

module="$(grep -Po 'module \K.*$' go.mod)"
version="$module/internal/vars.Version=$(git describe --tags --abbrev=0 2>/dev/null || echo 0.0.0)"
commit="$module/internal/vars.Commit=$(git rev-parse HEAD 2>/dev/null || :)"
date="$module/internal/vars.BuildTime=$(date -uIs)"
url="$module/internal/vars.URL=https://$module"

mkdir -p ./build
go mod tidy

if command -v golangci-lint &>/dev/null; then
  golangci-lint run || :
fi

if [ -z "${1:-}" ]; then
  build darwin amd64
  build darwin arm64
  build linux 386
  build linux amd64
  build linux arm
  build linux arm64
  build windows 386
  build windows amd64
  build windows arm64

  if command -v cyclonedx-gomod &>/dev/null; then
    cyclonedx-gomod app -json -packages -files -licenses \
      -output "./build/$BIN_NAME.sbom.json" -main "$WORK_DIR"
  fi
else
  build "${@}"
fi
