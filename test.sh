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

function run_sql() {
  engine="$1"
  sql_command="$2"
  output_name="$3"

  if [[ "${engine}" == "postgres" ]]; then
    docker exec -ti sqlow-${engine} psql -U sqlow sqlow -c "${sql_command}" > ${output_name}
  elif [[ "${engine}" == "maria" ]]; then
    docker exec -ti sqlow-maria mysql -usqlow -psqlow sqlow -e "${sql_command}" > ${output_name}
  fi
}

function compare_to_expected() {
  engine="$1"
  which_test="$2"
  diff --strip-trailing-cr "output.${which_test}.txt" "expected.${which_test}.txt" > /dev/null 2>&1
  if [[ $? -ne 0 ]]; then
    echo "❌ ${engine} - ${which_test}"
  else
    echo "✅ ${engine} - ${which_test}"
  fi
}


function matches_pattern() {
  engine="$1"
  which_test="$2"
  pattern="$3"
  cat "${DIR}/tests/${engine}/output.${which_test}.txt" | egrep "${pattern}" > /dev/null 2>&1
  if [[ $? -ne 0 ]]; then
    echo "❌ ${engine} - ${which_test}"
  else
    echo "✅ ${engine} - ${which_test}"
  fi
}


function run_for_engine() {
  engine="$1"
  cd ${DIR}/tests/${engine}/ || exit 1
  docker-compose up -d > /dev/null 2>&1
  sleep 5
  ${DIR}/dist/${GOOS}-${GOARCH}/${EXECUTABLE_NAME} -r -psqlow run ./migrations
  run_sql "${engine}" "select * from task_types;" "output.inline.txt"
  run_sql "${engine}" "select * from activities;" "output.files.txt"
  run_sql "${engine}" "select * from latest_upgrade;" "output.always.txt"
  RESULTS="${RESULTS}\n$(compare_to_expected ${engine} 'inline')"
  RESULTS="${RESULTS}\n$(compare_to_expected ${engine} 'files')"
  RESULTS="${RESULTS}\n$(matches_pattern ${engine} 'always' "[0-9]{4}-(10|11|12|0[1-9])-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}")"
  rm output*.txt
  docker-compose down
}


./build.sh --${OPTION}

if [[ "${ENGINE}" == "all" ]]; then
  run_for_engine "maria"
  run_for_engine "postgres"
else
  run_for_engine "${ENGINE}"
fi

echo -e "${RESULTS}"
