#!/bin/bash

SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do # resolve $SOURCE until the file is no longer a symlink
  DIR="$( cd -P "$( dirname "$SOURCE" )" >/dev/null && pwd )"
  SOURCE="$(readlink "$SOURCE")"
  [[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE" # if $SOURCE was a relative symlink, we need to resolve it relative to the path where the symlink file was located
done
DIR="$( cd -P "$( dirname "$SOURCE" )" >/dev/null && pwd )"
cd "${DIR}"

go env > .go.env
source .go.env

EXECUTABLE_NAME="sqlow"
if [[ "${GOOS}" == "windows" ]]; then
  EXECUTABLE_NAME="sqlow.exe" # For when it's --default
fi

DO_ALL=false
VERSION=$(cat ./config/version.go | tail -1 | tr '"' ' ' | awk '{print $4}')

for i in "$@"; do
  case $i in
  --all)
    DO_ALL=true
    shift
    ;;
  --default)
    shift
    ;;
  --go-arch=*)
    GOARCH="${i#*=}"
    shift
    ;;
  --windows)
    GOOS="windows"
    GOARCH="amd64"
    EXECUTABLE_NAME="sqlow.exe"
    shift
    ;;
  --linux)
    GOOS="linux"
    GOARCH="amd64"
    shift
    ;;
  --macos)
    GOOS="darwin"
    GOARCH="amd64"
    shift
    ;;
  --macos-m1)
    GOOS="darwin"
    GOARCH="arm64"
    shift
    ;;
  --version=*)
    VERSION="${i#*=}"
    shift
    ;;
  *)
    usage
    echo "Unknown option ${i}"
    exit 1
    ;;
  esac
done

if [[ "${DO_ALL}" == "true" ]]; then
  rm -rf ./dist
  ./build.sh --windows --version=${VERSION}
  ./build.sh --linux --version=${VERSION}
  ./build.sh --macos --version=${VERSION}
  ./build.sh --macos-m1 --version=${VERSION}
else
  mkdir -p dist/${GOOS}-${GOARCH}
  mv ./config/version.go ./config/version.tmp
  cat ./config/version.tmp | sed "s/development/${VERSION} ${GOOS}\/${GOARCH}/g" > ./config/version.go
  go build -o dist/${GOOS}-${GOARCH}/${EXECUTABLE_NAME} main.go
  mv ./config/version.tmp ./config/version.go
  pushd dist/${GOOS}-${GOARCH}/ > /dev/null 2>&1
  cp ${DIR}/config.template.yml config.yml
  zip -r "../${GOOS}-${GOARCH}.zip" ./* > /dev/null
  popd > /dev/null 2>&1
fi
