#!/bin/bash

set -e

echo ""
echo "--- BENCH PING PONG START ---"
echo ""

cd $(dirname "${BASH_SOURCE[0]}")
function cleanup() {
  echo "--- BENCH PING PONG DONE ---"
  #kill -9 $(jobs -rp)
  #wait $(jobs -rp) 2>/dev/null
}
trap cleanup EXIT

mkdir -p bin
$(pkill -9 net-echo-server || printf "")
$(pkill -9 evio-echo-server || printf "")
$(pkill -9 eviop-echo-server || printf "")
$(pkill -9 gev-echo-server || printf "")
$(pkill -9 hio-echo-server || printf "")

function gobench() {
  echo "--- $1 ---"
  if [ "$3" != "" ]; then
    go build -o $2 $3
  fi
  $2 --port $4 --loops 4 &

  sleep 1
  go run client/main.go -c 3000 -t 10 -m 16000 -a 127.0.0.1:$4

  pkill -9 $2 || printf ""
  echo "--- DONE ---"
  echo ""
}

gobench "hio" bin/hio-echo-server hio-echo-server/echo.go 16379

pkill echo