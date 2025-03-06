#!/bin/bash

# This script demonstrates features that will be translated using gexe

# Variables
NAME="World"
COUNT=5

# Simple command
echo "Hello, $NAME!"

# Command with pipe
ls -la | grep "go"

# Command with subshell
echo "Current directory: $(pwd)"

# Conditional with file test
if [ -f "go.mod" ]; then
    echo "go.mod exists"
else
    echo "go.mod does not exist"
fi

# Loop with command substitution
for file in $(ls *.sh); do
    echo "Found script: $file"
done

# Multiple commands in sequence
echo "Running multiple commands:"
date
whoami
hostname

exit 0 