#!/bin/bash

process_name=clickety-clack

# Find the parent process ID
parent_pid=$(pgrep -o "$process_name")

if [ -z "$parent_pid" ]; then
  echo "Process '$process_name' not found."
  exit 1
fi

# Get memory usage for the parent process and all its children
total_memory=$(ps -o pid= -p "$parent_pid" $(pgrep -P "$parent_pid") | xargs ps -o rss= -p | awk '{sum+=$1} END {print sum/1024}')

echo "Total memory used by '$process_name' and its sub-processes: ${total_memory%.2f} MB"
