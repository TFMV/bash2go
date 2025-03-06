#!/bin/bash
# comprehensive_test.sh
#
# This script tests all major features of bash2go transpiler
# It includes variables, functions, control flow, pipes, subshells, etc.

# ===== SECTION 1: Variables and Basic Commands =====
echo "===== SECTION 1: Variables and Basic Commands ====="

# Variable assignment and usage
NAME="World"
COUNT=5
EMPTY=""
QUOTED="Value with spaces"
ARRAY=("one" "two" "three")

# Echo with variable substitution
echo "Hello, $NAME!"
echo "Count is $COUNT"
echo "Empty variable: '$EMPTY'"
echo "Quoted value: $QUOTED"
echo "First array element: ${ARRAY[0]}"

# Command substitution
CURRENT_DIR=$(pwd)
echo "Current directory: $CURRENT_DIR"

DATE_OUTPUT=$(date +"%Y-%m-%d")
echo "Today's date: $DATE_OUTPUT"

# Arithmetic operations
RESULT=$((COUNT * 2))
echo "COUNT * 2 = $RESULT"

# Export variables
export EXPORTED_VAR="This is exported"
echo "Exported: $EXPORTED_VAR"

# ===== SECTION 2: Control Flow =====
echo "===== SECTION 2: Control Flow ====="

# If statement with file test
if [ -f "go.mod" ]; then
    echo "go.mod exists"
else
    echo "go.mod does not exist"
fi

# If statement with string comparison
if [ "$NAME" = "World" ]; then
    echo "Name is World"
elif [ "$NAME" = "Universe" ]; then
    echo "Name is Universe"
else
    echo "Name is something else"
fi

# If statement with numeric comparison
if [ $COUNT -gt 3 ]; then
    echo "Count is greater than 3"
fi

# For loop with range
echo "For loop with range:"
for i in {1..5}; do
    echo "  Item $i"
done

# For loop with list
echo "For loop with list:"
for item in apple banana cherry; do
    echo "  Fruit: $item"
done

# For loop with command substitution
echo "For loop with command substitution:"
for file in $(ls *.sh); do
    echo "  Script: $file"
done

# While loop
echo "While loop:"
counter=1
while [ $counter -le 3 ]; do
    echo "  Counter: $counter"
    counter=$((counter + 1))
done

# Until loop
echo "Until loop:"
counter=5
until [ $counter -le 0 ]; do
    echo "  Countdown: $counter"
    counter=$((counter - 1))
done

# Case statement
echo "Case statement:"
FRUIT="apple"
case $FRUIT in
    "apple")
        echo "  It's an apple"
        ;;
    "banana")
        echo "  It's a banana"
        ;;
    *)
        echo "  It's something else"
        ;;
esac

# ===== SECTION 3: Functions =====
echo "===== SECTION 3: Functions ====="

# Simple function
simple_function() {
    echo "This is a simple function"
}

# Function with parameters
greet() {
    local name=$1
    echo "Hello, $name!"
}

# Function with return value
get_sum() {
    local a=$1
    local b=$2
    echo $((a + b))
}

# Function with local variables
process_data() {
    local input=$1
    local result="Processed: $input"
    echo "$result"
}

# Call functions
simple_function
greet "John"
SUM=$(get_sum 5 7)
echo "Sum: $SUM"
PROCESSED=$(process_data "raw data")
echo "$PROCESSED"

# ===== SECTION 4: Pipes and Redirections =====
echo "===== SECTION 4: Pipes and Redirections ====="

# Simple pipe
echo "Simple pipe:"
ls -la | grep ".sh"

# Multiple pipes
echo "Multiple pipes:"
ls -la | grep ".sh" | wc -l

# Pipe with command substitution
echo "Files count: $(ls -la | grep ".sh" | wc -l)"

# Output redirection
echo "Output to file" > temp_output.txt
echo "Appending to file" >> temp_output.txt

# Input redirection
echo "Reading from file:"
cat < temp_output.txt

# Here document
cat << EOF
This is a here document.
It can span multiple lines.
Variables like $NAME are expanded.
EOF

# ===== SECTION 5: Subshells and Background Tasks =====
echo "===== SECTION 5: Subshells and Background Tasks ====="

# Subshell
echo "Subshell:"
(
    cd /tmp
    echo "  Current directory in subshell: $(pwd)"
)
echo "Current directory after subshell: $(pwd)"

# Command grouping
echo "Command grouping:"
{
    echo "  These commands"
    echo "  are grouped"
}

# Background task
echo "Background task:"
(sleep 1; echo "Background task completed") &
echo "Main script continues..."

# Wait for background tasks
wait
echo "All background tasks completed"

# ===== SECTION 6: Error Handling =====
echo "===== SECTION 6: Error Handling ====="

# Check exit status
ls /nonexistent 2>/dev/null
if [ $? -ne 0 ]; then
    echo "Command failed"
fi

# Set error handling
set -e  # Exit on error
echo "Error handling enabled"
# Uncomment to test: ls /nonexistent

# Trap signals
trap "echo 'Caught signal'; exit 1" SIGINT SIGTERM
echo "Signal trap set"

# Cleanup
rm -f temp_output.txt

echo "Test completed successfully"
exit 0 