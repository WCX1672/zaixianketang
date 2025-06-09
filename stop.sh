#!/bin/bash

if [ -f .pids ]; then
  echo "Stopping all services..."
  while read pid; do
    kill $pid
  done < .pids
  rm .pids
  echo "All services stopped."
else
  echo "No running services found."
fi