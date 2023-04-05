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
ENGINE="all"

for i in "$@"; do
  case $i in
  --engine)
    ENGINE="${i#*=}"
    shift
    ;;
  --windows)
    OPTION="windows"
    GOOS="windows"
    GOARCH="amd64"
    EXECUTABLE_NAME="sqlow.exe"
    shift
    ;;
  --linux)
    OPTION="linux"
    GOOS="linux"
    GOARCH="amd64"
    shift
    ;;
  --macos)
    OPTION="macos"
    GOOS="darwin"
    GOARCH="amd64"
    shift
    ;;
  --macos-m1)
    OPTION="macos-m1"
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

RESULTS="Results:"

function run_for_engine() {
  engine="$1"
  cd ${DIR}/tests/${engine}/ || exit 1
  docker-compose up -d > /dev/null 2>&1
  sleep 5
  ${DIR}/dist/${GOOS}-${GOARCH}/${EXECUTABLE_NAME} -r -psqlow run ./migrations
  if [[ "${engine}" == "postgres" ]]; then
    docker exec -ti sqlow-${engine} psql -U sqlow sqlow -c "select * from task_types;" > output.inline.txt
    docker exec -ti sqlow-${engine} psql -U sqlow sqlow -c "select * from activities;" > output.files.txt
  elif [[ "${engine}" == "maria" ]]; then
    docker exec -ti sqlow-maria mysql -usqlow -psqlow sqlow -e "select * from task_types;" > output.inline.txt
    docker exec -ti sqlow-maria mysql -usqlow -psqlow sqlow -e "select * from activities;" > output.files.txt
  fi

  diff --strip-trailing-cr output.inline.txt expected.inline.txt > /dev/null 2>&1
  if [[ $? -ne 0 ]]; then
    RESULTS="${RESULTS}\n ❌ ${engine} - Inline SQL"
  else
    RESULTS="${RESULTS}\n ✅ ${engine} - Inline SQL"
  fi
  diff --strip-trailing-cr output.files.txt expected.files.txt > /dev/null 2>&1
  if [[ $? -ne 0 ]]; then
    RESULTS="${RESULTS}\n ❌ ${engine} - Files SQL"
  else
    RESULTS="${RESULTS}\n ✅ ${engine} - Files SQL"
  fi
#  rm output*.txt
  docker-compose down
}

./build.sh --${OPTION}

if [[ "${ENGINE}" == "all" ]]; then
  run_for_engine "maria"
  run_for_engine "postgres"
  #wait $(jobs -p)
else
  run_for_engine "${ENGINE}"
fi

echo -e "${RESULTS}"
