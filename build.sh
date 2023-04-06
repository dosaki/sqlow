#!/bin/bash

SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ]; do # resolve $SOURCE until the file is no longer a symlink
  DIR="$( cd -P "$( dirname "$SOURCE" )" >/dev/null && pwd )"
  SOURCE="$(readlink "$SOURCE")"
  [[ $SOURCE != /* ]] && SOURCE="$DIR/$SOURCE" # if $SOURCE was a relative symlink, we need to resolve it relative to the path where the symlink file was located
done
DIR="$( cd -P "$( dirname "$SOURCE" )" >/dev/null && pwd )"
cd "${DIR}"

EXECUTABLE_NAME="sqlow"
DO_ALL=false

for i in "$@"; do
  case $i in
  --all)
    DO_ALL=true
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
  *)
    usage
    echo "Unknown option ${i}"
    exit 1
    ;;
  esac
done

if [[ "${DO_ALL}" == "true" ]]; then
  rm -rf ./dist
  ./build.sh --windows
  ./build.sh --linux
  ./build.sh --macos
  ./build.sh --macos-m1
else
  mkdir -p dist/${GOOS}-${GOARCH}
  go build -o dist/${GOOS}-${GOARCH}/${EXECUTABLE_NAME} main.go
  pushd dist/${GOOS}-${GOARCH}/ > /dev/null 2>&1
  cp ${DIR}/config.template.yml config.yml
  zip -r "../${GOOS}-${GOARCH}.zip" ./* > /dev/null
  popd > /dev/null 2>&1
fi
