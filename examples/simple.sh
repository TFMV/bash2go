#!/bin/bash

# Simple Bash script with various features
# This script demonstrates different Bash constructs that bash2go should be able to handle

# Variables
NAME="World"
COUNT=5

# Function definition
say_hello() {
    local greeting="Hello, $1!"
    echo "$greeting"
}

# Function call
say_hello "$NAME"

# Loop
for i in $(seq 1 $COUNT); do
    echo "Counter: $i"
done

# Conditional
if [ -f "/etc/hosts" ]; then
    echo "Hosts file exists"
else
    echo "Hosts file does not exist"
fi

# Pipe
echo "Testing pipes" | grep "Testing"

# Subshell
(
    cd /tmp
    echo "Current directory: $(pwd)"
)

# Back to original directory (not affected by subshell)
echo "Back in original directory: $(pwd)"

exit 0 