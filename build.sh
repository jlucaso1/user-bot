#!/bin/bash
set -e

BOT_NAME="bot.exe"
PORT=8000

restart_loop() {
  while true; do
    echo "killing any existing $BOT_NAME..."
    pkill -f "$BOT_NAME" || true

    echo "killing anything using port $PORT..."
    lsof -ti tcp:$PORT | xargs -r kill -9

    echo "building $BOT_NAME..."
    go build -o $BOT_NAME
    echo "build complete. starting $BOT_NAME..."
    ./$BOT_NAME

    echo "$BOT_NAME exited, restarting in 2 seconds..."
    sleep 2
  done
}

restart_loop
