#!/bin/bash

# Function to display the loader
show_loader() {
  local pid=$1
  local delay=0.1
  local spinstr='|/-\'

  echo -n "Loading "

  # While the background process (sleep) is running
  while [ "$(ps a | awk '{print $1}' | grep $pid)" ]; do
      local temp=${spinstr#?}
      printf " [%c]  " "$spinstr"
      local spinstr=$temp${spinstr%"$temp"}
      sleep $delay
      printf "\b\b\b\b\b\b" # Backspace to clear the spinner
  done

  printf " [Done]\n"
}

# 1. Run the "work" in the background (3 seconds)
sleep 3 &

# 2. Get the PID of the background process
WORK_PID=$!

# 3. Call the loader passing the PID
show_loader $WORK_PID
